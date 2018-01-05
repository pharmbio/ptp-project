#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p core
#SBATCH -n 4
#SBATCH -J ptp_extract
#SBATCH -t 2-00:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type BEGIN,FAIL,END
module load java/sun_jdk1.8.0_92
module load R/3.4.0
go run fillup_workflow.go components.go -threads 2 -maxtasks 4 -geneset smallest1 # Run as little as possible ... just get the extract, and filtering processes to run
