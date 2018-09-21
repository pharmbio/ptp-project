package main

import (
	"fmt"
	"path/filepath"
	"regexp"

	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
)

func main() {
	wf := sp.NewWorkflow("extract_valdata", 1)

	for confIdx, confLevel := range []string{"0p9", "0p8"} { // We use 'p' instead of '.' to avoid confusion in the file name
		confLevel := confLevel

		// ------------------------------------------------------------------------
		// Grab input files
		// ------------------------------------------------------------------------
		validateFiles := spc.NewFileGlobber(wf, "valstat_files_"+confLevel, "dat/validate/*/*1000.json")
		sts := spc.NewStreamToSubStream(wf, "sts_"+confLevel)
		sts.In().From(validateFiles.Out())

		// ------------------------------------------------------------------------
		// Extract data from JSON
		// ------------------------------------------------------------------------
		extractCmdTpl := `cat %s \
			| jq -c '[.molecule.activity,.prediction.predictedLabels[%d].labels[]]' \
			| tr -d "[" | tr -d "]" | tr -d '"' \
			| awk -F, 'BEGIN {
					d["A"]["A","N"]=0;
					d["A"]["A",""]=0;
					d["A"]["N",""]=0;
					d["A"]["",""]=0;
					d["N"]["A","N"]=0;
					d["N"]["A",""]=0;
					d["N"]["N",""]=0;
					d["N"]["",""]=0;
				}
				{d[$1][$2,$3]++ }
				END {
					print "a_pred_both,a_pred_a,a_pred_n,a_pred_none,n_pred_both,n_pred_a,n_pred_n,n_pred_none";
					a_pred_both=d["A"]["A","N"];
					a_pred_a=d["A"]["A",""];
					a_pred_n=d["A"]["N",""];
					a_pred_none=d["A"]["",""];
					n_pred_both=d["N"]["A","N"];
					n_pred_a=d["N"]["A",""];
					n_pred_n=d["N"]["N",""];
					n_pred_none=d["N"]["",""];
					print a_pred_both "," a_pred_a "," a_pred_n "," a_pred_none "," n_pred_both "," n_pred_a "," n_pred_n "," n_pred_none;
				}' > {o:valstats}`

		valDataAll := wf.NewProc("extract_valdata_all_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson|join: }", confIdx))
		valDataAll.SetOut("valstats", "res/validation/valstats."+confLevel+".tbl.csv")
		valDataAll.In("valjson").From(sts.OutSubStream())

		valDataPerTarget := wf.NewProc("extract_valdata_pertarget_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson}", confIdx))
		valDataPerTarget.SetOutFunc("valstats", func(t *sp.Task) string {
			inFile := filepath.Base(t.InPath("valjson"))
			replacePtn, err := regexp.Compile(`\..*$`)
			sp.Check(err)
			gene := replacePtn.ReplaceAllString(inFile, "")
			return "res/validation/" + gene + "/" + gene + "." + confLevel + ".valstats.tbl.csv"
		})
		valDataPerTarget.In("valjson").From(validateFiles.Out())

		valDataPerTargetSTS := spc.NewStreamToSubStream(wf, "valdata_per_target_sts_"+confLevel)
		valDataPerTargetSTS.In().From(valDataPerTarget.Out("valstats"))

		mergeValDataPerTgt := wf.NewProc("merge_valdata_per_tgt_"+confLevel, `i=0; for f in {i:valdata_per_tgt|join: }; do let "i++"; if [[ $i == 1 ]]; then head -n 1 $f; fi; tail -n +2 $f; done > {o:merged}`)
		mergeValDataPerTgt.SetOut("merged", "res/validation/valstats."+confLevel+".tbl.alltargets.csv")
		mergeValDataPerTgt.In("valdata_per_tgt").From(valDataPerTargetSTS.OutSubStream())
	}

	wf.Run()
}
