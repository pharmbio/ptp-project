#!/bin/bash -l
cp ../../../exp/20180426-wo-drugbank/res/validation/*/*.pdf .
for f in *.pdf; do pre=$(echo $f | sed 's/\.valstats.*//g' | tr "." "_"); mv $f $pre"_valplot.pdf"; done
cp ../../../exp/20180426-wo-drugbank/res/validation/*.pdf .

mv valstats.0p8.tsv.pdf alltargets_0p8_valplot.pdf
mv valstats.0p9.tsv.pdf alltargets_0p9_valplot.pdf
