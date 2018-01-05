#!/bin/bash
cat temp.tsv | awk -F"|" '(( $1 != p1 ) || ( $2 != p2)) && ( c[p1,p2] <= 1 ) && ( p1 != "" ) && ( p2 != "" ) { print p1 "|" p2 "|" p3 } { c[$1,$2]++; p1 = $1; p2 = $2; p3 = $3 } END { print $1 "|" $2 "|" $3 }' > temp.filtered.tsv
