// Workflow written in SciPipe.  // For more information about SciPipe, see: http://scipipe.org
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	str "strings"

	sp "github.com/scipipe/scipipe"
)

// ================================================================================

func indexOfStr(s string, strs []string) int {
	for i, _ := range strs {
		if strs[i] == s {
			return i
		}
	}
	sp.Error.Fatalf("Did not find index of string %s, in strins: %v\n", s, strs)
	return -1
}

// gene, efficiency, accuracy, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, cost)

// SummarizeCostGammaPerf is specialized a SciPipe Process that reads output
// from cpSign status output to extract information about the efficiency and
// accuracy of generated models for given cost and gamma values
type SummarizeCostGammaPerf struct {
	In           *sp.FilePort
	OutStats     *sp.FilePort
	procName     string
	FileName     string
	IncludeGamma bool
}

func NewSummarizeCostGammaPerf(wf *sp.Workflow, name string, filename string, includeGamma bool) *SummarizeCostGammaPerf {
	bcgs := &SummarizeCostGammaPerf{
		In:           sp.NewFilePort(),
		OutStats:     sp.NewFilePort(),
		procName:     name,
		FileName:     filename,
		IncludeGamma: includeGamma,
	}
	wf.AddProc(bcgs)
	return bcgs
}

func (p *SummarizeCostGammaPerf) Name() string {
	return p.procName
}

func (p *SummarizeCostGammaPerf) Run() {
	defer p.OutStats.Close()
	go p.In.RunMergeInputs()

	outIp := sp.NewInformationPacket(p.FileName)
	if outIp.Exists() {
		sp.Info.Printf("Process %s: Out-target %s already exists, so skipping\n", p.Name(), outIp.GetPath())
	} else {
		header := []string{"Gene", "Efficiency", "Accuracy", "ObsFuzzActive", "ObsFuzzNonactive", "ObsFuzzOverall", "ClassConfidence", "ClassCredibility", "Cost"}
		if p.IncludeGamma {
			header = append(header, "Gamma")
		}
		rows := [][]string{header}
		for iip := range p.In.InChan {
			gene := iip.GetParam("gene")
			efficiency := iip.GetKey("efficiency")
			accuracy := iip.GetKey("accuracy")
			cost := iip.GetParam("cost")

			obsFuzzActive := iip.GetKey("obsfuzz_active")
			obsFuzzNonactive := iip.GetKey("obsfuzz_nonactive")
			obsFuzzOverall := iip.GetKey("obsfuzz_overall")
			classConfidence := iip.GetKey("class_confidence")
			classCredibility := iip.GetKey("class_credibility")

			row := []string{gene, efficiency, accuracy, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, classConfidence, classCredibility, cost}
			if p.IncludeGamma {
				row = append(row, iip.GetParam("gamma"))
			}
			rows = append(rows, row)
		}
		ofh := outIp.OpenWriteTemp()
		tsvWriter := csv.NewWriter(ofh)
		tsvWriter.Comma = '\t'
		for _, row := range rows {
			tsvWriter.Write(row)
		}
		tsvWriter.Flush()
		ofh.Close()
		outIp.Atomize()
	}
	p.OutStats.Send(outIp)
}

func (p *SummarizeCostGammaPerf) IsConnected() bool {
	return p.In.IsConnected() && p.OutStats.IsConnected()
}

// ================================================================================

type BestCostGamma struct {
	procName                string
	InCSVFile               *sp.FilePort
	OutBestCost             *sp.ParamPort
	OutBestGamma            *sp.ParamPort
	OutBestAccuracy         *sp.ParamPort
	OutBestEfficiency       *sp.ParamPort
	OutBestObsFuzzClassAvg  *sp.ParamPort
	OutBestObsFuzzOverall   *sp.ParamPort
	OutBestObsFuzzActive    *sp.ParamPort
	OutBestObsFuzzNonactive *sp.ParamPort
	OutBestClassConfidence  *sp.ParamPort
	OutBestClassCredibility *sp.ParamPort
	Separator               rune
	Header                  bool
	EfficiencyValColIdx     int // Which column to check for the efficiency value
	AccuracyValColIdx       int // Which column to check for the accuracy value
	IncludeGamma            bool
}

