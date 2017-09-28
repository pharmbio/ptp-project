// Workflow written in SciPipe.
// For more information about SciPipe, see: http://scipipe.org
package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
)

const (
	bowesRiskGenes = []string{
		"ADORA2A",
		"ADRA1A",
		"ADRA2A",
		"ADRB1",
		"ADRB2",
		"CNR1",
		"CNR2",
		"CCKAR",
		"DRD1",
		"DRD2",
		"EDNRA",
		"HRH1",
		"HRH2",
		"OPRD1",
		"OPRK1",
		"OPRM1",
		"CHRM1",
		"CHRM2",
		"CHRM3",
		"HTR1A",
		"HTR1B",
		"HTR2A",
		"HTR2B",
		"AVPR1A",
		"CHRNA1",
		"CHRNA4",
		"CACNA1C",
		"GABRA1",
		"KCNH2",
		"KCNQ1",
		"KCNE1",
		"GRIN1",
		"HTR3A",
		"SCN5A",
		"ACHE",
		"PTGS1",
		"PTGS2",
		"MAOA",
		"PDE3A",
		"PDE4D",
		"LCK",
		"SLC6A3",
		"SLC6A2",
		"SLC6A4",
		"AR",
		"NR3C1",
	}
)

func main() {
	// --------------------------------
	// Create a pipeline runner
	// --------------------------------

	wf := sp.NewWorkflow("explore_excapedb")

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------

	dbFileName := "pubchem.chembl.dataset4publication_inchi_smiles.tsv.xz"
	dlExcapeDB := wf.NewProc("dlDB", fmt.Sprintf("wget https://zenodo.org/record/173258/files/%s -O {o:excapexz}", dbFileName))
	dlExcapeDB.SetPathStatic("excapexz", "dat/"+dbFileName)

	unPackDB := wf.NewProc("unPackDB", "xzcat {i:xzfile} > {o:unxzed}")
	unPackDB.SetPathReplace("xzfile", "unxzed", ".xz", "")
	// Slurm string
	//unPackDB.Prepend = "salloc -A snic2017-7-89 -n 2 -t 8:00:00 -J unpack_excapedb"

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------
	unPackDB.In("xzfile").Connect(dlExcapeDB.Out("excapexz"))

	// --------------------------------
	// Count ligands in targets
	// --------------------------------
	// for _, gene := range bowesRiskGenes {

	// }

	wf.ConnectLast(unPackDB.Out("unxzed"))

	// --------------------------------
	// Run the pipeline!
	// --------------------------------

	wf.Run()
}
