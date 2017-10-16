#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p node
#SBATCH -N 1
#SBATCH -J Run_PTP_WF_in_SciPipe
#SBATCH -t 7-00:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type FAIL,END
module load java/sun_jdk1.8.0_92
go run train_models.go -maxcores 14 -geneset bowes44
