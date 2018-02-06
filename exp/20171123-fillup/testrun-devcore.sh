#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p devcore
#SBATCH -n 2
#SBATCH -J Test_PTPWF_on_two_cores
#SBATCH -t 1:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type START,FAIL,END
module load java/sun_jdk1.8.0_92
module load R/3.4.0
go run fillup_workflow.go components.go -threads 1 -maxtasks 2 -geneset smallest3 # -debug
