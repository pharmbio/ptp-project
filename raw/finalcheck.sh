#!/bin/bash
for s in $(cat temp.conflicting.tsv); do grep -F $s temp.filtered.tsv; done
