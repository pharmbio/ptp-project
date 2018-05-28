package main

import (
	sp "github.com/scipipe/scipipe"
	spc "github.com/scipipe/scipipe/components"
)

func main() {
	wf := sp.NewWorkflow("build_manuscript", 1)

	copyFillup := wf.NewProc("copy_fillup_plot", "cp {i:origplot} {o:fillup_plot}")
	copyFillup.In("origplot").Connect(spc.NewFileSource(wf, "origplot", "../exp/20171123-fillup/res/final_models_summary.sorted.tsv.plot.png").Out())
	copyFillup.SetPathStatic("fillup_plot", "figures/allmodels_fillup.png")

	compileTex := wf.NewProc("compile_tex", `pdf={o:pdf} && latexmk -pdf -pdflatex="pdflatex --shell-escape" ptp.tex && mv ${pdf%.tmp} $pdf # Require: {i:copyfillup} Produces: {o:pdf}`)
	compileTex.In("copyfillup").Connect(copyFillup.Out("fillup_plot"))
	compileTex.SetPathStatic("pdf", "ptp.pdf")

	wf.Run()
}
