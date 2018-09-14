package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
)

func main() {
	wf := sp.NewWorkflow("extract_valdata", 4)

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
		extractCmdTpl := `echo -e "a_pred_both\ta_pred_a\ta_pred_n\ta_pred_none\tn_pred_both\tn_pred_a\tn_pred_n\tn_pred_none" > {o:valstats} \
			&& cat %s \
			| jq -c '[.molecule.activity,.prediction.predictedLabels[%d].labels[]]' \
			| tr -d "[" | tr -d "]" | tr -d '"' | \
			awk -F, '{
				d[$1][$2,$3]++ }
				END {
					print d["A"]["A","N"] "\t" d["A"]["A",""] "\t" d["A"]["N",""] "\t" d["A"]["",""] "\t" d["N"]["A","N"] "\t" d["N"]["A",""] "\t" d["N"]["N",""] "\t" d["N"]["",""]
				}' >> {o:valstats}`

		valDataAll := wf.NewProc("extract_valdata_all_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson|join: }", confIdx))
		valDataAll.SetOut("valstats", "res/validation/valstats."+confLevel+".tbl.tsv")
		valDataAll.In("valjson").From(sts.OutSubStream())

		valDataPerTarget := wf.NewProc("extract_valdata_pertarget_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson}", confIdx))
		valDataPerTarget.SetOutFunc("valstats", func(t *sp.Task) string {
			inFile := filepath.Base(t.InPath("valjson"))
			replacePtn, err := regexp.Compile(`\..*$`)
			sp.Check(err)
			gene := replacePtn.ReplaceAllString(inFile, "")
			return "res/validation/" + gene + "/" + gene + "." + confLevel + ".valstats.tbl.tsv"
		})
		valDataPerTarget.In("valjson").From(validateFiles.Out())

	wf.Run()
}
