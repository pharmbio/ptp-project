package main

import (
	"encoding/xml"
	"fmt"
	"os"

	sp "github.com/scipipe/scipipe"
)

func main() {
	wf := sp.NewWorkflow("exvsdb", 1)
	//wf.NewProc("dl", "curl -Lfv -o filename.zip -u ... https://www.drugbank.ca/releases/5-0-11/downloads/all-full-database")

	drugBank := sp.NewFileIPGenerator(wf, "drugbank_file", "dat/drugbank.xml")

	xmlParser := wf.NewProc("xmlparser", "# {i:xml}")
	xmlParser.In("xml").Connect(drugBank.Out())
	xmlParser.CustomExecute = func(t *sp.Task) {
		fh, err := os.Open(t.InPath("xml"))
		if err != nil {
			sp.Fail("Could not open file", t.InPath("xml"))
		}

		// Implement a streaming XML parser according to guide in
		// http://blog.davidsingleton.org/parsing-huge-xml-files-with-go
		xmlDec := xml.NewDecoder(fh)
		for {
			t, tokenErr := xmlDec.Token()
			if tokenErr != nil {
				sp.Fail("Failed to read token:", tokenErr)
			}
			if t == nil {
				break
			}
			switch startElem := t.(type) {
			case xml.StartElement:
				if startElem.Name.Local == "drug" {
					drug := &Drug{}
					decErr := xmlDec.DecodeElement(drug, &startElem)
					if err != nil {
						sp.Fail("Could not decode element", decErr)
					}
					fmt.Println("Drug:", drug.Name)
					for i, g := range drug.Groups {
						fmt.Printf("... Group %d: %s\n", i, g)
					}
				}
			}
		}
	}

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
	XMLName xml.Name `xml:"drugbank"`
	Drugs   []Drug   `xml:"drug"`
}

type Drug struct {
	XMLName xml.Name `xml:"drug"`
	Name    string   `xml:"name"`
	Groups  []string `xml:"groups>group"`
}

type Property struct {
	XMLName xml.Name `xml:"property"`
}

type Group struct {
	XMLName xml.Name `xml:"group"`
}
