package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	wf := sp.NewWorkflow("extract_valdata", 4)

	for confIdx, confLevel := range []string{"0p9", "0p8"} { // We use 'p' instead of '.' to avoid confusion in the file name
		confLevel := confLevel

		validateFiles := spc.NewFileGlobber(wf, "valstat_files_"+confLevel, "dat/validate/*/*1000.json")
		sts := spc.NewStreamToSubStream(wf, "sts_"+confLevel)
		sts.In().Connect(validateFiles.Out())

		// ------------------------------------------------------------------------
		// Grab input files
		// ------------------------------------------------------------------------

		// ------------------------------------------------------------------------
		// Extract data from JSON
		// ------------------------------------------------------------------------
		extractCmdTpl := `echo -e "orig_lab\tpred_none\tpred_a\tpred_n\tpred_both" > {o:valstats} \
		&& cat %s \
		| jq -c '[.molecule.activity,.prediction.predictedLabels[%d].labels[]]' \
		| tr -d "[" | tr -d "]" | tr -d '"' | \
		awk -F, '{
			d[$1][$2,$3]++ }
			END {
				print "A" "\t" d["A"]["",""] "\t" d["A"]["A",""] "\t" d["A"]["N",""] "\t" d["A"]["A","N"]
				print "N" "\t" d["N"]["",""] "\t" d["N"]["A",""] "\t" d["N"]["N",""] "\t" d["N"]["A","N"]
			}' >> {o:valstats}`

		valDataAll := wf.NewProc("extract_valdata_all_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson:r: }", confIdx))
		valDataAll.SetPathStatic("valstats", "res/validation/valstats."+confLevel+".tsv")
		valDataAll.In("valjson").Connect(sts.OutSubStream())

		valDataPerTarget := wf.NewProc("extract_valdata_pertarget_"+confLevel, fmt.Sprintf(extractCmdTpl, "{i:valjson}", confIdx))
		valDataPerTarget.SetPathCustom("valstats", func(t *sp.Task) string {
			inFile := filepath.Base(t.InPath("valjson"))
			replacePtn, err := regexp.Compile(`\..*$`)
			sp.Check(err)
			gene := replacePtn.ReplaceAllString(inFile, "")
			return "res/validation/" + gene + "/" + gene + "." + confLevel + ".valstats.tsv"
		})
		valDataPerTarget.In("valjson").Connect(validateFiles.Out())

		// ------------------------------------------------------------------------
		// Plot data
		// ------------------------------------------------------------------------
		extractGene := spc.NewMapToKeys(wf, "extract_gene_"+confLevel, func(ip *sp.FileIP) map[string]string {
			ptn, err := regexp.Compile(`\.0p.*\.tsv`)
			sp.Check(err)
			gene := strings.ToUpper(ptn.ReplaceAllString(filepath.Base(ip.Path()), ""))
			return map[string]string{"gene": gene}
		})
		extractGene.In().Connect(valDataPerTarget.Out("valstats"))

		plotValData := wf.NewProc("plot_valdata_"+confLevel, `Rscript bin/plot_valdata.r -i {i:valdata} -o {o:plot} -f pdf -g {k:valdata.gene}`)
		plotValData.SetPathExtend("valdata", "plot", ".pdf")
		plotValData.In("valdata").Connect(extractGene.Out())

		plotValDataAll := wf.NewProc("plot_valdata_all_"+confLevel, `Rscript bin/plot_valdata.r -i {i:valdata} -o {o:plot} -f pdf -g "all targets"`)
		plotValDataAll.SetPathExtend("valdata", "plot", ".pdf")
		plotValDataAll.In("valdata").Connect(valDataAll.Out("valstats"))
	}

	wf.Run()
}
