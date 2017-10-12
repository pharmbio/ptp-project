// Workflow written in SciPipe.
// For more information about SciPipe, see: http://scipipe.org
package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
	str "strings"
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
	smallestGene = []string{
		"GABRA1",
	}
	smallestThreeGenes = []string{
		"GABRA1",
		"CACNA1C",
		"CHRNA4",
	}
	costVals = []string{
		"1",
		"10",
		"100",
	}
	gammaVals = []string{
		"0.1",
		"0.01",
		"0.001",
	}
)

func main() {
	//sp.InitLogDebug()
	wf := sp.NewWorkflow("train_models", 4)

	// --------------------------------
	// Initialize processes and add to runner
	// --------------------------------
	dbFileName := "pubchem.chembl.dataset4publication_inchi_smiles.tsv.xz"
	dlExcapeDB := wf.NewProc("dlDB", fmt.Sprintf("wget https://zenodo.org/record/173258/files/%s -O {o:excapexz}", dbFileName))
	dlExcapeDB.SetPathStatic("excapexz", "../../raw/"+dbFileName)

	unPackDB := wf.NewProc("unPackDB", "xzcat {i:xzfile} > {o:unxzed}")
	unPackDB.SetPathReplace("xzfile", "unxzed", ".xz", "")
	unPackDB.In("xzfile").Connect(dlExcapeDB.Out("excapexz"))
	//unPackDB.Prepend = "salloc -A snic2017-7-89 -n 2 -t 8:00:00 -J unpack_excapedb"

	// --------------------------------
	// Set up gene-specific workflow branches
	// --------------------------------
	//for _, gene := range bowesRiskGenes {
	for _, gene := range smallestThreeGenes {
		geneLC := str.ToLower(gene)
		procName := "extract_target_data_" + geneLC

		extractTargetData := wf.NewProc(procName, fmt.Sprintf(`awk -F"\t" '$9 == "%s" { print $12"\t"$4 }' {i:raw_data} > {o:target_data}`, gene))
		extractTargetData.SetPathStatic("target_data", fmt.Sprintf("dat/%s/%s.tsv", geneLC, geneLC))
		extractTargetData.In("raw_data").Connect(unPackDB.Out("unxzed"))
		//extractTargetData.Prepend = "salloc -A snic2017-7-89 -n 4 -t 1:00:00 -J scipipe_cnt_comp_" + geneLC + " srun " // SLURM string

		for _, cost := range costVals {
			for _, gamma := range gammaVals {
				gene_cost_gamma := fmt.Sprintf("%s_%s_%s", geneLC, cost, gamma) // A string to make process names unique

				trainModel := wf.NewProc("train_"+gene_cost_gamma,
					sp.ExpandParams(`java -jar ../../bin/cpsign-0.6.2.jar train --license ../../bin/cpsign.lic --cptype 1 --trainfile {i:target_data} -i liblinear -l A, N --cost {p:cost} --gamma {p:gamma} --nr-models {p:nrmodels} --model-name "Ligand_binding_to_{p:gene}_gene" --model-out {o:model}`,
						map[string]string{
							"nrmodels": "3",
							"gene":     gene,
						}))
				trainModel.SetPathCustom("model", func(t *sp.SciTask) string {
					return t.InPath("target_data") + fmt.Sprintf(".c%s_g%s", t.Param("cost"), t.Param("gamma")) + ".cpsign"
				})
				trainModel.In("target_data").Connect(extractTargetData.Out("target_data"))
				trainModel.ParamPort("cost").ConnectStrings(cost)
				trainModel.ParamPort("gamma").ConnectStrings(gamma)
				//trainModel.Prepend = "salloc -A snic2017-7-89 -n 4 -t 1:00:00 -J cpsign_train_" + geneLC + " srun " // SLURM string

				wf.ConnectLast(trainModel.Out("model"))
			}
		}

	}

	// --------------------------------
	// Run the pipeline!
	// --------------------------------
	wf.Run()
}
