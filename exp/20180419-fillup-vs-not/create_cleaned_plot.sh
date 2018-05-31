#!/bin/bash
go run fillup_vs_not.go components.go -geneset smallest1 -graph
head -n 1 workflow.dot > workflow.cleaned.dot;
tail -n +2 workflow.dot \
    | head -n -2 \
    | grep -Pv "_r(2|3)" \
    | sed 's/r1/r{1,2,3}/g' \
    | grep -Pv "_(100|10)" \
    | sed 's/_1/_{1,10,100}/g' \
    | sed 's/pde3a/{GENE}/g' \
    | grep -Pv "_png" \
    | sort \
    | uniq >> workflow.cleaned.dot;
tail -n 1 workflow.dot >> workflow.cleaned.dot;
dot -Tpdf workflow.cleaned.dot -o workflow.cleaned.pdf \
    && exo-open workflow.cleaned.pdf;
