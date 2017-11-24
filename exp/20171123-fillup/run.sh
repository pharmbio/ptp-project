#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p node
#SBATCH -C mem256GB
#SBATCH -N 1
#SBATCH -J ptp_fillup_wf
#SBATCH -t 7-00:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type BEGIN,FAIL,END
module load java/sun_jdk1.8.0_92
module load R/3.4.0
go run fillup_workflow.go components.go -threads 2 -maxtasks 20 -geneset bowes44min100percls
