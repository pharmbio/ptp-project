#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p node
#SBATCH -C mem256GB
#SBATCH -n 20
#SBATCH -J ptp_fullwf_wo_drugbank
#SBATCH -t 4:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type BEGIN,FAIL,END
module load java/sun_jdk1.8.0_92
module load R/3.4.0
go run wo_drugbank_wf.go components.go -threads 1 -maxtasks 2 -geneset "smallest1" -procs "validate_drugbank_.*" -debug 2>&1 | tee log/scipipe-$(date +%Y%m%d-%H%M%S).log 
