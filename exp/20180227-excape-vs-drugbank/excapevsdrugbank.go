package main

import (
	"encoding/csv"
	"encoding/xml"
	"os"

	sp "github.com/scipipe/scipipe"
)

func main() {
	wf := sp.NewWorkflow("exvsdb", 1)
	//wf.NewProc("dl", "curl -Lfv -o filename.zip -u ... https://www.drugbank.ca/releases/5-0-11/downloads/all-full-database")

	excapeDB := sp.NewFileIPGenerator(wf, "excapedb", "../../raw/pubchem.chembl.dataset4publication_inchi_smiles.tsv")
	drugBank := sp.NewFileIPGenerator(wf, "drugbank", "dat/drugbank.xml")

	extractIA := wf.NewProc("extract_inchikey_activity", `awk -F"\t" '{ print $1 "\t" $4 }' {i:tsv} > {o:tsv}`)
	extractIA.SetPathStatic("tsv", "dat/excapedb_inchikey_activity.tsv")
	extractIA.In("tsv").Connect(excapeDB.Out())

	xmlToTSV := wf.NewProc("xml_to_tsv", "# {i:xml} {o:tsv}")
	xmlToTSV.SetPathExtend("xml", "tsv", ".tsv")
	xmlToTSV.In("xml").Connect(drugBank.Out())
	xmlToTSV.CustomExecute = func(t *sp.Task) {
		fh, err := os.Open(t.InPath("xml"))
		defer fh.Close()
		if err != nil {
			sp.Fail("Could not open file", t.InPath("xml"))
		}

		tsvWrt := csv.NewWriter(t.OutTargets["tsv"].OpenWriteTemp())
		tsvWrt.Comma = '\t'
		tsvHeader := []string{"status", "inchikey"}
		tsvWrt.Write(tsvHeader)

		// Implement a streaming XML parser according to guide in
		// http://blog.davidsingleton.org/parsing-huge-xml-files-with-go
		xmlDec := xml.NewDecoder(fh)
		for {
			t, tokenErr := xmlDec.Token()
			if t == nil {
				break
			}
			if tokenErr != nil {
				sp.Fail("Failed to read token:", tokenErr)
			}
			switch startElem := t.(type) {
			case xml.StartElement:
				if startElem.Name.Local == "drug" {
					var status string
					var inchikey string

					drug := &Drug{}
					decErr := xmlDec.DecodeElement(drug, &startElem)
					if err != nil {
						sp.Fail("Could not decode element", decErr)
					}
					for _, g := range drug.Groups {
						if g == "approved" {
							status = "approved"
						}
						// Withdrawn till "shadow" (what's the correct term?) approved status
						if g == "withdrawn" {
							status = "withdrawn"
						}
					}
					for _, p := range drug.CalculatedProperties {
						if p.Kind == "InChIKey" {
							inchikey = p.Value
						}
					}

					if status != "" && inchikey != "" {
						tsvWrt.Write([]string{status, inchikey})
					}
				}
			case xml.EndElement:
				break
			}
		}
	}

	wf.Run()
}

type Drugbank struct {
	XMLName xml.Name `xml:"drugbank"`
	Drugs   []Drug   `xml:"drug"`
}

type Drug struct {
	XMLName              xml.Name   `xml:"drug"`
	Name                 string     `xml:"name"`
	Groups               []string   `xml:"groups>group"`
	CalculatedProperties []Property `xml:"calculated-properties>property"`
}

type Property struct {
	XMLName xml.Name `xml:"property"`
	Kind    string   `xml:"kind"`
	Value   string   `xml:"value"`
	Source  string   `xml:"source"`
}
