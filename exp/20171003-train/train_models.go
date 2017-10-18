// Workflow written in SciPipe.
// For more information about SciPipe, see: http://scipipe.org
package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	sp "github.com/scipipe/scipipe"
	"io/ioutil"
	"regexp"
	"runtime"
	"strconv"
	str "strings"
)

var (
	maxTasks = flag.Int("maxtasks", 4, "Max number of local cores to use")
	threads  = flag.Int("threads", 1, "Number of threads that Go is allowed to start")
	geneSet  = flag.String("geneset", "smallest1", "Gene set to use (one of smallest1, smallest3, smallest4, bowes44)")
	runSlurm = flag.Bool("slurm", false, "Start computationally heavy jobs via SLURM")

	cpSignPath = "../../bin/cpsign-0.6.2.jar"
	geneSets   = map[string][]string{
		"bowes44": []string{
			// Not available in dataset: "CHRNA1", Not available in dataset:
			// "KCNE1". Instead we use MinK1 as they both share the same alias
			// "MinK", and also confirmed by Wes to be the same.
			"ADORA2A", "ADRA1A", "ADRA2A", "ADRB1", "ADRB2",
			"CNR1", "CNR2", "CCKAR", "DRD1", "DRD2",
			"EDNRA", "HRH1", "HRH2", "OPRD1", "OPRK1",
			"OPRM1", "CHRM1", "CHRM2", "CHRM3", "HTR1A",
			"HTR1B", "HTR2A", "HTR2B", "AVPR1A", "CHRNA4",
			"CACNA1C", "GABRA1", "KCNH2", "KCNQ1", "MINK1",
			"GRIN1", "HTR3A", "SCN5A", "ACHE", "PTGS1",
			"PTGS2", "MAOA", "PDE3A", "PDE4D", "LCK",
			"SLC6A3", "SLC6A2", "SLC6A4", "AR", "NR3C1",
		},
		"smallest1": []string{
			"GABRA1",
		},
		"smallest3": []string{
			"GABRA1",
			"CACNA1C",
			"CHRNA4",
		},
		"smallest4": []string{
			"GABRA1",
			"CACNA1C",
			"CHRNA4",
			"PDE3A",
		},
	}
	costVals = []string{
		"1",
		"10",
		"100",
	}
	gammaVals = []string{
		"0.1",
		"0.01",
		"0.001",
	}
)