func NewBestCostGamma(wf *sp.Workflow, procName string, separator rune, header bool, includeGamma bool) *BestCostGamma {
	sbcr := &BestCostGamma{
		procName:                procName,
		InCSVFile:               sp.NewFilePort(),
		OutBestAccuracy:         sp.NewParamPort(),
		OutBestEfficiency:       sp.NewParamPort(),
		OutBestObsFuzzClassAvg:  sp.NewParamPort(),
		OutBestObsFuzzOverall:   sp.NewParamPort(),
		OutBestObsFuzzActive:    sp.NewParamPort(),
		OutBestObsFuzzNonactive: sp.NewParamPort(),
		OutBestClassConfidence:  sp.NewParamPort(),
		OutBestClassCredibility: sp.NewParamPort(),
		OutBestCost:             sp.NewParamPort(),
		OutBestGamma:            sp.NewParamPort(),
		Separator:               separator,
		Header:                  header,
		IncludeGamma:            includeGamma,
	}
	wf.AddProc(sbcr)
	return sbcr
}

func (p *BestCostGamma) Name() string {
	return p.procName
}

func (p *BestCostGamma) Run() {
	defer p.OutBestCost.Close()
	if p.IncludeGamma {
		defer p.OutBestGamma.Close()
	}
	defer p.OutBestAccuracy.Close()
	defer p.OutBestEfficiency.Close()
	defer p.OutBestObsFuzzClassAvg.Close()
	defer p.OutBestObsFuzzOverall.Close()
	defer p.OutBestObsFuzzActive.Close()
	defer p.OutBestObsFuzzNonactive.Close()
	defer p.OutBestClassConfidence.Close()
	defer p.OutBestClassCredibility.Close()

	go p.InCSVFile.RunMergeInputs()

	for iip := range p.InCSVFile.InChan {
		csvData := iip.Read()

		bytesReader := bytes.NewReader(csvData)
		csvReader := csv.NewReader(bytesReader)
		csvReader.Comma = p.Separator

		bestClassAvgObsFuzz := 1000000.000 // N.B: The best efficiency in Conformal Prediction is the *minimal* one. Initializing here with an unreasonably large number in order to spot when something is wrong.

		var header []string

		var bestCost int64 = -1
		var bestGamma float64 = -1.0
		var bestAccuracy float64 = -1.0
		var bestEfficiency float64 = -1.0
		var bestObsFuzzOverall float64 = -1.0
		var bestObsFuzzActive float64 = -1.0
		var bestObsFuzzNonactive float64 = -1.0
		var bestClassConfidence float64 = -1.0
		var bestClassCredibility float64 = -1.0

		i := 0
		for {
			rec, err := csvReader.Read()
			if err != nil {
				break
			}
			i++
			if i == 1 {
				header = rec
				if !p.Header {
					continue
				}
			}

			obsFuzzActive, err := strconv.ParseFloat(rec[indexOfStr("ObsFuzzActive", header)], 64)
			sp.CheckErr(err)

			obsFuzzNonactive, err := strconv.ParseFloat(rec[indexOfStr("ObsFuzzNonactive", header)], 64)
			sp.CheckErr(err)

			classAvgObsFuzz := (obsFuzzActive + obsFuzzNonactive) / 2 // We take the average for the two classes, to get more equal influence of each class

			if classAvgObsFuzz < bestClassAvgObsFuzz { // Smaller is better
				bestClassAvgObsFuzz = classAvgObsFuzz

				sp.Debug.Printf("Proc:%s Raw cost value: %s\n", p.Name(), rec[indexOfStr("Cost", header)])
				bestCost, err = strconv.ParseInt(rec[indexOfStr("Cost", header)], 10, 0)
				sp.Debug.Printf("Proc:%s Parsed cost value: %d\n", p.Name(), bestCost)
				sp.CheckErr(err)

				if p.IncludeGamma {
					bestGamma, err = strconv.ParseFloat(rec[indexOfStr("Gamma", header)], 64)
					sp.CheckErr(err)
				}

				bestAccuracy, err = strconv.ParseFloat(rec[indexOfStr("Accuracy", header)], 64)
				sp.CheckErr(err)

				bestEfficiency, err = strconv.ParseFloat(rec[indexOfStr("Efficiency", header)], 64)
				sp.CheckErr(err)

				bestObsFuzzOverall, err = strconv.ParseFloat(rec[indexOfStr("ObsFuzzOverall", header)], 64)
				sp.CheckErr(err)

				bestObsFuzzActive = obsFuzzActive
				bestObsFuzzNonactive = obsFuzzNonactive

				bestClassConfidence, err = strconv.ParseFloat(rec[indexOfStr("ClassConfidence", header)], 64)
				sp.CheckErr(err)

				bestClassCredibility, err = strconv.ParseFloat(rec[indexOfStr("ClassCredibility", header)], 64)
				sp.CheckErr(err)
			}
		}
		sp.Debug.Printf("Final optimal (minimal) class-equalized observed fuzziness: %f (For: Cost:%03d)\n", bestClassAvgObsFuzz, bestCost)
		if p.IncludeGamma {
			sp.Debug.Printf("Final optimal (minimal) class-equalized observed fuzziness: %f (For: Cost:%03d, Gamma:%.3f)\n", bestClassAvgObsFuzz, bestCost, bestGamma)
		}
		p.OutBestCost.Send(fmt.Sprintf("%d", bestCost))
		if p.IncludeGamma {
			p.OutBestGamma.Send(fmt.Sprintf("%.3f", bestGamma))
		}
		p.OutBestAccuracy.Send(fmt.Sprintf("%.3f", bestAccuracy))
		p.OutBestEfficiency.Send(fmt.Sprintf("%.3f", bestEfficiency))
		p.OutBestObsFuzzClassAvg.Send(fmt.Sprintf("%.3f", bestClassAvgObsFuzz))
		p.OutBestObsFuzzOverall.Send(fmt.Sprintf("%.3f", bestObsFuzzOverall))
		p.OutBestObsFuzzActive.Send(fmt.Sprintf("%.3f", bestObsFuzzActive))
		p.OutBestObsFuzzNonactive.Send(fmt.Sprintf("%.3f", bestObsFuzzNonactive))
		p.OutBestClassConfidence.Send(fmt.Sprintf("%.3f", bestClassConfidence))
		p.OutBestClassCredibility.Send(fmt.Sprintf("%.3f", bestClassCredibility))
	}
}

