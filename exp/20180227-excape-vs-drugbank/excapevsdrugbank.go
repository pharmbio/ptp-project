package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"

	sp "github.com/scipipe/scipipe"
)

// --------------------------------------------------------------------------------
// Main Workflow
// --------------------------------------------------------------------------------

func main() {
	wf := sp.NewWorkflow("exvsdb", 2)

	// DrugBank XML
	dlDrugBank := wf.NewProc("dl", "curl -Lfv -o {o:zip} -u $(cat drugbank_userinfo.txt) https://www.drugbank.ca/releases/5-0-11/downloads/all-full-database")
	dlDrugBank.SetPathStatic("zip", "dat/drugbank.zip")

	unzipDrugBank := wf.NewProc("unzip_drugbank", `unzip -d dat/ {i:zip}; mv "dat/full database.xml" {o:xml}`)
	unzipDrugBank.SetPathStatic("xml", "dat/drugbank.xml")
	unzipDrugBank.In("zip").Connect(dlDrugBank.Out("zip"))

	// Approved
	dlApproved := wf.NewProc("dl_approv", "curl -Lfv -o {o:zip} -u $(cat drugbank_userinfo.txt) https://www.drugbank.ca/releases/5-0-11/downloads/approved-structure-links")
	dlApproved.SetPathStatic("zip", "dat/drugbank_approved_csv.zip")

	unzipApproved := wf.NewProc("unzip_approved", `unzip -d dat/approved/ {i:zip}; mv "dat/approved/structure links.csv" {o:csv}`)
	unzipApproved.SetPathStatic("csv", "dat/drugbank_approved.csv")
	unzipApproved.In("zip").Connect(dlApproved.Out("zip"))

	// Withdrawn
	dlWithdrawn := wf.NewProc("dl_withdrawn", "curl -Lfv -o {o:zip} -u $(cat drugbank_userinfo.txt) https://www.drugbank.ca/releases/5-0-11/downloads/withdrawn-structure-links")
	dlWithdrawn.SetPathStatic("zip", "dat/drugbank_withdrawn_csv.zip")

	unzipWithdrawn := wf.NewProc("unzip_withdrawn", `unzip -d dat/withdrawn/ {i:zip}; mv "dat/withdrawn/structure links.csv" {o:csv}`)
	unzipWithdrawn.SetPathStatic("csv", "dat/drugbank_withdrawn.csv")
	unzipWithdrawn.In("zip").Connect(dlWithdrawn.Out("zip"))

	// Compound IDs
	drugBankCompIDCmd := "csvcut -K 1 -c 11,14 {i:drugbankcsv} | sed '/^,$/d' | sort -V > {o:compids}"

	drugBankCompIDsApprov := wf.NewProc("drugbank_compids_appr", drugBankCompIDCmd)
	drugBankCompIDsApprov.SetPathExtend("drugbankcsv", "compids", ".compids.csv")
	drugBankCompIDsApprov.In("drugbankcsv").Connect(unzipApproved.Out("csv"))

	drugBankCompIDsWithdr := wf.NewProc("drugbank_compids_withdr", drugBankCompIDCmd)
	drugBankCompIDsWithdr.SetPathExtend("drugbankcsv", "compids", ".compids.csv")
	drugBankCompIDsWithdr.In("drugbankcsv").Connect(unzipWithdrawn.Out("csv"))

	mergeApprWithdr := wf.NewProc("merge_appr_withdr", "cat {i:in1} {i:in2} | sort -V | uniq > {o:out}")
	mergeApprWithdr.SetPathStatic("out", "dat/drugbank_compids_all.csv")
	mergeApprWithdr.In("in1").Connect(drugBankCompIDsApprov.Out("compids"))
	mergeApprWithdr.In("in2").Connect(drugBankCompIDsWithdr.Out("compids"))

	// ExcapeDB
	excapeDB := sp.NewFileIPGenerator(wf, "excapedb", "../../raw/pubchem.chembl.dataset4publication_inchi_smiles.tsv")

	extractIA := wf.NewProc("extract_inchikey_activity", `awk -F"\t" '{ print $1 "\t" $4 }' {i:tsv} > {o:tsv}`)
	extractIA.SetPathStatic("tsv", "dat/excapedb_inchikey_activity.tsv")
	extractIA.In("tsv").Connect(excapeDB.Out())

	excapeDBOrigIDsAll := wf.NewProc("excapedb_origids_all", `tail -n +2 {i:excapedb} | awk -F'\t' '{ print $1 "\t" $2 }' | awk -F'\t' '{ print $2 }' | sort -V > {o:entries}`)
	excapeDBOrigIDsAll.SetPathExtend("excapedb", "entries", ".origids.all.tsv")
	excapeDBOrigIDsAll.In("excapedb").Connect(excapeDB.Out())

	excapeDBOrigIDsUnique := wf.NewProc("excapedb_origids_unique", `tail -n +2 {i:excapedb} | awk -F'\t' '{ print $1 "\t" $2 }' | uniq -w 23 | awk -F'\t' '{ print $2 }' | sort -V > {o:entries}`)
	excapeDBOrigIDsUnique.SetPathExtend("excapedb", "entries", ".origids.uniq.tsv")
	excapeDBOrigIDsUnique.In("excapedb").Connect(excapeDB.Out())

	//xmlToTSV := wf.NewProc("xml_to_tsv", "# Custom Go code with input: {i:xml} and output: {o:tsv}")
	//xmlToTSV.SetPathExtend("xml", "tsv", ".extr.tsv")
	//xmlToTSV.In("xml").Connect(unzipDrugBank.Out("xml"))
	//xmlToTSV.CustomExecute = NewXMLToTSVFunc()

	//sortTsv := wf.NewProc("sort_tsv", "head -n 1 {i:unsorted} > {o:sorted}; tail -n +2 {i:unsorted} | sort >> {o:sorted}")
	//sortTsv.SetPathExtend("unsorted", "sorted", ".sorted.tsv")
	//sortTsv.In("unsorted").Connect(xmlToTSV.Out("tsv"))

	excapeDBVsDrugBank := wf.NewProc("exc_vs_drb", "# Custom Go function with inputs: {i:excapedb_ids_uniq}, {i:excapedb_ids_all} {i:approv_ids}, {i:withdr_ids} and output: {o:stats}")
	excapeDBVsDrugBank.SetPathStatic("stats", "dat/excapedb_vs_drugbank_stats.json")
	excapeDBVsDrugBank.In("excapedb_ids_uniq").Connect(excapeDBOrigIDsUnique.Out("entries"))
	excapeDBVsDrugBank.In("excapedb_ids_all").Connect(excapeDBOrigIDsAll.Out("entries"))
	excapeDBVsDrugBank.In("approv_ids").Connect(drugBankCompIDsApprov.Out("compids"))
	excapeDBVsDrugBank.In("withdr_ids").Connect(drugBankCompIDsWithdr.Out("compids"))
	excapeDBVsDrugBank.CustomExecute = NewExcapeDBVsDrugBankFunc()

	wf.Run()
}

