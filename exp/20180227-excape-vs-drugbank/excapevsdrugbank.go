package main

import (
	sp "github.com/scipipe/scipipe"
)

func main() {
	wf := sp.NewWorkflow("exvsdb", 1)
	wf.NewProc("dl", "curl -Lfv -o filename.zip -u ... https://www.drugbank.ca/releases/5-0-11/downloads/all-full-database")

	//apprZip := wf.NewProc("download_approved", "curl -Lf -o {o:zipfile} -u ... https://www.drugbank.ca/releases/5-0-11/downloads/approved-drug-sequences")
	//apprZip.SetPathStatic("zipfile", "drug_sequences_approved.zip")

	//withdrZip := wf.NewProc("download_withdrawn", "curl -Lf -o {o:zipfile} -u ... https://www.drugbank.ca/releases/5-0-11/downloads/withdrawn-drug-sequences")
	//withdrZip.SetPathStatic("zipfile", "drug_sequences_withdrawn.zip")

	//unpack := wf.NewProc("unzip", "unzip {i:zipfile}")
	//unpack.In("zipfile").Connect(apprZip.Out("zipfile"))
	//unpack.In("zipfile").Connect(withdrZip.Out("zipfile"))

	wf.Run()
}

type Drugbank struct {
}
