#!/bin/bash -l
go run wo_drugbank_wf.go components.go -threads 1 -maxtasks 10 -geneset "bowes44min100percls" -procs "validate_drugbank_.*" &> log/scipipe-$(date +%Y%m%d-%H%M%S).log # -debug