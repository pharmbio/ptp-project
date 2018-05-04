#!/bin/bash -l
#SBATCH -A snic2017-7-89
#SBATCH -p node
#SBATCH -n 20
#SBATCH -J ptp_fullwf_wo_drugbank
#SBATCH -t 2-00:00:00
#SBATCH --mail-user samuel.lampa@farmbio.uu.se
#SBATCH --mail-type BEGIN,FAIL,END
module load java/sun_jdk1.8.0_92
module load R/3.4.0
go run wo_drugbank_wf.go components.go -threads 1 -maxtasks 19 -geneset "bowes44min100percls" -procs "plot_summary_.*" 2>&1 | tee log/scipipe-$(date +%Y%m%d-%H%M%S).log # -debug
