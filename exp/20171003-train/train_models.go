// Workflow written in SciPipe.  // For more information about SciPipe, see: http://scipipe.org
package main

import (
	"flag"
	"fmt"
	"runtime"
	"strconv"
	str "strings"

	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
)

var (
	maxTasks = flag.Int("maxtasks", 4, "Max number of local cores to use")
	threads  = flag.Int("threads", 1, "Number of threads that Go is allowed to start")
	geneSet  = flag.String("geneset", "smallest1", "Gene set to use (one of smallest1, smallest3, smallest4, bowes44)")
	runSlurm = flag.Bool("slurm", false, "Start computationally heavy jobs via SLURM")
	debug    = flag.Bool("debug", false, "Increase logging level to include DEBUG messages")

	cpSignPath = "../../bin/cpsign-0.6.3.jar"
	geneSets   = map[string][]string{
		"bowes44": []string{
			// Not available in dataset: "CHRNA1".
			// Not available in dataset: "KCNE1"
			// Instead we use MinK1 as they both share the same alias
			// "MinK", and also confirmed by Wes to be the same.
			"ADORA2A", "ADRA1A", "ADRA2A", "ADRB1", "ADRB2", "CNR1", "CNR2", "CCKAR", "DRD1", "DRD2",
			"EDNRA", "HRH1", "HRH2", "OPRD1", "OPRK1", "OPRM1", "CHRM1", "CHRM2", "CHRM3", "HTR1A",
			"HTR1B", "HTR2A", "HTR2B", "AVPR1A", "CHRNA4", "CACNA1C", "GABRA1", "KCNH2", "KCNQ1", "MINK1",
			"GRIN1", "HTR3A", "SCN5A", "ACHE", "PTGS1", "PTGS2", "MAOA", "PDE3A", "PDE4D", "LCK",
			"SLC6A3", "SLC6A2", "SLC6A4", "AR", "NR3C1",
		},
		"bowes44min100percls": []string{
			"PDE3A", "SCN5A", "CCKAR", "ADRB1", "PTGS1", "CHRM3", "CHRM2", "EDNRA", "MAOA", "LCK",
			"PTGS2", "SLC6A2", "ACHE", "CNR2", "CNR1", "ADORA2A", "OPRD1", "NR3C1", "AR", "SLC6A4",
			"OPRM1", "HTR1A", "SLC6A3", "OPRK1", "AVPR1A", "ADRB2", "DRD2", "KCNH2", "DRD1", "HTR2A",
			"CHRM1",
		},
		"smallest1": []string{
			"PDE3A",
		},
		"smallest3": []string{
			"PDE3A", "SCN5A", "CCKAR",
		},
		"smallest4": []string{
			"PDE3A", "SCN5A", "CCKAR", "ADRB1",
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
	replicates = []string{
		"r1", "r2", "r3",
	}
)

func main() {
	// --------------------------------
	// Parse flags and stuff
	// --------------------------------
	flag.Parse()
	if *debug {
		sp.InitLogDebug()
	} else {
		sp.InitLogAudit()
	}
	if len(geneSets[*geneSet]) == 0 {
		names := []string{}
		for n, _ := range geneSets {
			names = append(names, n)
		}
		sp.Error.Fatalf("Incorrect gene set %s specified! Only allowed values are: %s\n", *geneSet, str.Join(names, ", "))
	}
	runtime.GOMAXPROCS(*threads)

	// --------------------------------
	// Show startup messages
	// --------------------------------
	sp.Info.Printf("Using max %d OS threads to schedule max %d tasks\n", *threads, *maxTasks)
	sp.Info.Printf("Starting workflow for %s geneset\n", *geneSet)

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------
	wf := sp.NewWorkflow("train_models", *maxTasks)

	dbFileName := "pubchem.chembl.dataset4publication_inchi_smiles.tsv.xz"
	dlExcapeDB := wf.NewProc("dlDB", fmt.Sprintf("wget https://zenodo.org/record/173258/files/%s -O {o:excapexz}", dbFileName))
	dlExcapeDB.SetPathStatic("excapexz", "../../raw/"+dbFileName)

	unPackDB := wf.NewProc("unPackDB", "xzcat {i:xzfile} > {o:unxzed}")
	unPackDB.SetPathReplace("xzfile", "unxzed", ".xz", "")
	unPackDB.In("xzfile").Connect(dlExcapeDB.Out("excapexz"))
	//unPackDB.Prepend = "salloc -A snic2017-7-89 -n 2 -t 8:00:00 -J unpack_excapedb"

	finalModelsSummary := NewFinalModelSummarizer(wf, "finalmodels_summary_creator", "res/final_models_summary.tsv", '\t')
	// --------------------------------
	// Set up gene-specific workflow branches
	// --------------------------------
	for _, gene := range geneSets[*geneSet] {
		geneLC := str.ToLower(gene)

		// --------------------------------------------------------------------------------
		// Extract target data step
		// --------------------------------------------------------------------------------
		procName := "extract_target_data_" + geneLC
		extractTargetData := wf.NewProc(procName, `awk -F"\t" '$9 == "{p:gene}" { print $12"\t"$4 }' {i:raw_data} > {o:target_data}`)
		extractTargetData.ParamPort("gene").ConnectStr(gene)
		extractTargetData.SetPathStatic("target_data", fmt.Sprintf("dat/%s/%s.tsv", geneLC, geneLC))
		extractTargetData.In("raw_data").Connect(unPackDB.Out("unxzed"))
		if *runSlurm {
			extractTargetData.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1:00:00 -J scipipe_cnt_comp_" + geneLC // SLURM string
		}

		countTargetDataRows := wf.NewProc("cnt_targetdata_rows_"+geneLC, `awk '$2 == "A" { a += 1 } $2 == "N" { n += 1 } END { print a "\t" n }' {i:targetdata} > {o:count} # {p:gene}`)
		countTargetDataRows.SetPathExtend("targetdata", "count", ".count")
		countTargetDataRows.In("targetdata").Connect(extractTargetData.Out("target_data"))
		countTargetDataRows.ParamPort("gene").ConnectStr(gene)

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

		for _, replicate := range replicates {
			// --------------------------------------------------------------------------------
			// Optimize cost/gamma-step
			// --------------------------------------------------------------------------------
			includeGamma := false // For liblinear
			summarize := NewSummarizeCostGammaPerf(wf,
				"summarize_cost_gamma_perf_"+geneLC+"_"+replicate,
				"dat/"+geneLC+"/"+replicate+"/"+geneLC+"_cost_gamma_perf_stats.tsv",
				includeGamma)

			for _, cost := range costVals {
				// If Liblinear
				unique_string := fmt.Sprintf("liblin_%s_%s_%s", geneLC, cost, replicate) // A string to make process names unique
				evalCostCmdLiblin := `java -jar ` + cpSignPath + ` crossvalidate \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--trainfile {i:traindata} \
									--impl liblinear \
									--labels A, N \
									--nr-models {p:nrmdl} \
									--cost {p:cost} \
									--cv-folds {p:cvfolds} \
									--output-format json \
									--confidence {p:confidence} | grep -P "^{" > {o:stats} # {p:gene} {p:replicate}`
				evalCostPathFunc := func(t *sp.SciTask) string {
					c, err := strconv.ParseInt(t.Param("cost"), 10, 0)
					sp.CheckErr(err)
					return str.Replace(t.InPath("traindata"), geneLC+".tsv", replicate+"/"+geneLC+".tsv", 1) + fmt.Sprintf(".liblin_c%03d", c) + "_crossval_stats.json"
				}
				evalCost := wf.NewProc("crossval_"+unique_string, evalCostCmdLiblin)
				evalCost.SetPathCustom("stats", evalCostPathFunc)
				evalCost.In("traindata").Connect(extractTargetData.Out("target_data"))
				evalCost.ParamPort("nrmdl").ConnectStr("10")
				evalCost.ParamPort("cvfolds").ConnectStr("10")
				evalCost.ParamPort("confidence").ConnectStr("0.9")
				evalCost.ParamPort("gene").ConnectStr(gene)
				evalCost.ParamPort("replicate").ConnectStr(replicate)
				evalCost.ParamPort("cost").ConnectStr(cost)
				if *runSlurm {
					evalCost.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J evalcg_" + unique_string // SLURM string
				}

				extractCostGammaStats := spc.NewMapToKeys(wf, "extract_cgstats_"+unique_string, func(ip *sp.InformationPacket) map[string]string {
					crossValOut := &cpSignCrossValOutput{}
					ip.UnMarshalJson(crossValOut)
					newKeys := map[string]string{}
					newKeys["validity"] = fmt.Sprintf("%.3f", crossValOut.Validity)
					newKeys["efficiency"] = fmt.Sprintf("%.3f", crossValOut.Efficiency)
					newKeys["obsfuzz_active"] = fmt.Sprintf("%.3f", crossValOut.ObservedFuzziness.Active)
					newKeys["obsfuzz_nonactive"] = fmt.Sprintf("%.3f", crossValOut.ObservedFuzziness.Nonactive)
					newKeys["obsfuzz_overall"] = fmt.Sprintf("%.3f", crossValOut.ObservedFuzziness.Overall)
					return newKeys
				})
				extractCostGammaStats.In.Connect(evalCost.Out("stats"))

				summarize.In.Connect(extractCostGammaStats.Out)
				//}
			}
			selectBest := NewBestCostGamma(wf,
				"select_best_cost_gamma_"+geneLC+"_"+replicate,
				'\t',
				false,
				includeGamma)
			selectBest.InCSVFile.Connect(summarize.OutStats)
			// --------------------------------------------------------------------------------
			// Train step
			// --------------------------------------------------------------------------------
			cpSignTrain := wf.NewProc("cpsign_train_"+geneLC+"_"+replicate,
				`java -jar `+cpSignPath+` train \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--modelfile {i:model} \
									--labels A, N \
									--impl liblinear \
									--nr-models {p:nrmdl} \
									--cost {p:cost} \
									--model-out {o:model} \
									--model-name "{p:gene} target profile" # (Cost-Equalized Observed Fuzziness: {p:clsavgobsfuzz}, Efficiency: {p:efficiency}, Validity: {p:validity}, Replicate: {p:replicate})`)
			cpSignTrainPathFunc := func(t *sp.SciTask) string {
				return fmt.Sprintf("dat/final_models/"+replicate+"/%s_%s_c%s_nrmdl%s.mdl",
					str.ToLower(t.Param("gene")),
					"liblin",
					t.Param("cost"),
					t.Param("nrmdl"))
			}

			cpSignTrain.In("model").Connect(cpSignPrecomp.Out("precomp"))
			cpSignTrain.ParamPort("nrmdl").ConnectStr("10")
			cpSignTrain.ParamPort("cost").Connect(selectBest.OutBestCost)
			cpSignTrain.ParamPort("gene").ConnectStr(gene)
			cpSignTrain.ParamPort("replicate").ConnectStr(replicate)
			cpSignTrain.ParamPort("clsavgobsfuzz").Connect(selectBest.OutBestClassAvgObsFuzz)
			cpSignTrain.ParamPort("efficiency").Connect(selectBest.OutBestEfficiency)
			cpSignTrain.ParamPort("validity").Connect(selectBest.OutBestValidity)
			cpSignTrain.SetPathCustom("model", cpSignTrainPathFunc)
			if *runSlurm {
				cpSignTrain.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J train_" + geneLC // SLURM string
			}

			finalModelsSummary.InModel.Connect(cpSignTrain.Out("model"))
			finalModelsSummary.InTargetDataCount.Connect(countTargetDataRows.Out("count"))
		} // end for replicate
	} // end for gene

	sortSummaryOnDataSize := wf.NewProc("sort_summary", "sort -n -k 10 {i:summary} > {o:sorted}")
	sortSummaryOnDataSize.SetPathReplace("summary", "sorted", ".tsv", ".sorted.tsv")
	sortSummaryOnDataSize.In("summary").Connect(finalModelsSummary.OutSummary)

	plotSummary := wf.NewProc("plot_summary", "Rscript bin/plot_summary.r -i {i:summary} -o {o:plot} -f png")
	plotSummary.SetPathExtend("summary", "plot", ".plot.png")
	plotSummary.In("summary").Connect(sortSummaryOnDataSize.Out("sorted"))

	wf.ConnectLast(plotSummary.Out("plot"))

	// --------------------------------
	// Run the pipeline!
	// --------------------------------
	wf.Run()
}

// --------------------------------------------------------------------------------
// JSON types
// --------------------------------------------------------------------------------
// JSON output of cpSign crossvalidate
// {
//     "classConfidence": 0.855,
//     "observedFuzziness": {
//         "A": 0.253,
//         "N": 0.207,
//         "overall": 0.231
//     },
//     "validity": 0.917,
//     "efficiency": 0.333,
//     "classCredibility": 0.631
// }
// --------------------------------------------------------------------------------

type cpSignCrossValOutput struct {
	ClassConfidence   float64                 `json:"classConfidence"`
	ObservedFuzziness cpSignObservedFuzziness `json:"observedFuzziness"`
	Validity          float64                 `json:"validity"`
	Efficiency        float64                 `json:"efficiency"`
	ClassCredibility  float64                 `json:"classCredibility"`
}

type cpSignObservedFuzziness struct {
	Active    float64 `json:"A"`
	Nonactive float64 `json:"N"`
	Overall   float64 `json:"overall"`
}
