// Workflow written in SciPipe.  // For more information about SciPipe, see: http://scipipe.org
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
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
		header := []string{"Gene", "Efficiency", "Accuracy", "ObsFuzzActive", "ObsFuzzNonactive", "ObsFuzzOverall", "ClassConfidence", "ClassCredibility", "Cost"}
		if p.IncludeGamma {
			header = append(header, "Gamma")
		}
		rows := [][]string{header}
		for iip := range p.In().Chan {
			gene := iip.Param("gene")
			efficiency := iip.Key("efficiency")
			accuracy := iip.Key("accuracy")
			cost := iip.Param("cost")

			obsFuzzActive := iip.Key("obsfuzz_active")
			obsFuzzNonactive := iip.Key("obsfuzz_nonactive")
			obsFuzzOverall := iip.Key("obsfuzz_overall")
			classConfidence := iip.Key("class_confidence")
			classCredibility := iip.Key("class_credibility")

			row := []string{gene, efficiency, accuracy, obsFuzzActive, obsFuzzNonactive, obsFuzzOverall, classConfidence, classCredibility, cost}
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
	AccuracyValColIdx   int // Which column to check for the accuracy value
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
	sbcr.InitParamOutPort(sbcr, "best_accuracy")
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
func (p *BestCostGamma) OutBestAccuracy() *sp.ParamOutPort {
	return p.ParamOutPort("best_accuracy")
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
	defer p.OutBestAccuracy().Close()
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

		bestObsFuzzOverall := 1000000.000 // N.B: The best efficiency in Conformal Prediction is the *minimal* one. Initializing here with an unreasonably large number in order to spot when something is wrong.

		var header []string

		var bestCost int64 = -1
		var bestGamma float64 = -1.0
		var bestAccuracy float64 = -1.0
		var bestEfficiency float64 = -1.0
		var bestObsFuzzActive float64 = -1.0
		var bestObsFuzzNonactive float64 = -1.0
		var bestClassConfidence float64 = -1.0
		var bestClassCredibility float64 = -1.0
		var bestClassAvgObsFuzz float64 = -1.0

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

			obsFuzzOverall, err := strconv.ParseFloat(rec[indexOfStr("ObsFuzzOverall", header)], 64)
			sp.Check(err)

			if obsFuzzOverall < bestObsFuzzOverall { // Smaller is better
				bestObsFuzzOverall = obsFuzzOverall

				sp.Debug.Printf("Proc:%s Raw cost value: %s\n", p.Name(), rec[indexOfStr("Cost", header)])
				bestCost, err = strconv.ParseInt(rec[indexOfStr("Cost", header)], 10, 0)
				sp.Debug.Printf("Proc:%s Parsed cost value: %d\n", p.Name(), bestCost)
				sp.CheckWithMsg(err, "Could not parse best cost value")

				if p.IncludeGamma {
					bestGamma, err = strconv.ParseFloat(rec[indexOfStr("Gamma", header)], 64)
					sp.Check(err)
				}

				bestAccuracy, err = strconv.ParseFloat(rec[indexOfStr("Accuracy", header)], 64)
				sp.Check(err)

				bestEfficiency, err = strconv.ParseFloat(rec[indexOfStr("Efficiency", header)], 64)
				sp.Check(err)

				bestObsFuzzActive, err := strconv.ParseFloat(rec[indexOfStr("ObsFuzzActive", header)], 64)
				sp.Check(err)

				bestObsFuzzNonactive, err := strconv.ParseFloat(rec[indexOfStr("ObsFuzzNonactive", header)], 64)
				sp.Check(err)

				bestClassAvgObsFuzz = (bestObsFuzzActive + bestObsFuzzNonactive) / 2 // We take the average for the two classes, to get more equal influence of each class

				bestClassConfidence, err = strconv.ParseFloat(rec[indexOfStr("ClassConfidence", header)], 64)
				sp.Check(err)

				bestClassCredibility, err = strconv.ParseFloat(rec[indexOfStr("ClassCredibility", header)], 64)
				sp.Check(err)
			}
		}
		sp.Debug.Printf("Final optimal (minimal) observed fuzziness (overall): %f (For: Cost:%03d)\n", bestObsFuzzOverall, bestCost)
		if p.IncludeGamma {
			sp.Debug.Printf("Final optimal (minimal) observed fuzziness (overall): %f (For: Cost:%03d, Gamma:%.3f)\n", bestObsFuzzOverall, bestCost, bestGamma)
		}
		p.OutBestCost().Send(fmt.Sprintf("%d", bestCost))
		if p.IncludeGamma {
			p.OutBestGamma().Send(fmt.Sprintf("%.3f", bestGamma))
		}
		p.OutBestAccuracy().Send(fmt.Sprintf("%.3f", bestAccuracy))
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
		runSet := tdip.Param("runset")
		uniq := gene + "_" + runSet

		strs := str.Split(string(tdip.Read()), "\t")
		activeStr := str.TrimSuffix(strs[0], "\n")
		activeCnt, err := strconv.ParseInt(activeStr, 10, 64)
		sp.CheckWithMsg(err, "Could not parse active count value")
		nonActiveStr := str.TrimSuffix(strs[1], "\n")
		nonActiveCnt, err := strconv.ParseInt(nonActiveStr, 10, 64)
		sp.CheckWithMsg(err, "Could not parse non-active count value")
		activeCounts[uniq] = activeCnt
		nonActiveCounts[uniq] = nonActiveCnt
		totalCompounds[uniq] = activeCnt + nonActiveCnt
	}

	rows := [][]string{[]string{
		"Gene",
		"Replicate",
		"Runset",
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
	for iip := range p.InModel().Chan {
		uniq := iip.Param("gene") + "_" + iip.Param("runset")
		row := []string{
			iip.Param("gene"),
			iip.Param("replicate"),
			iip.Param("runset"),
			iip.Param("accuracy"),
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
			fmt.Sprintf("%d", activeCounts[uniq]),
			fmt.Sprintf("%d", nonActiveCounts[uniq]),
			fmt.Sprintf("%d", totalCompounds[uniq]),
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

// ================================================================================

type EmbedAuditLogInJar struct {
	*sp.Process
}

func (p *EmbedAuditLogInJar) InJarFile() *sp.InPort   { return p.In("in_jar") }
func (p *EmbedAuditLogInJar) OutJarFile() *sp.OutPort { return p.Out("out_jar") }

func NewEmbedAuditLogInJar(wf *sp.Workflow, procName string) *EmbedAuditLogInJar {
	p := &EmbedAuditLogInJar{wf.NewProc(procName, "# EmbedAuditLogInJar custom process. Ports: {i:in_jar} {o:out_jar}")}
	p.SetPathExtend("in_jar", "out_jar", ".withaudit.jar")
	p.CustomExecute = func(t *sp.Task) {
		jarFilePath := t.InPath("in_jar")
		unpackDirPath := jarFilePath + ".unpack"
		auditFilePath := filepath.Base(t.InIP("in_jar").AuditFilePath())
		sp.ExecCmd(fmt.Sprintf(`origDir=$(pwd)/$(dirname %s); mkdir %s && cd %s && jar xvf ../%s && cp ../%s . && jar cf $origDir/%s *`,
			t.OutIP("out_jar").TempPath(),
			unpackDirPath,
			unpackDirPath,
			filepath.Base(t.InPath("in_jar")),
			auditFilePath,
			filepath.Base(t.OutIP("out_jar").TempPath())))
	}
	return p
}