func main() {
	sp.InitLogDebug()
	flag.Parse()

	sp.Info.Printf("Using max %d OS threads to schedule max %d tasks\n", *threads, *maxTasks)
	sp.Info.Printf("Starting workflow for %s geneset\n", *geneSet)

	runtime.GOMAXPROCS(*threads)

	//sp.InitLogDebug()
	wf := sp.NewWorkflow("train_models", *maxTasks)

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------
	dbFileName := "pubchem.chembl.dataset4publication_inchi_smiles.tsv.xz"
	dlExcapeDB := wf.NewProc("dlDB", fmt.Sprintf("wget https://zenodo.org/record/173258/files/%s -O {o:excapexz}", dbFileName))
	dlExcapeDB.SetPathStatic("excapexz", "../../raw/"+dbFileName)

	unPackDB := wf.NewProc("unPackDB", "xzcat {i:xzfile} > {o:unxzed}")
	unPackDB.SetPathReplace("xzfile", "unxzed", ".xz", "")
	unPackDB.In("xzfile").Connect(dlExcapeDB.Out("excapexz"))
	//unPackDB.Prepend = "salloc -A snic2017-7-89 -n 2 -t 8:00:00 -J unpack_excapedb"

	// --------------------------------
	// Set up gene-specific workflow branches
	// --------------------------------
	for _, gene := range geneSets[*geneSet] {
		geneLC := str.ToLower(gene)

		// --------------------------------------------------------------------------------
		// Extract target data step
		// --------------------------------------------------------------------------------
		procName := "extract_target_data_" + geneLC
		extractTargetData := wf.NewProc(procName, fmt.Sprintf(`awk -F"\t" '$9 == "%s" { print $12"\t"$4 }' {i:raw_data} > {o:target_data}`, gene))
		extractTargetData.SetPathStatic("target_data", fmt.Sprintf("dat/%s/%s.tsv", geneLC, geneLC))
		extractTargetData.In("raw_data").Connect(unPackDB.Out("unxzed"))
		if *runSlurm {
			extractTargetData.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1:00:00 -J scipipe_cnt_comp_" + geneLC // SLURM string
		}

		// --------------------------------------------------------------------------------
		// Optimize cost/gamma-step
		// --------------------------------------------------------------------------------
		includeGamma := false // For liblinear
		summarize := NewSummarizeCostGammaPerf(wf, "summarize_cost_gamma_perf_"+geneLC, "dat/"+geneLC+"/"+geneLC+"_cost_gamma_perf_stats.tsv", includeGamma)
		for _, cost := range costVals {
			//for _, gamma := range gammaVals {

			// If Liblinear
			unique_string := fmt.Sprintf("liblin_%s_%s", geneLC, cost) // A string to make process names unique
			crossvalCmdLiblin := `java -jar ` + cpSignPath + ` crossvalidate \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--trainfile {i:traindata} \
									--impl liblinear \
									--labels A, N \
									--nr-models {p:nrmdl} \
									--cost {p:cost} \
									--cv-folds {p:cvfolds} \
									--confidence {p:confidence} > {o:stats} # {p:gene}`
			pathFuncLibLin := func(t *sp.SciTask) string {
				c, err := strconv.ParseInt(t.Param("cost"), 10, 0)
				sp.CheckErr(err)
				return t.InPath("traindata") + fmt.Sprintf(".liblin_c%03d", c) + ".stats.txt"
			}

			// If LibSVM
			//unique_string := fmt.Sprintf("libsvm_%s_%s_%s", geneLC, cost, gamma) // A string to make process names unique
			//crossvalCmdLibSVM := `java -jar ` + cpSignPath + ` crossvalidate \
			//					--license ../../bin/cpsign.lic \
			//					--cptype 1 \
			//					--trainfile {i:traindata} \
			//					--impl libsvm \
			//					--labels A, N \
			//					--nr-models {p:nrmdl} \
			//					--cost {p:cost} \
			//					--gamma {p:gamma} \
			//					--cv-folds {p:cvfolds} \
			//					--confidence {p:confidence} > {o:stats} # {p:gene}`
			//pathFuncLibSVM := func(t *sp.SciTask) string {
			//	c, err := strconv.ParseInt(t.Param("cost"), 10, 0)
			//	sp.CheckErr(err)
			//	g, err := strconv.ParseFloat(t.Param("gamma"), 64)
			//	sp.CheckErr(err)
			//	return t.InPath("traindata") + fmt.Sprintf(".libsvm_c%03d_g%.3f", c, g) + ".stats.txt"
			//}

			evalCostGamma := wf.NewProc("crossval_"+unique_string, crossvalCmdLiblin)
			evalCostGamma.SetPathCustom("stats", pathFuncLibLin)
			// Connect
			evalCostGamma.In("traindata").Connect(extractTargetData.Out("target_data"))
			evalCostGamma.ParamPort("nrmdl").ConnectStr("10")
			evalCostGamma.ParamPort("cvfolds").ConnectStr("10")
			evalCostGamma.ParamPort("confidence").ConnectStr("0.9")
			evalCostGamma.ParamPort("gene").ConnectStr(gene)
			evalCostGamma.ParamPort("cost").ConnectStr(cost)
			//evalCostGamma.ParamPort("gamma").ConnectStr(gamma) // Only used with liblinear
			if *runSlurm {
				evalCostGamma.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J evalcg_" + unique_string // SLURM string
			}

			summarize.In.Connect(evalCostGamma.Out("stats"))
			//}
		}
		selectBest := NewBestEffCostGamma(wf, "select_best_cost_gamma_"+geneLC, '\t', false, 1, includeGamma)
		selectBest.InCSVFile.Connect(summarize.OutStats)

		// --------------------------------------------------------------------------------
		// Pre-compute step
		// --------------------------------------------------------------------------------
		cpSignPrecomp := wf.NewProc("cpsign_precomp_"+geneLC,
			`java -jar `+cpSignPath+` precompute \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--trainfile {i:traindata} \
									--labels A, N \
									--model-out {o:precomp} \
									--model-name "`+gene+` target profile"`)
		cpSignPrecomp.In("traindata").Connect(extractTargetData.Out("target_data"))
		cpSignPrecomp.SetPathExtend("traindata", "precomp", ".precomp")
		if *runSlurm {
			cpSignPrecomp.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J precmp_" + geneLC // SLURM string
		}

		// --------------------------------------------------------------------------------
		// Train step
		// --------------------------------------------------------------------------------
		// For LibSVM
		//cpSignTrain:= wf.NewProc("cpsign_train_"+geneLC,
		//	`java -jar `+cpSignPath+` train \
		//							--license ../../bin/cpsign.lic \
		//							--cptype 1 \
		//							--modelfile {i:model} \
		//							--labels A, N \
		//							--impl libsvm \
		//							--nr-models {p:nrmdl} \
		//							--cost {p:cost} \
		//							--gamma {p:gamma} \
		//							--model-out {o:model} \
		//							--model-name "{p:gene} target profile" # Efficiency: {p:efficiency}`)
		//cpSignTrainPathFunc := func(t *sp.SciTask) string {
		//	return fmt.Sprintf("dat/final_models/%s_c%s_g%s_nrmdl%s.mdl",
		//		str.ToLower(t.Param("gene")),
		//		t.Param("cost"),
		//		t.Param("gamma"),
		//		t.Param("nrmdl"))
		//}
		cpSignTrain := wf.NewProc("cpsign_train_"+geneLC,
			`java -jar `+cpSignPath+` train \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--modelfile {i:model} \
									--labels A, N \
									--impl liblinear \
									--nr-models {p:nrmdl} \
									--cost {p:cost} \
									--model-out {o:model} \
									--model-name "{p:gene} target profile" # Efficiency: {p:efficiency}`)
		cpSignTrainPathFunc := func(t *sp.SciTask) string {
			return fmt.Sprintf("dat/final_models/%s_%s_c%s_nrmdl%s.mdl",
				str.ToLower(t.Param("gene")),
				"liblin",
				t.Param("cost"),
				t.Param("nrmdl"))
		}

		cpSignTrain.In("model").Connect(cpSignPrecomp.Out("precomp"))
		cpSignTrain.ParamPort("nrmdl").ConnectStr("10")
		cpSignTrain.ParamPort("cost").Connect(selectBest.OutBestCost)
		//cpSignTrain.ParamPort("gamma").Connect(selectBest.OutBestGamma)
		cpSignTrain.ParamPort("gene").ConnectStr(gene)
		cpSignTrain.ParamPort("efficiency").Connect(selectBest.OutBestEfficiency)
		cpSignTrain.SetPathCustom("model", cpSignTrainPathFunc)
		if *runSlurm {
			cpSignTrain.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J train_" + geneLC // SLURM string
		}

		//paramPrinter := NewParamPrinter(wf, "param_printer_"+geneLC, "dat/best_cost_gamma_"+geneLC+".txt")
		//paramPrinter.GetParamPort("cost").Connect(selectBest.OutBestCost)
		//paramPrinter.GetParamPort("gamma").Connect(selectBest.OutBestGamma)
		//paramPrinter.GetParamPort("efficiency").Connect(selectBest.OutBestEfficiency)

		wf.ConnectLast(cpSignTrain.Out("model"))
	}

	// --------------------------------
	// Run the pipeline!
	// --------------------------------
	wf.Run()
}

