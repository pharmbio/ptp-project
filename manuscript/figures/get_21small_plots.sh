#!/bin/bash
for c in orig fill; do 
    cp ../../exp/20180419-fillup-vs-not/res/final_models_summary.sorted.tsv."$c".pdf 21small_"$c".pdf;
done
