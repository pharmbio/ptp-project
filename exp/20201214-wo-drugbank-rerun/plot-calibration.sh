#!/bin/bash -l
go run wo_drugbank_wf.go components.go -threads 1 -maxtasks 2 -geneset "bowes44min100percls" -procs "plot_calib.*"
