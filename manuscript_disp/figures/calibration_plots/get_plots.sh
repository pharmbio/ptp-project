#!/bin/bash -l
cp ../../../exp/20180426-wo-drugbank/dat/*/r1/fill/*pdf .
for f in *; do g=$(echo $f | sed s/.r1.*//g); mv $f $g"_calib.pdf"; done