// --------------------------------------------------------------------------------
// Components and stuff
// --------------------------------------------------------------------------------

// NewExcapeDBVsDrugBankFunc returns a func to be used in the excapeDBVsDrugBank
// process in the workflow above
func NewExcapeDBVsDrugBankFunc() func(t *sp.Task) {
	return func(t *sp.Task) {
		approvIds := map[string]bool{}
		approvFile := t.InIP("approv_ids").Open()
		approvCsvReader := csv.NewReader(approvFile)
		for {
			rec, err := approvCsvReader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				sp.Fail(err)
			}
			approvIds[rec[0]] = true
			approvIds[rec[1]] = true
		}
		approvFile.Close()

		withdrIds := map[string]bool{}
		withdrFile := t.InIP("withdr_ids").Open()
		withdrCsvReader := csv.NewReader(withdrFile)
		for {
			rec, err := withdrCsvReader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				sp.Fail(err)
			}
			withdrIds[rec[0]] = true
			withdrIds[rec[1]] = true
		}
		withdrFile.Close()

		excIdsUniqFile := t.InIP("excapedb_ids_uniq").Open()
		excIdsUniqCsvReader := csv.NewReader(excIdsUniqFile)
		excIdsUniqCnt := 0
		approvUniqCnt := 0
		withdrUniqCnt := 0
		for {
			rec, err := excIdsUniqCsvReader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				sp.Fail(err)
			}
			excIdsUniqCnt++
			if _, ok := approvIds[rec[0]]; ok {
				approvUniqCnt++
			}
			if _, ok := withdrIds[rec[0]]; ok {
				withdrUniqCnt++
			}
		}
		excIdsUniqFile.Close()

		compoundsCnt := map[string]int{}
		compoundsCnt["excapedb_compounds_in_drugbank_approved"] = approvUniqCnt
		compoundsCnt["excapedb_compounds_in_drugbank_withdrawn"] = withdrUniqCnt
		compoundsCnt["excapedb_compounds_in_drugbank_total"] = approvUniqCnt + withdrUniqCnt
		compoundsCnt["excapedb_compounds_total"] = excIdsUniqCnt
		compCntJSON, cerr := json.Marshal(compoundsCnt)
		if cerr != nil {
			sp.Fail(cerr)
		}

		compoundsFrac := map[string]float64{}
		compoundsFrac["excapedb_fraction_compounds_in_drugbank_approved"] = float64(approvUniqCnt) / float64(excIdsUniqCnt)
		compoundsFrac["excapedb_fraction_compounds_in_drugbank_withdrawn"] = float64(withdrUniqCnt) / float64(excIdsUniqCnt)
		compoundsFrac["excapedb_fraction_compounds_in_drugbank_total"] = float64(approvUniqCnt+withdrUniqCnt) / float64(excIdsUniqCnt)
		compFracJSON, ferr := json.Marshal(compoundsFrac)
		if ferr != nil {
			sp.Fail(ferr)
		}

		excIdsAllFile := t.InIP("excapedb_ids_all").Open()
		excIdsAllCsvReader := csv.NewReader(excIdsAllFile)
		excIdsAllCnt := 0
		approvAllCnt := 0
		withdrAllCnt := 0
		for {
			rec, err := excIdsAllCsvReader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				sp.Fail(err)
			}
			excIdsAllCnt++
			if _, ok := approvIds[rec[0]]; ok {
				approvAllCnt++
			}
			if _, ok := withdrIds[rec[0]]; ok {
				withdrAllCnt++
			}
		}
		excIdsAllFile.Close()

		entriesCnt := map[string]int{}
		entriesCnt["excapedb_entries_in_drugbank_approved"] = approvAllCnt
		entriesCnt["excapedb_entries_in_drugbank_withdrawn"] = withdrAllCnt
		entriesCnt["excapedb_entries_in_drugbank_total"] = approvAllCnt + withdrAllCnt
		entriesCnt["excapedb_entries_total"] = excIdsAllCnt
		entrCntJSON, cerr := json.Marshal(entriesCnt)
		if cerr != nil {
			sp.Fail(cerr)
		}

		entriesFrac := map[string]float64{}
		entriesFrac["excapedb_fraction_entries_in_drugbank_approved"] = float64(approvAllCnt) / float64(excIdsAllCnt)
		entriesFrac["excapedb_fraction_entries_in_drugbank_withdrawn"] = float64(withdrAllCnt) / float64(excIdsAllCnt)
		entriesFrac["excapedb_fraction_entries_in_drugbank_total"] = float64(approvAllCnt+withdrAllCnt) / float64(excIdsAllCnt)
		entrFracJSON, ferr := json.Marshal(entriesFrac)
		if ferr != nil {
			sp.Fail(ferr)
		}

		t.OutIP("stats").Write(append(append(append(compCntJSON, compFracJSON...), entrCntJSON...), entrFracJSON...))
	}
}

// NewXMLToTSVFunc returns a CustomExecute function to be used by the XML to TSV
// component in the workflow above
func NewXMLToTSVFunc() func(t *sp.Task) {
	return func(t *sp.Task) {
		fh, err := os.Open(t.InPath("xml"))
		if err != nil {
			sp.Fail("Could not open file", t.InPath("xml"))
		}

		tsvWrt := csv.NewWriter(t.OutIP("tsv").OpenWriteTemp())
		tsvWrt.Comma = '\t'
		tsvHeader := []string{"inchikey", "status", "chembl_id", "pubchem_sid", "pubchem_cid"}
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
					var pubchemSID string
					var pubchemCID string

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
							status = "W"
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
							pubchemSID = eid.Identifier
						} else if eid.Resource == "PubChem Compound" {
							pubchemCID = eid.Identifier
						}
					}

					tsvWrt.Write([]string{inchiKey, status, chemblID, pubchemSID, pubchemCID})
				}
			case xml.EndElement:
				continue
			}
		}
		tsvWrt.Flush()
		fh.Close()
	}
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
