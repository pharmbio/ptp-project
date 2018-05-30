package main

import (
	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
	"path/filepath"
	"regexp"
)

func main() {
	wf := sp.NewWorkflow("extract_valdata", 1)

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

	wf.Run()
}

func getExtractCmd(infilePtn string) string {
	cmd := `echo -e "A->none\tA->A\tA->N\tA->both\tN->none\tN->A\tN->N\tN->both" > {o:valstats} \
	&& cat ` + infilePtn + ` \
	| jq -c '[.molecule.activity,.prediction.predictedLabels[0].labels[]]' \
    | tr -d "[" | tr -d "]" | tr -d '"' | \
    awk -F, '{ 
        d[$1][$2,$3]++ } 
        END { 
            print d["A"]["",""] "\t" d["A"]["A",""] "\t" d["A"]["N",""] "\t" d["A"]["A","N"] "\t" d["N"]["",""] "\t" d["N"]["A",""] "\t" d["N"]["N",""] "\t" d["N"]["A","N"]
		}' >> {o:valstats}`
	return cmd
}
