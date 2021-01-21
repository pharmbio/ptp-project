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

		validateFiles := spc.NewFileGlobber(wf, "valstat_files_"+confLevel, "dat/validate/*/*1000.json")
		sts := spc.NewStreamToSubStream(wf, "sts_"+confLevel)
		sts.In().From(validateFiles.Out())

		// ------------------------------------------------------------------------
		// Grab input files
		// ------------------------------------------------------------------------

		// ------------------------------------------------------------------------
		// Extract data from JSON
		// ------------------------------------------------------------------------
		extractCmdTpl := `echo -e "orig_lab\tpred_both\tpred_a\tpred_n\tpred_none" > {o:valstats} \
		&& cat %s \
		| jq -c '[.molecule.activity,.prediction.predictedLabels[%d].labels[]]' \
		| tr -d "[" | tr -d "]" | tr -d '"' | \
		awk -F, '{
			d[$1][$2,$3]++ }
			END {
				print "A" "\t" d["A"]["A","N"] "\t" d["A"]["A",""] "\t" d["A"]["N",""] "\t" d["A"]["",""]
				print "N" "\t" d["N"]["A","N"] "\t" d["N"]["A",""] "\t" d["N"]["N",""] "\t" d["N"]["",""]
			}' >> {o:valstats}`

		valDataAll := wf.NewProc("extract_valdata_all_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson|join: }", confIdx))
		valDataAll.SetOut("valstats", "res/validation/valstats."+confLevel+".tsv")
		valDataAll.In("valjson").From(sts.OutSubStream())

		valDataPerTarget := wf.NewProc("extract_valdata_pertarget_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson}", confIdx))
		valDataPerTarget.SetOutFunc("valstats", func(t *sp.Task) string {
			confLevel := confLevel
			inFile := filepath.Base(t.InPath("valjson"))
			replacePtn, err := regexp.Compile(`\..*$`)
			sp.Check(err)
			gene := replacePtn.ReplaceAllString(inFile, "")
			return "res/validation/" + gene + "/" + gene + "." + confLevel + ".valstats.tsv"
		})
		valDataPerTarget.In("valjson").From(validateFiles.Out())

		// ------------------------------------------------------------------------
		// Plot data
		// ------------------------------------------------------------------------
		extractGene := spc.NewMapToTags(wf, "extract_gene_"+confLevel, func(ip *sp.FileIP) map[string]string {
			ptn, err := regexp.Compile(`\.0p.*\.tsv`)
			sp.Check(err)
			gene := strings.ToUpper(ptn.ReplaceAllString(filepath.Base(ip.Path()), ""))
			return map[string]string{"gene": gene}
		})
		extractGene.In().From(valDataPerTarget.Out("valstats"))

		plotValData := wf.NewProc("plot_valdata_"+confLevel, `Rscript ../bin/plot_valdata.r -i {i:valdata} -o {o:plot} -f pdf -g {k:valdata.gene} -c `+strings.Replace(confLevel, "p", ".", 1))
		plotValData.SetOut("plot", "{i:valdata}.pdf")
		plotValData.In("valdata").From(extractGene.Out())

		plotValDataAll := wf.NewProc("plot_valdata_all_"+confLevel, `Rscript ../bin/plot_valdata.r -i {i:valdata} -o {o:plot} -f pdf -g "all targets" -c `+strings.Replace(confLevel, "p", ".", 1))
		plotValDataAll.SetOut("plot", "{i:valdata}.pdf")
		plotValDataAll.In("valdata").From(valDataAll.Out("valstats"))
	}

	wf.Run()
}
