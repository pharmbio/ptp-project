#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p devel
#SBATCH -N 1
#SBATCH -J Test_PTPWF_on_one_node
#SBATCH -t 1:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type FAIL,END
go run fillup_workflow.go components.go -threads 2 -maxtasks 4 -geneset smallest3 # -debug
