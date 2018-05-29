package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	wf := sp.NewWorkflow("extract_valdata", 1)

	extractValidationData := wf.NewProc("extract_validation_data", `cat dat/validate/*/*1000.json \
    | jq -c '[.molecule.activity,.prediction.predictedLabels[0].labels[]]' \
    | tr -d "[" | tr -d "]" | tr -d '"' | \
    awk -F, '{ 
        d[$1][$2,$3]++ } 
        END { 
            print d["A"]["",""] "\t" d["A"]["A",""] "\t" d["A"]["N",""] "\t" d["A"]["A","N"]
            print d["N"]["",""] "\t" d["N"]["A",""] "\t" d["N"]["N",""] "\t" d["N"]["A","N"]
		}' > {o:valstats}`)
	extractValidationData.SetPathStatic("valstats", "validation_stats.tsv")

	wf.Run()
}