func (p *BestCostGamma) IsConnected() bool {
	if p.IncludeGamma {
		return p.InCSVFile.IsConnected() &&
			p.OutBestAccuracy.IsConnected() &&
			p.OutBestEfficiency.IsConnected() &&
			p.OutBestObsFuzzClassAvg.IsConnected() &&
			p.OutBestObsFuzzOverall.IsConnected() &&
			p.OutBestObsFuzzActive.IsConnected() &&
			p.OutBestObsFuzzNonactive.IsConnected() &&
			p.OutBestClassConfidence.IsConnected() &&
			p.OutBestClassCredibility.IsConnected() &&
			p.OutBestCost.IsConnected() &&
			p.OutBestGamma.IsConnected()
	}
	return p.InCSVFile.IsConnected() &&
		p.OutBestAccuracy.IsConnected() &&
		p.OutBestEfficiency.IsConnected() &&
		p.OutBestObsFuzzClassAvg.IsConnected() &&
		p.OutBestObsFuzzOverall.IsConnected() &&
		p.OutBestObsFuzzActive.IsConnected() &&
		p.OutBestObsFuzzNonactive.IsConnected() &&
		p.OutBestClassConfidence.IsConnected() &&
		p.OutBestClassCredibility.IsConnected() &&
		p.OutBestCost.IsConnected()
}

// ================================================================================

type ParamPrinter struct {
	sp.SciProcess
	procName           string
	InParamPorts       map[string]*sp.ParamPort
	OutBestParamsFile  *sp.FilePort
	BestParamsFileName string
}

func NewParamPrinter(wf *sp.Workflow, procName string, fileName string) *ParamPrinter {
	pp := &ParamPrinter{
		procName:           procName,
		InParamPorts:       make(map[string]*sp.ParamPort),
		OutBestParamsFile:  sp.NewFilePort(),
		BestParamsFileName: fileName,
	}
	wf.AddProc(pp)
	return pp
}

func (p *ParamPrinter) GetParamPort(portName string) *sp.ParamPort {
	if p.InParamPorts[portName] == nil {
		p.InParamPorts[portName] = sp.NewParamPort()
	}
	return p.InParamPorts[portName]
}

func (p *ParamPrinter) Name() string {
	return p.procName
}

