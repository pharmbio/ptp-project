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

	run := sp.NewPipelineRunner()

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------

	excapeDBFileName := "dat/pubchem.chembl.dataset4publication_inchi_smiles.tsv.xz"
	dlExcapeDB := sp.NewFromShell("dlDB", fmt.Sprintf("wget https://zenodo.org/record/173258/files/%s -O {o:excapexz}", excapeDBFileName))
	dlExcapeDB.SetPathStatic("excapexz", excapeDBFileName)
	run.AddProcess(dlExcapeDB)

	unPackDB := sp.NewFromShell("unPackDB", "xzcat {i:xzfile} > {o:unxzed}")
	unPackDB.SetPathReplace("xzfile", "unxzed", ".xz", "")
	run.AddProcess(unPackDB)

	sink := sp.NewSink()
	run.AddProcess(sink)

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------
	unPackDB.In["xzfile"].Connect(dlExcapeDB.Out["excapexz"])
	sink.Connect(unPackDB.Out["unxzed"])

	// --------------------------------
	// Run the pipeline!
	// --------------------------------

	run.Run()
}