// --------------------------------------------------------------------------------

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

// --------------------------------------------------------------------------------

type BestEffCostGamma struct {
	ProcName          string
	InCSVFile         *sp.FilePort
	OutBestCost       *sp.ParamPort
	OutBestGamma      *sp.ParamPort
	OutBestEfficiency *sp.ParamPort
	Separator         rune
	Header            bool
	EffValColIdx      int // Which column to check for the efficiency value
	IncludeGamma      bool
}

func NewBestEffCostGamma(wf *sp.Workflow, procName string, separator rune, header bool, effValColIdx int, includeGamma bool) *BestEffCostGamma {
	sbcr := &BestEffCostGamma{
		ProcName:          procName,
		InCSVFile:         sp.NewFilePort(),
		OutBestCost:       sp.NewParamPort(),
		OutBestGamma:      sp.NewParamPort(),
		OutBestEfficiency: sp.NewParamPort(),
		Separator:         separator,
		Header:            header,
		EffValColIdx:      effValColIdx,
		IncludeGamma:      includeGamma,
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
	go p.InCSVFile.RunMergeInputs()

	for iip := range p.InCSVFile.InChan {
		csvData := iip.Read()

		bytesReader := bytes.NewReader(csvData)
		csvReader := csv.NewReader(bytesReader)
		csvReader.Comma = p.Separator

		minEff := 1000000.000 // N.B: The best efficiency in Conformal Prediction is the *minimal* one. Initializing here with an unreasonably large number in order to spot when something is wrong.
		var bestCost int64
		var bestGamma float64 // Only used for libSVM

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
			eff, err := strconv.ParseFloat(rec[p.EffValColIdx], 64)
			sp.CheckErr(err)
			if eff < minEff {
				minEff = eff

				sp.Debug.Printf("Proc:%s Raw cost value: %s\n", p.Name(), rec[3])
				bestCost, err = strconv.ParseInt(rec[3], 10, 0)
				sp.Debug.Printf("Proc:%s Parsed cost value: %d\n", p.Name(), rec[3])
				sp.CheckErr(err)

				if p.IncludeGamma {
					bestGamma, err = strconv.ParseFloat(rec[4], 64)
					sp.CheckErr(err)
				}
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
	}
}

func (p *BestEffCostGamma) IsConnected() bool {
	if p.IncludeGamma {
		return p.InCSVFile.IsConnected() && p.OutBestCost.IsConnected() && p.OutBestGamma.IsConnected() && p.OutBestEfficiency.IsConnected()
	}
	return p.InCSVFile.IsConnected() && p.OutBestCost.IsConnected() && p.OutBestEfficiency.IsConnected()
}

// --------------------------------------------------------------------------------

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
