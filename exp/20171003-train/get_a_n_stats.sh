#!/bin/bash
for f in $(ls -Sr dat/*/*.tsv); do 
    # Extract active/nonactive counts
    a=$(awk '{ print $2 }' $f | grep A | wc -l); 
    n=$(awk '{ print $2 }' $f | grep N | wc -l); 
    # Clean up extra cruft in filename
    f2=${f#dat/}; 
    echo ${f2%.tsv},$a,$n; 
done | tr ',' '\t' | column -t | tee a_n_stats.tsv
