package main

import (
	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	wf := sp.NewWorkflow("extract_valdata", 4)

	validateFiles := spc.NewFileGlobber(wf, "valstat_files", "dat/validate/*/*1000.json")

	sts := spc.NewStreamToSubStream(wf, "sts")
	sts.In().Connect(validateFiles.Out())

	valDataAll := wf.NewProc("extract_valdata_all", getExtractCmd("{i:valjson:r: }"))
	valDataAll.SetPathStatic("valstats", "res/validation/valstats.tsv")
	valDataAll.In("valjson").Connect(sts.OutSubStream())

	valDataPerTarget := wf.NewProc("extract_valdata_pertarget", getExtractCmd("{i:valjson}"))
	valDataPerTarget.SetPathCustom("valstats", func(t *sp.Task) string {
		inFile := filepath.Base(t.InPath("valjson"))
		replacePtn, err := regexp.Compile(`\..*$`)
		sp.Check(err)
		gene := replacePtn.ReplaceAllString(inFile, "")
		return "res/validation/" + gene + "/" + gene + ".valstats.tsv"
	})
	valDataPerTarget.In("valjson").Connect(validateFiles.Out())

	extractGene := spc.NewMapToKeys(wf, "extract_gene", func(ip *sp.FileIP) map[string]string {
		ptn, err := regexp.Compile(`\.valstats\.tsv`)
		sp.Check(err)
		gene := strings.ToUpper(ptn.ReplaceAllString(filepath.Base(ip.Path()), ""))
		return map[string]string{"gene": gene}
	})
	extractGene.In().Connect(valDataPerTarget.Out("valstats"))

	plotValData := wf.NewProc("plot_valdata", `Rscript bin/plot_valdata.r -i {i:valdata} -o {o:plot} -f pdf -g {k:valdata.gene}`)
	plotValData.SetPathExtend("valdata", "plot", ".pdf")
	plotValData.In("valdata").Connect(extractGene.Out())

	plotValDataAll := wf.NewProc("plot_valdata_all", `Rscript bin/plot_valdata.r -i {i:valdata} -o {o:plot} -f pdf -g "all targets"`)
	plotValDataAll.SetPathExtend("valdata", "plot", ".pdf")
	plotValDataAll.In("valdata").Connect(valDataAll.Out("valstats"))

	wf.Run()
}

func getExtractCmd(infilePtn string) string {
	cmd := `echo -e "orig_lab\tpred_none\tpred_a\tpred_n\tpred_both" > {o:valstats} \
	&& cat ` + infilePtn + ` \
	| jq -c '[.molecule.activity,.prediction.predictedLabels[0].labels[]]' \
    | tr -d "[" | tr -d "]" | tr -d '"' | \
    awk -F, '{ 
        d[$1][$2,$3]++ } 
        END { 
            print "A" "\t" d["A"]["",""] "\t" d["A"]["A",""] "\t" d["A"]["N",""] "\t" d["A"]["A","N"]
			print "N" "\t" d["N"]["",""] "\t" d["N"]["A",""] "\t" d["N"]["N",""] "\t" d["N"]["A","N"]
		}' >> {o:valstats}`
	return cmd
}
