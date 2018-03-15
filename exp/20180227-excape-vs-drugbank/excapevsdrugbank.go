package main

import (
	"encoding/csv"
	"encoding/xml"
	"os"

	sp "github.com/scipipe/scipipe"
)

func main() {
	wf := sp.NewWorkflow("exvsdb", 2)
	dlDrugBank := wf.NewProc("dl", "curl -Lfv -o {o:zipfile} -u $(cat drugbank_userinfo.txt) https://www.drugbank.ca/releases/5-0-11/downloads/all-full-database")
	dlDrugBank.SetPathStatic("zipfile", "dat/drugbank.zip")

	unzipDrugBank := wf.NewProc("unzip_drugbank", `unzip -d dat/ {i:zipfile}; mv "dat/full database.xml" {o:drugbankxml}`)
	unzipDrugBank.SetPathStatic("drugbankxml", "dat/drugbank.xml")
	unzipDrugBank.In("zipfile").Connect(dlDrugBank.Out("zipfile"))

	excapeDB := sp.NewFileIPGenerator(wf, "excapedb", "../../raw/pubchem.chembl.dataset4publication_inchi_smiles.tsv")

	extractIA := wf.NewProc("extract_inchikey_activity", `awk -F"\t" '{ print $1 "\t" $4 }' {i:tsv} > {o:tsv}`)
	extractIA.SetPathStatic("tsv", "dat/excapedb_inchikey_activity.tsv")
	extractIA.In("tsv").Connect(excapeDB.Out())

	xmlToTSV := wf.NewProc("xml_to_tsv", "# Custom Go code with input: {i:xml} and output: {o:tsv}")
	xmlToTSV.SetPathExtend("xml", "tsv", ".extr.tsv")
	xmlToTSV.In("xml").Connect(unzipDrugBank.Out("drugbankxml"))
	xmlToTSV.CustomExecute = func(t *sp.Task) {
		fh, err := os.Open(t.InPath("xml"))
		if err != nil {
			sp.Fail("Could not open file", t.InPath("xml"))
		}

		tsvWrt := csv.NewWriter(t.OutTargets["tsv"].OpenWriteTemp())
		tsvWrt.Comma = '\t'
		tsvHeader := []string{"inchikey", "status", "chembl_id", "pubchem_substance_id", "pubchem_compound_id"}
		tsvWrt.Write(tsvHeader)

		// Implement a streaming XML parser according to guide in
		// http://blog.davidsingleton.org/parsing-huge-xml-files-with-go
		xmlDec := xml.NewDecoder(fh)
		for {
			t, tokenErr := xmlDec.Token()
			if tokenErr != nil {
				if tokenErr.Error() == "EOF" {
					break
				} else {
					sp.Fail("Failed to read token:", tokenErr)
				}
			}
			switch startElem := t.(type) {
			case xml.StartElement:
				if startElem.Name.Local == "drug" {
					var status string
					var inchiKey string
					var chemblID string
					var pubChemSubstanceID string
					var pubChemCompoundID string

					drug := &Drug{}
					decErr := xmlDec.DecodeElement(drug, &startElem)
					if err != nil {
						sp.Fail("Could not decode element", decErr)
					}
					for _, g := range drug.Groups {
						if g == "approved" {
							status = "A"
						}
						// Withdrawn till "shadow" (what's the correct term?) approved status
						if g == "withdrawn" {
							status = "N"
						}
					}
					for _, p := range drug.CalculatedProperties {
						if p.Kind == "InChIKey" {
							inchiKey = p.Value
						}
					}

					for _, eid := range drug.ExternalIdentifiers {
						if eid.Resource == "ChEMBL" {
							chemblID = eid.Identifier
						} else if eid.Resource == "PubChem Substance" {
							pubChemSubstanceID = eid.Identifier
						} else if eid.Resource == "PubChem Compound" {
							pubChemCompoundID = eid.Identifier
						}
					}

					if status != "" && inchiKey != "" {
						tsvWrt.Write([]string{inchiKey, status, chemblID, pubChemSubstanceID, pubChemCompoundID})
					}
				}
			case xml.EndElement:
				continue
			}
		}
		tsvWrt.Flush()
		fh.Close()
	}

	sortTsv := wf.NewProc("sort_tsv", "head -n 1 {i:unsorted} > {o:sorted}; tail -n +2 {i:unsorted} | sort >> {o:sorted}")
	sortTsv.SetPathExtend("unsorted", "sorted", ".sorted.tsv")
	sortTsv.In("unsorted").Connect(xmlToTSV.Out("tsv"))

	wf.Run()
}

type Drugbank struct {
	XMLName xml.Name `xml:"drugbank"`
	Drugs   []Drug   `xml:"drug"`
}

type Drug struct {
	XMLName              xml.Name             `xml:"drug"`
	Name                 string               `xml:"name"`
	Groups               []string             `xml:"groups>group"`
	CalculatedProperties []Property           `xml:"calculated-properties>property"`
	ExternalIdentifiers  []ExternalIdentifier `xml:"external-identifiers>external-identifier"`
}

type Property struct {
	XMLName xml.Name `xml:"property"`
	Kind    string   `xml:"kind"`
	Value   string   `xml:"value"`
	Source  string   `xml:"source"`
}

type ExternalIdentifier struct {
	XMLName    xml.Name `xml:"external-identifier"`
	Resource   string   `xml:"resource"`
	Identifier string   `xml:"identifier"`
}
