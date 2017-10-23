// Workflow written in SciPipe.  // For more information about SciPipe, see: http://scipipe.org
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	str "strings"

	sp "github.com/scipipe/scipipe"
)

// ================================================================================

// SummarizeCostGammaPerf is specialized a SciPipe Process that reads output
// from cpSign status output to extract information about the efficiency and
// validity of generated models for given cost and gamma values
type SummarizeCostGammaPerf struct {
	In           *sp.FilePort
	OutStats     *sp.FilePort
	ProcName     string
	FileName     string
	IncludeGamma bool
}

func NewSummarizeCostGammaPerf(wf *sp.Workflow, name string, filename string, includeGamma bool) *SummarizeCostGammaPerf {
	bcgs := &SummarizeCostGammaPerf{
		In:           sp.NewFilePort(),
		OutStats:     sp.NewFilePort(),
		ProcName:     name,
		FileName:     filename,
		IncludeGamma: includeGamma,
	}
	wf.AddProc(bcgs)
	return bcgs
}

func (p *SummarizeCostGammaPerf) Name() string {
	return p.ProcName
}

func (p *SummarizeCostGammaPerf) Run() {
	defer p.OutStats.Close()
	go p.In.RunMergeInputs()

	outIp := sp.NewInformationPacket(p.FileName)

	if outIp.Exists() {
		sp.Info.Printf("Process %s: Out-target %s already exists, so not skipping\n", p.Name(), outIp.GetPath())
	} else {
		// Set up regexes
		rEffic, err := regexp.Compile("Efficiency=([0-9.]+)")
		sp.CheckErr(err)

		rValid, err := regexp.Compile("Validity=([0-9.]+)")
		sp.CheckErr(err)

		outStr := "Gene\tEfficiency\tValidity\tCost\n"
		if p.IncludeGamma {
			outStr = "Gene\tEfficiency\tValidity\tCost\tGamma\n"
		}
		for iip := range p.In.InChan {
			dat := string(iip.Read())

			efficiency := 0.0
			validity := 0.0

			effMatches := rEffic.FindStringSubmatch(dat)
			if len(effMatches) > 1 {
				efficiency, err = strconv.ParseFloat(effMatches[1], 64)
				sp.CheckErr(err)
			}

			validMatches := rValid.FindStringSubmatch(dat)
			if len(validMatches) > 1 {
				validity, err = strconv.ParseFloat(validMatches[1], 64)
				sp.CheckErr(err)
			}

			auditInfo := iip.GetAuditInfo()
			cost := auditInfo.Params["cost"]
			gene := auditInfo.Params["gene"]
			infoString := fmt.Sprintf("%s\t%.3f\t%.3f\t%s\n", gene, efficiency, validity, cost)
			if p.IncludeGamma {
				gamma := auditInfo.Params["gamma"]
				infoString = fmt.Sprintf("%s\t%.3f\t%.3f\t%s\t%s\n", gene, efficiency, validity, cost, gamma)
			}

			outStr = outStr + infoString
		}
		ioutil.WriteFile(p.FileName, []byte(outStr), 0644)
	}

	p.OutStats.Send(outIp)
}

func (p *SummarizeCostGammaPerf) IsConnected() bool {
	return p.In.IsConnected() && p.OutStats.IsConnected()
}

// ================================================================================

type BestEffCostGamma struct {
	ProcName            string
	InCSVFile           *sp.FilePort
	OutBestCost         *sp.ParamPort
	OutBestGamma        *sp.ParamPort
	OutBestEfficiency   *sp.ParamPort
	OutBestValidity     *sp.ParamPort
	Separator           rune
	Header              bool
	EfficiencyValColIdx int // Which column to check for the efficiency value
	ValidityValColIdx   int // Which column to check for the validity value
	IncludeGamma        bool
}

func NewBestEffCostGamma(wf *sp.Workflow, procName string, separator rune, header bool, effValColIdx int, valValColIdx int, includeGamma bool) *BestEffCostGamma {
	sbcr := &BestEffCostGamma{
		ProcName:            procName,
		InCSVFile:           sp.NewFilePort(),
		OutBestCost:         sp.NewParamPort(),
		OutBestGamma:        sp.NewParamPort(),
		OutBestEfficiency:   sp.NewParamPort(),
		OutBestValidity:     sp.NewParamPort(),
		Separator:           separator,
		Header:              header,
		EfficiencyValColIdx: effValColIdx,
		ValidityValColIdx:   valValColIdx,
		IncludeGamma:        includeGamma,
	}
	wf.AddProc(sbcr)
	return sbcr
}

func (p *BestEffCostGamma) Name() string {
	return p.ProcName
}

