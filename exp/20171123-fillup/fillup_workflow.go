// Workflow written in SciPipe.  // For more information about SciPipe, see: http://scipipe.org
package main

import (
	"flag"
	"fmt"
	"path/filepath"
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
		"bowes44min100percls_small": []string{ // These are the ones for which we want to fill up with assumed negative
			"PDE3A",
			"SCN5A",
			"CCKAR",
			"ADRB1",
			"PTGS1",
			"CHRM3",
			"CHRM2",
			"EDNRA",
			"MAOA",
			"LCK",
			"PTGS2",
			"SLC6A2",
			"ACHE",
			"CNR2",
			"CNR1",
			"ADORA2A",
			"OPRD1",
			"NR3C1",
			"AR",
			"SLC6A4",
			"OPRM1",
		},
		"bowes44min100percls_large": []string{
			"HTR1A",
			"SLC6A3",
			"OPRK1",
			"AVPR1A",
			"ADRB2",
			"DRD2",
			"KCNH2",
			"DRD1",
			"HTR2A",
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
	sp.Audit.Printf("Using max %d OS threads to schedule max %d tasks\n", *threads, *maxTasks)
	sp.Audit.Printf("Starting workflow for %s geneset\n", *geneSet)

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

	// extractGSA extracts a file with only Gene symbol, SMILES and the
	// Activity flag, into a .tsv file, for easier subsequent parsing
	extractGSA := wf.NewProc("extract_gene_smiles_activity", `awk -F "\t" '{ print $9 "\t" $12 "\t" $4 }' {i:excapedb} | sort -uV > {os:gene_smiles_activity}`) // os = output (streaming) ... stream output via a fifo file
	extractGSA.SetPathReplace("excapedb", "gene_smiles_activity", ".tsv", ".ext_gene_smiles_activity.tsv")
	extractGSA.In("excapedb").Connect(unPackDB.Out("unxzed"))

	// removeConflicting removes (or, SHOULD remove) rows which have the same values on both row 1 and 2 (I think ...)
	removeConflicting := wf.NewProc("remove_conflicting", `awk -F"\t" '(( $1 != p1 ) || ( $2 != p2)) && ( c[p1,p2] <= 1 ) && ( p1 != "" ) && ( p2 != "" ) { print p1 "\t" p2 "\t" p3 }
																	  { c[$1,$2]++; p1 = $1; p2 = $2; p3 = $3 }
																	  END { print $1 "\t" $2 "\t" $3 }' \
																	  {i:gene_smiles_activity} > {o:gene_smiles_activity}`)
	removeConflicting.SetPathReplace("gene_smiles_activity", "gene_smiles_activity", ".tsv", ".dedup.tsv")
	removeConflicting.In("gene_smiles_activity").Connect(extractGSA.Out("gene_smiles_activity"))

	//createSQLiteDB := wf.NewProc("CreateSQLiteDB", `sqlite3 {o:dbfile} 'CREATE TABLE {p:tablename} (Ambit_InchiKey TEXT, Original_Entry_ID TEXT, Entrez_ID INTEGER, Activity_Flag TEXT, pXC50 REAL, DB TEXT, Original_Assay_ID INTEGER, Tax_ID INTEGER, Gene_Symbol TEXT, Ortholog_Group INTEGER, InChI TEXT, SMILES TEXT);'`)
	//createSQLiteDB.SetPathStatic("dbfile", "../../raw/excapedb.db")
	//createSQLiteDB.ParamInPort("tablename").ConnectStr("excapedb")

	//importIntoSQLite := wf.NewProc("importIntoSQLite", `(echo '.mode tabs'; echo '.import {i:tsvfile} {p:tablename}'; echo 'CREATE INDEX Gene_Symbol ON excapedb(Gene_Symbol);'; echo 'CREATE INDEX Activity_Flag ON excapedb(Activity_Flag);') | sqlite3 {i:dbfile} && echo done > {o:doneflagfile};`)
	//importIntoSQLite.SetPathExtend("tsvfile", "doneflagfile", ".importdone")
	//importIntoSQLite.ParamInPort("tablename").ConnectStr("excapedb")
	//importIntoSQLite.In("dbfile").Connect(createSQLiteDB.Out("dbfile"))
	//importIntoSQLite.In("tsvfile").Connect(unPackDB.Out("unxzed"))

	finalModelsSummary := NewFinalModelSummarizer(wf, "finalmodels_summary_creator", "res/final_models_summary.tsv", '\t')

	// --------------------------------
	// Set up gene-specific workflow branches
	// --------------------------------
	for _, geneUppercase := range geneSets[*geneSet] {
		doFillUp := strInSlice(geneUppercase, geneSets["bowes44min100percls_small"])

		geneLowerCase := str.ToLower(geneUppercase)
		uniqStrGene := geneLowerCase

		// extractTargetData extract all data for the specific target, into a separate file
		extractTargetData := wf.NewProc("extract_target_data_"+uniqStrGene, `awk -F"\t" '$1 == "{p:gene}" { print $2"\t"$3 }' {i:raw_data} > {o:target_data}`)
		extractTargetData.ParamInPort("gene").ConnectStr(geneUppercase)
		extractTargetData.SetPathStatic("target_data", fmt.Sprintf("dat/%s/%s.tsv", geneLowerCase, geneLowerCase))
		extractTargetData.In("raw_data").Connect(removeConflicting.Out("gene_smiles_activity"))
		if *runSlurm {
			extractTargetData.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1:00:00 -J scipipe_extract_" + geneLowerCase // SLURM string
		}

		countTargetDataRows := wf.NewProc("cnt_targetdata_rows_"+uniqStrGene, `awk '$2 == "A" { a += 1 } $2 == "N" { n += 1 } END { print a "\t" n }' {i:targetdata} > {o:count} # {p:gene}`)
		countTargetDataRows.SetPathExtend("targetdata", "count", ".count")
		countTargetDataRows.In("targetdata").Connect(extractTargetData.Out("target_data"))
		countTargetDataRows.ParamInPort("gene").ConnectStr(geneUppercase)

		var targetDataPort *sp.OutPort

		if doFillUp {
			createRandomBytes := wf.NewProc("create_random_bytes_"+uniqStrGene, "dd if=/dev/urandom of={o:rand} bs=1048576 count=1024")
			createRandomBytes.SetPathStatic("rand", "dat/"+geneLowerCase+"_random_bytes.bin")

			fillAssumedNonbinding := wf.NewProc("fillup_"+uniqStrGene, `
			cat {i:targetdata} > {o:filledup} && \
				let "fillup_lines_cnt = "$(wc -l {i:targetdata} | awk '{ printf $1 }')" * 2" \
				&& cat {i:rawdata} \
				| awk -F"\t" '$1 != "{p:gene}" { print $2 "\tN" }' \
				| shuf --random-source={i:randsrc} -n $fillup_lines_cnt >> {o:filledup}
			`)
			fillAssumedNonbinding.SetPathReplace("targetdata", "filledup", ".tsv", ".incl_assumed_neg.tsv")
			fillAssumedNonbinding.In("targetdata").Connect(extractTargetData.Out("target_data"))
			fillAssumedNonbinding.In("rawdata").Connect(removeConflicting.Out("gene_smiles_activity"))
			fillAssumedNonbinding.ParamInPort("gene").ConnectStr(geneUppercase)
			fillAssumedNonbinding.In("randsrc").Connect(createRandomBytes.Out("rand"))
			targetDataPort = fillAssumedNonbinding.Out("filledup")
		} else {
			targetDataPort = extractTargetData.Out("target_data")
		}

		// --------------------------------------------------------------------------------
		// Pre-compute step
		// --------------------------------------------------------------------------------
		cpSignPrecomp := wf.NewProc("cpsign_precomp_"+uniqStrGene,
			`java -jar `+cpSignPath+` precompute \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--trainfile {i:traindata} \
									--labels A, N \
									--model-out {o:precomp} \
									--model-name "`+geneUppercase+` target profile"`)
		cpSignPrecomp.In("traindata").Connect(targetDataPort)
		cpSignPrecomp.SetPathExtend("traindata", "precomp", ".precomp")
		if *runSlurm {
			cpSignPrecomp.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J precmp_" + geneLowerCase // SLURM string
		}

		for _, replicate := range replicates {
			uniqStrRepl := uniqStrGene + "_" + replicate

			// --------------------------------------------------------------------------------
			// Optimize cost/gamma-step
			// --------------------------------------------------------------------------------
			includeGamma := false // For liblinear
			summarize := NewSummarizeCostGammaPerf(wf,
				"summarize_cost_gamma_perf_"+uniqStrRepl,
				"dat/"+geneLowerCase+"/"+replicate+"/"+geneLowerCase+"_cost_gamma_perf_stats.tsv",
				includeGamma)

			for _, cost := range costVals {
				uniqStrCost := uniqStrRepl + "_" + cost
				// If Liblinear
				evalCost := wf.NewProc("crossval_"+uniqStrCost, `java -jar `+cpSignPath+` crossvalidate \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--trainfile {i:traindata} \
									--impl liblinear \
									--labels A, N \
									--nr-models {p:nrmdl} \
									--cost {p:cost} \
									--cv-folds {p:cvfolds} \
									--output-format json \
									--confidence {p:confidence} | grep -P "^{" > {o:stats} # {p:gene} {p:replicate}`)
				evalCost.SetPathCustom("stats", func(t *sp.Task) string {
					c, err := strconv.ParseInt(t.Param("cost"), 10, 0)
					sp.CheckErr(err)
					trainDataBasePath := filepath.Base(t.InPath("traindata"))
					return str.Replace(t.InPath("traindata"), trainDataBasePath, t.Param("replicate")+"/"+trainDataBasePath, 1) + fmt.Sprintf(".liblin_c%03d", c) + "_crossval_stats.json"
				})
				evalCost.In("traindata").Connect(targetDataPort)
				evalCost.ParamInPort("nrmdl").ConnectStr("10")
				evalCost.ParamInPort("cvfolds").ConnectStr("10")
				evalCost.ParamInPort("confidence").ConnectStr("0.9")
				evalCost.ParamInPort("gene").ConnectStr(geneUppercase)
				evalCost.ParamInPort("replicate").ConnectStr(replicate)
				evalCost.ParamInPort("cost").ConnectStr(cost)
				if *runSlurm {
					evalCost.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J evalcg_" + uniqStrCost // SLURM string
				}

				extractCostGammaStats := spc.NewMapToKeys(wf, "extract_cgstats_"+uniqStrCost, func(ip *sp.IP) map[string]string {
					crossValOut := &cpSignCrossValOutput{}
					ip.UnMarshalJSON(crossValOut)
					newKeys := map[string]string{}
					newKeys["validity"] = fmt.Sprintf("%.3f", crossValOut.Validity)
					newKeys["efficiency"] = fmt.Sprintf("%.3f", crossValOut.Efficiency)
					newKeys["class_confidence"] = fmt.Sprintf("%.3f", crossValOut.ClassConfidence)
					newKeys["class_credibility"] = fmt.Sprintf("%.3f", crossValOut.ClassCredibility)
					newKeys["obsfuzz_active"] = fmt.Sprintf("%.3f", crossValOut.ObservedFuzziness.Active)
					newKeys["obsfuzz_nonactive"] = fmt.Sprintf("%.3f", crossValOut.ObservedFuzziness.Nonactive)
					newKeys["obsfuzz_overall"] = fmt.Sprintf("%.3f", crossValOut.ObservedFuzziness.Overall)
					return newKeys
				})
				extractCostGammaStats.In().Connect(evalCost.Out("stats"))

				summarize.In().Connect(extractCostGammaStats.Out())
			} // end for cost

			// TODO: Let select best operate directly on the stream of IPs, not
			// via the summarize component, so that we can retain the keys in
			// the IP!
			selectBest := NewBestCostGamma(wf,
				"select_best_cost_gamma_"+uniqStrRepl,
				'\t',
				false,
				includeGamma)
			selectBest.InCSVFile().Connect(summarize.OutStats())

			// --------------------------------------------------------------------------------
			// Train step
			// --------------------------------------------------------------------------------
			cpSignTrain := wf.NewProc("cpsign_train_"+uniqStrRepl,
				`java -jar `+cpSignPath+` train \
									--license ../../bin/cpsign.lic \
									--cptype 1 \
									--modelfile {i:model} \
									--labels A, N \
									--impl liblinear \
									--nr-models {p:nrmdl} \
									--cost {p:cost} \
									--model-out {o:model} \
									--model-name "{p:gene} target profile" # {p:replicate} Validity: {p:validity} Efficiency: {p:efficiency} Class-Equalized Observed Fuzziness: {p:obsfuzz_classavg} Observed Fuzziness (Overall): {p:obsfuzz_overall} Observed Fuzziness (Active class): {p:obsfuzz_active} Observed Fuzziness (Non-active class): {p:obsfuzz_nonactive} Class Confidence: {p:class_confidence} Class Credibility: {p:class_credibility}`)
			cpSignTrain.In("model").Connect(cpSignPrecomp.Out("precomp"))
			cpSignTrain.ParamInPort("nrmdl").ConnectStr("10")
			cpSignTrain.ParamInPort("gene").ConnectStr(geneUppercase)
			cpSignTrain.ParamInPort("replicate").ConnectStr(replicate)
			cpSignTrain.ParamInPort("validity").Connect(selectBest.OutBestValidity())
			cpSignTrain.ParamInPort("efficiency").Connect(selectBest.OutBestEfficiency())
			cpSignTrain.ParamInPort("obsfuzz_classavg").Connect(selectBest.OutBestObsFuzzClassAvg())
			cpSignTrain.ParamInPort("obsfuzz_overall").Connect(selectBest.OutBestObsFuzzOverall())
			cpSignTrain.ParamInPort("obsfuzz_active").Connect(selectBest.OutBestObsFuzzActive())
			cpSignTrain.ParamInPort("obsfuzz_nonactive").Connect(selectBest.OutBestObsFuzzNonactive())
			cpSignTrain.ParamInPort("class_confidence").Connect(selectBest.OutBestClassConfidence())
			cpSignTrain.ParamInPort("class_credibility").Connect(selectBest.OutBestClassCredibility())
			cpSignTrain.ParamInPort("cost").Connect(selectBest.OutBestCost())
			cpSignTrain.SetPathCustom("model", func(t *sp.Task) string {
				return fmt.Sprintf("dat/final_models/%s/%s_c%s_nrmdl%s_%s.mdl",
					str.ToLower(t.Param("gene")),
					"liblin",
					t.Param("cost"),
					t.Param("nrmdl"),
					t.Param("replicate"))
			})
			if *runSlurm {
				cpSignTrain.Prepend = "salloc -A snic2017-7-89 -n 4 -c 4 -t 1-00:00:00 -J train_" + uniqStrRepl // SLURM string
			}

			finalModelsSummary.InModel().Connect(cpSignTrain.Out("model"))
		} // end for replicate
		finalModelsSummary.InTargetDataCount().Connect(countTargetDataRows.Out("count"))
	} // end for gene

	sortSummaryOnDataSize := wf.NewProc("sort_summary", "head -n 1 {i:summary} > {o:sorted} && tail -n +2 {i:summary} | sort -nk 16 >> {o:sorted}")
	sortSummaryOnDataSize.SetPathReplace("summary", "sorted", ".tsv", ".sorted.tsv")
	sortSummaryOnDataSize.In("summary").Connect(finalModelsSummary.OutSummary())

	plotSummary := wf.NewProc("plot_summary", "Rscript bin/plot_summary.r -i {i:summary} -o {o:plot} -f png # {i:gene_smiles_activity}")
	plotSummary.SetPathExtend("summary", "plot", ".plot.png")
	plotSummary.In("summary").Connect(sortSummaryOnDataSize.Out("sorted"))
	//plotSummary.In("importedflagfile").Connect(importIntoSQLite.Out("doneflagfile"))
	plotSummary.In("gene_smiles_activity").Connect(removeConflicting.Out("gene_smiles_activity"))

	// --------------------------------
	// Run the pipeline!
	// --------------------------------

	procsToRun := []string{}
	for _, gene := range geneSets[*geneSet] {
		procsToRun = append(procsToRun, "fillup_"+str.ToLower(gene))
	}
	wf.RunToProcNames(procsToRun...)
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

func strInSlice(searchStr string, strings []string) bool {
	for _, str := range strings {
		if searchStr == str {
			return true
		}
	}
	return false
}
