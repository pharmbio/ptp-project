#!/bin/bash
go run fillup_propertrain_wf.go components.go -threads 1 -maxtasks 3 -geneset smallest3 # -debug | tee scipipe-$(date +%Y%m%d-%H%M%S).log
