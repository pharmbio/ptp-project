#!/bin/bash
awk -F"|" '{ c[$1,$2]++; x[$1,$2][c[$1,$2]] = $0; } (c[$1,$2] > 1) { for (i in x[$1,$2]) print x[$1,$2][i] }' temp.tsv > temp.conflicting.tsv
