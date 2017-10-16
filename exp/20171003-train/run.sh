#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p core
#SBATCH -n 2
#SBATCH -J Run_PTP_WF_in_SciPipe
#SBATCH -t 7-00:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type FAIL,END
module load java/sun_jdk1.8.0_92
go run train_models.go -maxcores 500 -geneset bowes44 -slurm
