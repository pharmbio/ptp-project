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
		// Not available in dataset: "CHRNA1",
		"CHRNA4",
		"CACNA1C",
		"GABRA1",
		"KCNH2",
		"KCNQ1",
		// Not available in dataset: "KCNE1",
		"MINK1", // Used instead of KCNE1.
		"GRIN1", // They both share the same alias "MinK", and also confirmed by Wes to be the same.
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
	wf := sp.NewWorkflow("train_models")

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------
	dbFileName := "pubchem.chembl.dataset4publication_inchi_smiles.tsv.xz"
	dlExcapeDB := wf.NewProc("dlDB", fmt.Sprintf("wget https://zenodo.org/record/173258/files/%s -O {o:excapexz}", dbFileName))
	dlExcapeDB.SetPathStatic("excapexz", "../../raw/"+dbFileName)

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
	wf.AddProc(unPackDBFanOut) // Oh, this is so easy to forget!!!

	// --------------------------------
	// Count ligands in targets
	// --------------------------------
	for _, gene := range bowesRiskGenes {
		geneLC := strings.ToLower(gene)
		procName := "extract_target_data_" + geneLC

		extractTargetData := wf.NewProc(procName, fmt.Sprintf(`awk -F"\t" '$9 == "%s" { print $12"\t"$4 }' {i:raw_data} > {o:target_data}`, gene))
		extractTargetData.SetPathStatic("target_data", fmt.Sprintf("dat/%s/%s.tsv", geneLC, geneLC))
		extractTargetData.Prepend = "salloc -A snic2017-7-89 -n 4 -t 1:00:00 -J scipipe_cnt_comp_" + geneLC + " srun " // SLURM string
		extractTargetData.In("raw_data").Connect(unPackDBFanOut.Out("unxzed"))

		trainModel := wf.NewProc("train_model_"+geneLC,
			fmt.Sprintf(`cpsign-train --cptype 1 --train-file {i:target_data} -i liblinear --nr-models %d --model-name "Ligand binding to %s gene" --model-out {o:model}`,
				3,
				gene))
		trainModel.SetPathExtend("target_data", "model", ".cpsign")
		trainModel.Prepend = "salloc -A snic2017-7-89 -n 4 -t 1:00:00 -J cpsign_train_" + geneLC + " srun " // SLURM string
		trainModel.In("target_data").Connect(extractTargetData.Out("target_data"))

		wf.ConnectLast(trainModel.Out("model"))
	}

	// --------------------------------
	// Run the pipeline!
	// --------------------------------
	wf.Run()
}
