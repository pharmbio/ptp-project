#!/bin/bash
latexmk -pdf -pdflatex="pdflatex --shell-escape" -pvc ptp.tex