func (p *BestEffCostGamma) Run() {
	defer p.OutBestCost.Close()
	if p.IncludeGamma {
		defer p.OutBestGamma.Close()
	}
	defer p.OutBestEfficiency.Close()
	defer p.OutBestValidity.Close()

	go p.InCSVFile.RunMergeInputs()

	for iip := range p.InCSVFile.InChan {
		csvData := iip.Read()

		bytesReader := bytes.NewReader(csvData)
		csvReader := csv.NewReader(bytesReader)
		csvReader.Comma = p.Separator

		minEff := 1000000.000 // N.B: The best efficiency in Conformal Prediction is the *minimal* one. Initializing here with an unreasonably large number in order to spot when something is wrong.

		var bestCost int64
		var bestGamma float64 // Only used for libSVM
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

			eff, err := strconv.ParseFloat(rec[p.EfficiencyValColIdx], 64)
			sp.CheckErr(err)

			if eff < minEff {
				minEff = eff

				sp.Debug.Printf("Proc:%s Raw cost value: %s\n", p.Name(), rec[3])
				bestCost, err = strconv.ParseInt(rec[3], 10, 0)

				sp.Debug.Printf("Proc:%s Parsed cost value: %d\n", p.Name(), bestCost)
				sp.CheckErr(err)

				if p.IncludeGamma {
					bestGamma, err = strconv.ParseFloat(rec[4], 64)
					sp.CheckErr(err)
				}

				bestValidity, err = strconv.ParseFloat(rec[p.ValidityValColIdx], 64)
				sp.CheckErr(err)
			}
		}
		sp.Debug.Printf("Final optimal (minimal) efficiency: %f (For: Cost:%03d)\n", minEff, bestCost)
		if p.IncludeGamma {
			sp.Debug.Printf("Final optimal (minimal) efficiency: %f (For: Cost:%03d, Gamma:%.3f)\n", minEff, bestCost, bestGamma)
		}
		p.OutBestCost.Send(fmt.Sprintf("%d", bestCost))
		if p.IncludeGamma {
			p.OutBestGamma.Send(fmt.Sprintf("%.3f", bestGamma))
		}
		p.OutBestEfficiency.Send(fmt.Sprintf("%.3f", minEff))
		p.OutBestValidity.Send(fmt.Sprintf("%.3f", bestValidity))
	}
}

func (p *BestEffCostGamma) IsConnected() bool {
	if p.IncludeGamma {
		return p.InCSVFile.IsConnected() && p.OutBestCost.IsConnected() && p.OutBestGamma.IsConnected() && p.OutBestEfficiency.IsConnected() && p.OutBestValidity.IsConnected()
	}
	return p.InCSVFile.IsConnected() && p.OutBestCost.IsConnected() && p.OutBestEfficiency.IsConnected() && p.OutBestValidity.IsConnected()
}

// ================================================================================

type ParamPrinter struct {
	sp.SciProcess
	ProcName           string
	InParamPorts       map[string]*sp.ParamPort
	OutBestParamsFile  *sp.FilePort
	BestParamsFileName string
}

func NewParamPrinter(wf *sp.Workflow, procName string, fileName string) *ParamPrinter {
	pp := &ParamPrinter{
		ProcName:           procName,
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
	return p.ProcName
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
	ProcName          string
	SummaryFileName   string
	Separator         rune
	InModel           *sp.FilePort
	InTargetDataCount *sp.FilePort
	OutSummary        *sp.FilePort
}

func NewFinalModelSummarizer(wf *sp.Workflow, name string, fileName string, separator rune) *FinalModelSummarizer {
	fms := &FinalModelSummarizer{
		ProcName:          name,
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
	return p.ProcName
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
	for tdip := range p.InTargetDataCount.InChan {
		ai := tdip.GetAuditInfo()
		gene := ai.Params["gene"]
		strs := str.Split(string(tdip.Read()), "\t")
		activeStr := str.TrimSuffix(strs[0], "\n")
		activeCnt, err := strconv.ParseInt(activeStr, 10, 64)
		nonActiveStr := str.TrimSuffix(strs[0], "\n")
		nonActiveCnt, err := strconv.ParseInt(nonActiveStr, 10, 64)
		sp.CheckErr(err)
		activeCounts[gene] = activeCnt
		nonActiveCounts[gene] = nonActiveCnt
	}

	rows := [][]string{[]string{"Gene", "Efficiency", "Validity", "Cost", "ExecTimeMS", "ModelFileSize", "Active", "Nonactive"}}
	for iip := range p.InModel.InChan {
		ai := iip.GetAuditInfo()
		row := []string{
			ai.Params["gene"],
			ai.Params["efficiency"],
			ai.Params["validity"],
			ai.Params["cost"],
			fmt.Sprintf("%d", ai.ExecTimeMS),
			fmt.Sprintf("%d", iip.GetSize()),
			fmt.Sprintf("%d", activeCounts[ai.Params["gene"]]),
			fmt.Sprintf("%d", nonActiveCounts[ai.Params["gene"]]),
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
