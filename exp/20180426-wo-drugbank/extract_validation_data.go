package main

import (
	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
)

func main() {
	wf := sp.NewWorkflow("extract_valdata", 1)

	validateFiles := spc.NewFileGlobber(wf, "valstat_files", "dat/validate/*/*1000.json")

	sts := spc.NewStreamToSubStream(wf, "sts")
	sts.In().Connect(validateFiles.Out())

	extractValidationData := wf.NewProc("extract_validation_data", `cat {i:valjson:r: } \
    | jq -c '[.molecule.activity,.prediction.predictedLabels[0].labels[]]' \
    | tr -d "[" | tr -d "]" | tr -d '"' | \
    awk -F, '{ 
        d[$1][$2,$3]++ } 
        END { 
            print d["A"]["",""] "\t" d["A"]["A",""] "\t" d["A"]["N",""] "\t" d["A"]["A","N"]
            print d["N"]["",""] "\t" d["N"]["A",""] "\t" d["N"]["N",""] "\t" d["N"]["A","N"]
		}' > {o:valstats}`)
	extractValidationData.SetPathStatic("valstats", "valstats.tsv")
	extractValidationData.In("valjson").Connect(sts.OutSubStream())

	wf.Run()
}
