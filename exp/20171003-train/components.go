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

var fid = map[string]int{
	"gene":             0,
	"efficiency":       1,
	"validity":         2,
	"obsFuzzActive":    3,
	"obsFuzzNonactive": 4,
	"obsFuzzOverall":   5,
	"cost":             6,
	"gamma":            7,
}

// gene, efficiency, validity, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, cost)

// SummarizeCostGammaPerf is specialized a SciPipe Process that reads output
// from cpSign status output to extract information about the efficiency and
// validity of generated models for given cost and gamma values
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
		outStr := "Gene\tEfficiency\tValidity\tObsFuzzActive\tObsFuzzNonactive\tObsFuzzOverall\tCost\n"
		if p.IncludeGamma {
			outStr = "Gene\tEfficiency\tValidity\tObsFuzzActive\tObsFuzzNonactive\tObsFuzzOverall\tCost\tGamma\n"
		}
		for iip := range p.In.InChan {
			gene := iip.GetParam("gene")
			efficiency := iip.GetKey("efficiency")
			validity := iip.GetKey("validity")
			cost := iip.GetParam("cost")

			obsFuzzActive := iip.GetKey("obsfuzz_active")
			obsFuzzNonactive := iip.GetKey("obsfuzz_nonactive")
			obsFuzzOverall := iip.GetKey("obsfuzz_overall")

			infoString := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", gene, efficiency, validity, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, cost)
			if p.IncludeGamma {
				gamma := iip.GetParam("gamma")
				infoString = fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", gene, efficiency, validity, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, cost, gamma)
			}
			outStr = outStr + infoString
		}
		outIp.WriteTempFile([]byte(outStr))
		outIp.Atomize()
	}
	p.OutStats.Send(outIp)
}

func (p *SummarizeCostGammaPerf) IsConnected() bool {
	return p.In.IsConnected() && p.OutStats.IsConnected()
}

// ================================================================================

type BestCostGamma struct {
	procName               string
	InCSVFile              *sp.FilePort
	OutBestCost            *sp.ParamPort
	OutBestGamma           *sp.ParamPort
	OutBestClassAvgObsFuzz *sp.ParamPort
	OutBestEfficiency      *sp.ParamPort
	OutBestValidity        *sp.ParamPort
	Separator              rune
	Header                 bool
	EfficiencyValColIdx    int // Which column to check for the efficiency value
	ValidityValColIdx      int // Which column to check for the validity value
	IncludeGamma           bool
}

func NewBestCostGamma(wf *sp.Workflow, procName string, separator rune, header bool, includeGamma bool) *BestCostGamma {
	sbcr := &BestCostGamma{
		procName:               procName,
		InCSVFile:              sp.NewFilePort(),
		OutBestCost:            sp.NewParamPort(),
		OutBestGamma:           sp.NewParamPort(),
		OutBestClassAvgObsFuzz: sp.NewParamPort(),
		OutBestEfficiency:      sp.NewParamPort(),
		OutBestValidity:        sp.NewParamPort(),
		Separator:              separator,
		Header:                 header,
		IncludeGamma:           includeGamma,
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
	defer p.OutBestClassAvgObsFuzz.Close()
	defer p.OutBestEfficiency.Close()
	defer p.OutBestValidity.Close()

	go p.InCSVFile.RunMergeInputs()

	for iip := range p.InCSVFile.InChan {
		csvData := iip.Read()

		bytesReader := bytes.NewReader(csvData)
		csvReader := csv.NewReader(bytesReader)
		csvReader.Comma = p.Separator

		minClassAvgObsFuzz := 1000000.000 // N.B: The best efficiency in Conformal Prediction is the *minimal* one. Initializing here with an unreasonably large number in order to spot when something is wrong.

		var bestCost int64
		var bestGamma float64 // Only used for libSVM
		var bestEfficiency float64
		var bestValidity float64

		i := 0
		for {
			rec, err := csvReader.Read()
			if err != nil {
				break
			}
			i++
			if i == 1 && !p.Header {
				continue
			}

			obsFuzzActive, err := strconv.ParseFloat(rec[fid["obsFuzzActive"]], 64)
			sp.CheckErr(err)

			obsFuzzNonactive, err := strconv.ParseFloat(rec[fid["obsFuzzNonactive"]], 64)
			sp.CheckErr(err)

			classAvgObsFuzz := (obsFuzzActive + obsFuzzNonactive) / 2 // We take the average for the two classes, to get more equal influence of each class

			if classAvgObsFuzz < minClassAvgObsFuzz {
				minClassAvgObsFuzz = classAvgObsFuzz

				sp.Debug.Printf("Proc:%s Raw cost value: %s\n", p.Name(), rec[fid["cost"]])
				bestCost, err = strconv.ParseInt(rec[fid["cost"]], 10, 0)
				sp.Debug.Printf("Proc:%s Parsed cost value: %d\n", p.Name(), bestCost)
				sp.CheckErr(err)

				if p.IncludeGamma {
					bestGamma, err = strconv.ParseFloat(rec[fid["gamma"]], 64)
					sp.CheckErr(err)
				}

				bestEfficiency, err = strconv.ParseFloat(rec[fid["efficiency"]], 64)
				sp.CheckErr(err)
				bestValidity, err = strconv.ParseFloat(rec[fid["validity"]], 64)
				sp.CheckErr(err)
			}
		}
		sp.Debug.Printf("Final optimal (minimal) class-equalized observed fuzziness: %f (For: Cost:%03d)\n", minClassAvgObsFuzz, bestCost)
		if p.IncludeGamma {
			sp.Debug.Printf("Final optimal (minimal) class-equalized observed fuzziness: %f (For: Cost:%03d, Gamma:%.3f)\n", minClassAvgObsFuzz, bestCost, bestGamma)
		}
		p.OutBestCost.Send(fmt.Sprintf("%d", bestCost))
		if p.IncludeGamma {
			p.OutBestGamma.Send(fmt.Sprintf("%.3f", bestGamma))
		}
		p.OutBestClassAvgObsFuzz.Send(fmt.Sprintf("%.3f", minClassAvgObsFuzz))
		p.OutBestEfficiency.Send(fmt.Sprintf("%.3f", bestEfficiency))
		p.OutBestValidity.Send(fmt.Sprintf("%.3f", bestValidity))
	}
}

func (p *BestCostGamma) IsConnected() bool {
	if p.IncludeGamma {
		return p.InCSVFile.IsConnected() && p.OutBestCost.IsConnected() && p.OutBestGamma.IsConnected() && p.OutBestClassAvgObsFuzz.IsConnected() && p.OutBestValidity.IsConnected()
	}
	return p.InCSVFile.IsConnected() && p.OutBestCost.IsConnected() && p.OutBestClassAvgObsFuzz.IsConnected() && p.OutBestValidity.IsConnected()
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

	rows := [][]string{[]string{"Gene", "Replicate", "ClassAvgObsFuzz", "Efficiency", "Validity", "Cost", "ExecTimeMS", "ModelFileSize", "Active", "Nonactive", "TotalCompounds"}}
	for iip := range p.InModel.InChan {
		row := []string{
			iip.GetParam("gene"),
			iip.GetParam("replicate"),
			iip.GetParam("clsavgobsfuzz"),
			iip.GetParam("efficiency"),
			iip.GetParam("validity"),
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
	oip.Atomize()
	p.OutSummary.Send(oip)
}
