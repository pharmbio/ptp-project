#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p node
#SBATCH -N 1
#SBATCH -J Run_PTPWF_on_one_node
#SBATCH -t 4-00:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type BEGIN,FAIL,END
module load java/sun_jdk1.8.0_92
go run train_models.go components.go -threads 2 -maxtasks 10 -geneset bowes44
