#!/bin/bash
go run fillup_vs_not.go components.go -threads 1 -maxtasks 3 -geneset smallest1 2>&1 | tee scipipe-$(date +%Y%m%d-%H%M%S).log # -debug
