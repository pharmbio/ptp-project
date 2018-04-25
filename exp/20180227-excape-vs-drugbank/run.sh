#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p core
#SBATCH -n 2
#SBATCH -J ExcapeDBvsDrugBank
#SBATCH -t 8:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type BEGIN,FAIL,END
module load java/sun_jdk1.8.0_92
module load R/3.4.0
go run excapedbvsdrugbank.go