func (p *ParamPrinter) Run() {
	defer p.OutBestParamsFile.Close()

	oip := sp.NewInformationPacket(p.BestParamsFileName)
	if !oip.Exists() && !oip.TempFileExists() {
		rows := []map[string]string{}
		for len(p.InParamPorts) > 0 {
			row := map[string]string{}
			for pname, pport := range p.InParamPorts {
				param, ok := <-pport.Chan
				if !ok {
					delete(p.InParamPorts, pname)
					continue
				}
				row[pname] = param
			}
			rows = append(rows, row)
		}

		var outContent string

		for _, row := range rows {
			for name, val := range row {
				outContent += fmt.Sprintf("%s=%s\n", name, val)
			}
		}
		oip.WriteTempFile([]byte(outContent))
		oip.Atomize()
	} else {
		sp.Info.Printf("Target file (or temp file) exists for: %s, so skipping\n", oip.GetPath())
	}

	p.OutBestParamsFile.Send(oip)
}

// ================================================================================

type FinalModelSummarizer struct {
	procName          string
	SummaryFileName   string
	Separator         rune
	InModel           *sp.FilePort
	InTargetDataCount *sp.FilePort
	OutSummary        *sp.FilePort
}

func NewFinalModelSummarizer(wf *sp.Workflow, name string, fileName string, separator rune) *FinalModelSummarizer {
	fms := &FinalModelSummarizer{
		procName:          name,
		SummaryFileName:   fileName,
		InModel:           sp.NewFilePort(),
		InTargetDataCount: sp.NewFilePort(),
		OutSummary:        sp.NewFilePort(),
		Separator:         separator,
	}
	wf.AddProc(fms)
	return fms
}

func (p *FinalModelSummarizer) Name() string {
	return p.procName
}

func (p *FinalModelSummarizer) IsConnected() bool {
	return p.InModel.IsConnected() && p.OutSummary.IsConnected()
}

func (p *FinalModelSummarizer) Run() {
	defer p.OutSummary.Close()
	go p.InModel.RunMergeInputs()
	go p.InTargetDataCount.RunMergeInputs()

	activeCounts := map[string]int64{}
	nonActiveCounts := map[string]int64{}
	totalCompounds := map[string]int64{}
	for tdip := range p.InTargetDataCount.InChan {
		gene := tdip.GetParam("gene")
		strs := str.Split(string(tdip.Read()), "\t")
		activeStr := str.TrimSuffix(strs[0], "\n")
		activeCnt, err := strconv.ParseInt(activeStr, 10, 64)
		nonActiveStr := str.TrimSuffix(strs[1], "\n")
		nonActiveCnt, err := strconv.ParseInt(nonActiveStr, 10, 64)
		sp.CheckErr(err)
		activeCounts[gene] = activeCnt
		nonActiveCounts[gene] = nonActiveCnt
		totalCompounds[gene] = activeCnt + nonActiveCnt
	}

	rows := [][]string{[]string{
		"Gene",
		"Replicate",
		"Accuracy",
		"Efficiency",
		"ObsFuzzClassAvg",
		"ObsFuzzOverall",
		"ObsFuzzActive",
		"ObsFuzzNonactive",
		"ClassConfidence",
		"ClassCredibility",
		"Cost",
		"ExecTimeMS",
		"SizeBytes",
		"ActiveCnt",
		"NonactiveCnt",
		"TotalCnt"}}
	for iip := range p.InModel.InChan {
		row := []string{
			iip.GetParam("gene"),
			iip.GetParam("replicate"),
			iip.GetParam("accuracy"),
			iip.GetParam("efficiency"),
			iip.GetParam("obsfuzz_classavg"),
			iip.GetParam("obsfuzz_overall"),
			iip.GetParam("obsfuzz_active"),
			iip.GetParam("obsfuzz_nonactive"),
			iip.GetParam("class_confidence"),
			iip.GetParam("class_credibility"),
			iip.GetParam("cost"),
			fmt.Sprintf("%d", iip.GetAuditInfo().ExecTimeMS),
			fmt.Sprintf("%d", iip.GetSize()),
			fmt.Sprintf("%d", activeCounts[iip.GetParam("gene")]),
			fmt.Sprintf("%d", nonActiveCounts[iip.GetParam("gene")]),
			fmt.Sprintf("%d", totalCompounds[iip.GetParam("gene")]),
		}
		rows = append(rows, row)
	}

	oip := sp.NewInformationPacket(p.SummaryFileName)
	fh := oip.OpenWriteTemp()
	csvWriter := csv.NewWriter(fh)
	csvWriter.Comma = p.Separator
	for _, row := range rows {
		csvWriter.Write(row)
	}
	csvWriter.Flush()
	fh.Close()
	oip.Atomize()
	p.OutSummary.Send(oip)
}
