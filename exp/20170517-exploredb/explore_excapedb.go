// Workflow written in SciPipe.
// For more information about SciPipe, see: http://scipipe.org
package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
	"strings"
)

var (
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
	// SLURM string
	//unPackDB.Prepend = "salloc -A snic2017-7-89 -n 2 -t 8:00:00 -J unpack_excapedb"

	// --------------------------------
	// Connect workflow dependency network
	// --------------------------------
	unPackDB.In("xzfile").Connect(dlExcapeDB.Out("excapexz"))

	unPackDBFanOut := spc.NewFanOut("unpackdb_fanout")
	unPackDBFanOut.InFile.Connect(unPackDB.Out("unxzed"))
	wf.Add(unPackDBFanOut) // Oh, this is so easy to forget!!!

	// --------------------------------
	// Count ligands in targets
	// --------------------------------
	for _, gene := range bowesRiskGenes {
		geneLC := strings.ToLower(gene)
		procName := "cnt_comp_" + geneLC

		countCompoundsPerTarget := wf.NewProc(procName, "awk '$9 == \""+gene+"\" { SUM += 1 } END { print SUM }' {i:tsvfile} > {o:compound_count}")
		countCompoundsPerTarget.SetPathStatic("compound_count", "dat/compound_count_"+geneLC+".txt")
		countCompoundsPerTarget.In("tsvfile").Connect(unPackDBFanOut.Out("to_" + procName))
		// SLURM string
		countCompoundsPerTarget.Prepend = "salloc -A snic2017-7-89 -n 2 -t 1:00:00 -J scipipe_cnt_comp_" + geneLC + " srun "
		wf.ConnectLast(countCompoundsPerTarget.Out("compound_count"))
	}

	// --------------------------------
	// Run the pipeline!
	// --------------------------------
	wf.Run()
}
