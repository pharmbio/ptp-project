// Workflow written in SciPipe.
// For more information about SciPipe, see: http://scipipe.org
package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
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
	unPackDB.Prepend = "salloc -A snic2017-7-89 -n 2 -t 8:00:00 -J unpack_excapedb"

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------
	unPackDB.In("xzfile").Connect(dlExcapeDB.Out("excapexz"))
	wf.ConnectLast(unPackDB.Out("unxzed"))

	// --------------------------------
	// Run the pipeline!
	// --------------------------------

	wf.Run()
}
