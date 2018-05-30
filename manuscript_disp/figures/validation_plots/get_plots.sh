#!/bin/bash -l
cp ../../../exp/20180426-wo-drugbank/res/validation/*/*.pdf .
for f in *.pdf; do g=$(echo $f | sed 's/\.valstats.*//g'); mv $f $g"_valplot.pdf"; done
cp ../../../exp/20180426-wo-drugbank/res/validation/*.pdf .
mv valstats.tsv.pdf alltargets_valplot.pdf
