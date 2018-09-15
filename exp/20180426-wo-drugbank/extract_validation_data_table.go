package main

import (
	"fmt"
	"path/filepath"
	"regexp"

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
		extractCmdTpl := `cat %s \
			| jq -c '[.molecule.activity,.prediction.predictedLabels[%d].labels[]]' \
			| tr -d "[" | tr -d "]" | tr -d '"' \
			| awk -F, '{d[$1][$2,$3]++ }
				END {
					print "a_pred_both\ta_pred_a\ta_pred_n\ta_pred_none\tn_pred_both\tn_pred_a\tn_pred_n\tn_pred_none\tsens_excl\tsens_incl\tspec_excl\tspec_incl";
					a_pred_both=d["A"]["A","N"];
					a_pred_a=d["A"]["A",""];
					a_pred_n=d["A"]["N",""];
					a_pred_none=d["A"]["",""];
					n_pred_both=d["N"]["A","N"];
					n_pred_a=d["N"]["A",""];
					n_pred_n=d["N"]["N",""];
					n_pred_none=d["N"]["",""];
					if (a_pred_a > 0) {
						sens_excl=a_pred_a/(a_pred_a + a_pred_n);
						sens_incl=a_pred_a/(a_pred_a + a_pred_n + a_pred_both + a_pred_none);
					};
					if (n_pred_n > 0) {
						spec_excl=n_pred_n/(n_pred_n + n_pred_a);
						spec_incl=n_pred_n/(n_pred_n + n_pred_a + n_pred_both + n_pred_none);
					};
					print a_pred_both "\t" a_pred_a "\t" a_pred_n "\t" a_pred_none "\t" n_pred_both "\t" n_pred_a "\t" n_pred_n "\t" n_pred_none "\t" sens_excl "\t" sens_incl "\t" spec_excl "\t" spec_incl;
				}' > {o:valstats}`

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
	}

	wf.Run()
}
