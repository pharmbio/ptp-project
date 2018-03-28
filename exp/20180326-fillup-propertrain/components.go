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

// gene, efficiency, validity, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, cost)

// SummarizeCostGammaPerf is specialized a SciPipe Process that reads output
// from cpSign status output to extract information about the efficiency and
// validity of generated models for given cost and gamma values
type SummarizeCostGammaPerf struct {
	sp.BaseProcess
	procName     string
	FileName     string
	IncludeGamma bool
}

func NewSummarizeCostGammaPerf(wf *sp.Workflow, name string, filename string, includeGamma bool) *SummarizeCostGammaPerf {
	p := &SummarizeCostGammaPerf{
		BaseProcess:  sp.NewBaseProcess(wf, name),
		FileName:     filename,
		IncludeGamma: includeGamma,
	}
	p.InitInPort(p, "in")
	p.InitOutPort(p, "out_stats")
	wf.AddProc(p)
	return p
}

func (p *SummarizeCostGammaPerf) In() *sp.InPort        { return p.InPort("in") }
func (p *SummarizeCostGammaPerf) OutStats() *sp.OutPort { return p.OutPort("out_stats") }

func (p *SummarizeCostGammaPerf) Run() {
	defer p.OutStats().Close()

	outIp := sp.NewFileIP(p.FileName)
	if outIp.Exists() {
		sp.Info.Printf("Process %s: Out-target %s already exists, so skipping\n", p.Name(), outIp.Path())
	} else {
		header := []string{"Gene", "Efficiency", "Validity", "ObsFuzzActive", "ObsFuzzNonactive", "ObsFuzzOverall", "ClassConfidence", "ClassCredibility", "Cost"}
		if p.IncludeGamma {
			header = append(header, "Gamma")
		}
		rows := [][]string{header}
		for iip := range p.In().Chan {
			gene := iip.Param("gene")
			efficiency := iip.Key("efficiency")
			validity := iip.Key("validity")
			cost := iip.Param("cost")

			obsFuzzActive := iip.Key("obsfuzz_active")
			obsFuzzNonactive := iip.Key("obsfuzz_nonactive")
			obsFuzzOverall := iip.Key("obsfuzz_overall")
			classConfidence := iip.Key("class_confidence")
			classCredibility := iip.Key("class_credibility")

			row := []string{gene, efficiency, validity, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, classConfidence, classCredibility, cost}
			if p.IncludeGamma {
				row = append(row, iip.Param("gamma"))
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
	p.OutStats().Send(outIp)
}

// ================================================================================

type BestCostGamma struct {
	sp.BaseProcess
	procName            string
	Separator           rune
	Header              bool
	EfficiencyValColIdx int // Which column to check for the efficiency value
	ValidityValColIdx   int // Which column to check for the validity value
	IncludeGamma        bool
}

func NewBestCostGamma(wf *sp.Workflow, procName string, separator rune, header bool, includeGamma bool) *BestCostGamma {
	sbcr := &BestCostGamma{
		BaseProcess:  sp.NewBaseProcess(wf, procName),
		Separator:    separator,
		Header:       header,
		IncludeGamma: includeGamma,
	}
	sbcr.InitInPort(sbcr, "csv_file")
	sbcr.InitParamOutPort(sbcr, "best_validity")
	sbcr.InitParamOutPort(sbcr, "best_eff")
	sbcr.InitParamOutPort(sbcr, "best_obsfuzz_classavg")
	sbcr.InitParamOutPort(sbcr, "best_obsfuzz_overall")
	sbcr.InitParamOutPort(sbcr, "best_obsfuzz_active")
	sbcr.InitParamOutPort(sbcr, "best_obsfuzz_nonactive")
	sbcr.InitParamOutPort(sbcr, "best_class_confidence")
	sbcr.InitParamOutPort(sbcr, "best_class_credibility")
	sbcr.InitParamOutPort(sbcr, "best_cost")
	sbcr.InitParamOutPort(sbcr, "best_gamma")
	wf.AddProc(sbcr)
	return sbcr
}

func (p BestCostGamma) InCSVFile() *sp.InPort {
	return p.InPort("csv_file")
}
func (p *BestCostGamma) OutBestValidity() *sp.ParamOutPort {
	return p.ParamOutPort("best_validity")
}
func (p *BestCostGamma) OutBestEfficiency() *sp.ParamOutPort {
	return p.ParamOutPort("best_eff")
}
func (p *BestCostGamma) OutBestObsFuzzClassAvg() *sp.ParamOutPort {
	return p.ParamOutPort("best_obsfuzz_classavg")
}
func (p *BestCostGamma) OutBestObsFuzzOverall() *sp.ParamOutPort {
	return p.ParamOutPort("best_obsfuzz_overall")
}
func (p *BestCostGamma) OutBestObsFuzzActive() *sp.ParamOutPort {
	return p.ParamOutPort("best_obsfuzz_active")
}
func (p *BestCostGamma) OutBestObsFuzzNonactive() *sp.ParamOutPort {
	return p.ParamOutPort("best_obsfuzz_nonactive")
}
func (p *BestCostGamma) OutBestClassConfidence() *sp.ParamOutPort {
	return p.ParamOutPort("best_class_confidence")
}
func (p *BestCostGamma) OutBestClassCredibility() *sp.ParamOutPort {
	return p.ParamOutPort("best_class_credibility")
}
func (p *BestCostGamma) OutBestCost() *sp.ParamOutPort {
	return p.ParamOutPort("best_cost")
}
func (p *BestCostGamma) OutBestGamma() *sp.ParamOutPort {
	return p.ParamOutPort("best_gamma")
}

func (p *BestCostGamma) Run() {
	defer p.OutBestCost().Close()
	if p.IncludeGamma {
		defer p.OutBestGamma().Close()
	}
	defer p.OutBestValidity().Close()
	defer p.OutBestEfficiency().Close()
	defer p.OutBestObsFuzzClassAvg().Close()
	defer p.OutBestObsFuzzOverall().Close()
	defer p.OutBestObsFuzzActive().Close()
	defer p.OutBestObsFuzzNonactive().Close()
	defer p.OutBestClassConfidence().Close()
	defer p.OutBestClassCredibility().Close()

	for iip := range p.InCSVFile().Chan {
		csvData := iip.Read()

		bytesReader := bytes.NewReader(csvData)
		csvReader := csv.NewReader(bytesReader)
		csvReader.Comma = p.Separator

		bestClassAvgObsFuzz := 1000000.000 // N.B: The best efficiency in Conformal Prediction is the *minimal* one. Initializing here with an unreasonably large number in order to spot when something is wrong.

		var header []string

		var bestCost int64 = -1
		var bestGamma float64 = -1.0
		var bestValidity float64 = -1.0
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
			sp.Check(err)

			obsFuzzNonactive, err := strconv.ParseFloat(rec[indexOfStr("ObsFuzzNonactive", header)], 64)
			sp.Check(err)

			classAvgObsFuzz := (obsFuzzActive + obsFuzzNonactive) / 2 // We take the average for the two classes, to get more equal influence of each class

			if classAvgObsFuzz < bestClassAvgObsFuzz { // Smaller is better
				bestClassAvgObsFuzz = classAvgObsFuzz

				sp.Debug.Printf("Proc:%s Raw cost value: %s\n", p.Name(), rec[indexOfStr("Cost", header)])
				bestCost, err = strconv.ParseInt(rec[indexOfStr("Cost", header)], 10, 0)
				sp.Debug.Printf("Proc:%s Parsed cost value: %d\n", p.Name(), bestCost)
				sp.Check(err)

				if p.IncludeGamma {
					bestGamma, err = strconv.ParseFloat(rec[indexOfStr("Gamma", header)], 64)
					sp.Check(err)
				}

				bestValidity, err = strconv.ParseFloat(rec[indexOfStr("Validity", header)], 64)
				sp.Check(err)

				bestEfficiency, err = strconv.ParseFloat(rec[indexOfStr("Efficiency", header)], 64)
				sp.Check(err)

				bestObsFuzzOverall, err = strconv.ParseFloat(rec[indexOfStr("ObsFuzzOverall", header)], 64)
				sp.Check(err)

				bestObsFuzzActive = obsFuzzActive
				bestObsFuzzNonactive = obsFuzzNonactive

				bestClassConfidence, err = strconv.ParseFloat(rec[indexOfStr("ClassConfidence", header)], 64)
				sp.Check(err)

				bestClassCredibility, err = strconv.ParseFloat(rec[indexOfStr("ClassCredibility", header)], 64)
				sp.Check(err)
			}
		}
		sp.Debug.Printf("Final optimal (minimal) class-equalized observed fuzziness: %f (For: Cost:%03d)\n", bestClassAvgObsFuzz, bestCost)
		if p.IncludeGamma {
			sp.Debug.Printf("Final optimal (minimal) class-equalized observed fuzziness: %f (For: Cost:%03d, Gamma:%.3f)\n", bestClassAvgObsFuzz, bestCost, bestGamma)
		}
		p.OutBestCost().Send(fmt.Sprintf("%d", bestCost))
		if p.IncludeGamma {
			p.OutBestGamma().Send(fmt.Sprintf("%.3f", bestGamma))
		}
		p.OutBestValidity().Send(fmt.Sprintf("%.3f", bestValidity))
		p.OutBestEfficiency().Send(fmt.Sprintf("%.3f", bestEfficiency))
		p.OutBestObsFuzzClassAvg().Send(fmt.Sprintf("%.3f", bestClassAvgObsFuzz))
		p.OutBestObsFuzzOverall().Send(fmt.Sprintf("%.3f", bestObsFuzzOverall))
		p.OutBestObsFuzzActive().Send(fmt.Sprintf("%.3f", bestObsFuzzActive))
		p.OutBestObsFuzzNonactive().Send(fmt.Sprintf("%.3f", bestObsFuzzNonactive))
		p.OutBestClassConfidence().Send(fmt.Sprintf("%.3f", bestClassConfidence))
		p.OutBestClassCredibility().Send(fmt.Sprintf("%.3f", bestClassCredibility))
	}
}

// ================================================================================

type ParamPrinter struct {
	sp.BaseProcess
	procName           string
	BestParamsFileName string
}

func NewParamPrinter(wf *sp.Workflow, procName string, fileName string) *ParamPrinter {
	p := &ParamPrinter{
		BaseProcess:        sp.NewBaseProcess(wf, procName),
		procName:           procName,
		BestParamsFileName: fileName,
	}
	p.InitOutPort(p, "best_param")
	wf.AddProc(p)
	return p
}

func (p *ParamPrinter) OutBestParamsFile() *sp.OutPort {
	return p.OutPort("best_param")
}

func (p *ParamPrinter) GetNewParamInPort(portName string) *sp.ParamInPort {
	if _, ok := p.ParamInPorts()[portName]; !ok {
		p.InitParamInPort(p, portName)
	}
	return p.ParamInPort(portName)
}

func (p *ParamPrinter) Run() {
	defer p.OutBestParamsFile().Close()

	oip := sp.NewFileIP(p.BestParamsFileName)
	if !oip.Exists() && !oip.TempFileExists() {
		rows := []map[string]string{}
		for len(p.ParamInPorts()) > 0 {
			row := map[string]string{}
			for pname, pport := range p.ParamInPorts() {
				param, ok := <-pport.Chan
				if !ok {
					p.DeleteParamInPort(pname) // This we should implement in the BaseProcess instead!
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
		oip.Write([]byte(outContent))
		oip.Atomize()
	} else {
		sp.Info.Printf("Target file (or temp file) exists for: %s, so skipping\n", oip.Path())
	}

	p.OutBestParamsFile().Send(oip)
}

// ================================================================================

type FinalModelSummarizer struct {
	sp.BaseProcess
	SummaryFileName string
	Separator       rune
}

func NewFinalModelSummarizer(wf *sp.Workflow, procName string, fileName string, separator rune) *FinalModelSummarizer {
	p := &FinalModelSummarizer{
		BaseProcess:     sp.NewBaseProcess(wf, procName),
		SummaryFileName: fileName,
		Separator:       separator,
	}
	p.InitInPort(p, "model")
	p.InitInPort(p, "target_data_count")
	p.InitOutPort(p, "summary")
	// InModel:           sp.NewInPort(),
	// InTargetDataCount: sp.NewInPort(),
	// OutSummary:        sp.NewOutPort(),
	wf.AddProc(p)
	return p
}

func (p *FinalModelSummarizer) InModel() *sp.InPort           { return p.InPort("model") }
func (p *FinalModelSummarizer) InTargetDataCount() *sp.InPort { return p.InPort("target_data_count") }
func (p *FinalModelSummarizer) OutSummary() *sp.OutPort       { return p.OutPort("summary") }

func (p *FinalModelSummarizer) Run() {
	defer p.OutSummary().Close()

	activeCounts := map[string]int64{}
	nonActiveCounts := map[string]int64{}
	totalCompounds := map[string]int64{}
	for tdip := range p.InTargetDataCount().Chan {
		gene := tdip.Param("gene")
		strs := str.Split(string(tdip.Read()), "\t")
		activeStr := str.TrimSuffix(strs[0], "\n")
		activeCnt, err := strconv.ParseInt(activeStr, 10, 64)
		nonActiveStr := str.TrimSuffix(strs[1], "\n")
		nonActiveCnt, err := strconv.ParseInt(nonActiveStr, 10, 64)
		sp.Check(err)
		activeCounts[gene] = activeCnt
		nonActiveCounts[gene] = nonActiveCnt
		totalCompounds[gene] = activeCnt + nonActiveCnt
	}

	rows := [][]string{[]string{
		"Gene",
		"Replicate",
		"Validity",
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
	for iip := range p.InModel().Chan {
		row := []string{
			iip.Param("gene"),
			iip.Param("replicate"),
			iip.Param("validity"),
			iip.Param("efficiency"),
			iip.Param("obsfuzz_classavg"),
			iip.Param("obsfuzz_overall"),
			iip.Param("obsfuzz_active"),
			iip.Param("obsfuzz_nonactive"),
			iip.Param("class_confidence"),
			iip.Param("class_credibility"),
			iip.Param("cost"),
			fmt.Sprintf("%d", iip.AuditInfo().ExecTimeMS),
			fmt.Sprintf("%d", iip.Size()),
			fmt.Sprintf("%d", activeCounts[iip.Param("gene")]),
			fmt.Sprintf("%d", nonActiveCounts[iip.Param("gene")]),
			fmt.Sprintf("%d", totalCompounds[iip.Param("gene")]),
		}
		rows = append(rows, row)
	}

	oip := sp.NewFileIP(p.SummaryFileName)
	fh := oip.OpenWriteTemp()
	csvWriter := csv.NewWriter(fh)
	csvWriter.Comma = p.Separator
	for _, row := range rows {
		csvWriter.Write(row)
	}
	csvWriter.Flush()
	fh.Close()
	oip.Atomize()
	p.OutSummary().Send(oip)
}
