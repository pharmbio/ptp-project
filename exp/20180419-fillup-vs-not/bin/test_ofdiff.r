#!/usr/bin/Rscript

# ------------------------------------------------------------------------
# Commandline parsing
# ------------------------------------------------------------------------
library(getopt)
optspec = matrix(c(
  'infile', 'i', 1, 'character',
  'outfile', 'o', 1, 'character'
), byrow=TRUE, ncol=4);
opt = getopt(optspec);

# if help was asked for print a friendly message
# and exit with a non-zero error code
if ( is.null(opt$infile) || is.null(opt$outfile) ) {
  cat('Usage: Rscript plot_heatmap.r -i infile -o outfile\n');
  q(status=1);
}

d <- read.csv(opt$infile, sep = '\t', header = TRUE);
# --------------------------------------------------------------------------------
# Read in file manually for debugging:
# --------------------------------------------------------------------------------
#setwd(dir = "/home/samuel/mnt/ptp/exp/20180419-fillup-vs-not/")
#d <- read.csv("res/final_models_summary.sorted.tsv", sep = '\t', dec=".", header = TRUE, quote="")
# --------------------------------------------------------------------------------

dr <- split(d, d$Runset)

ofOrigMean <- aggregate( dr$orig$ObsFuzzOverall, by=list(Gene = dr$orig$Gene), FUN=mean)
ofFillMean <- aggregate( dr$fill$ObsFuzzOverall, by=list(Gene = dr$fill$Gene), FUN=mean)
ofDiffs <- ofFillMean$x - ofOrigMean$x

# Do a Wilcoxon Mann-Whitney, non-parametric test of the difference between the
# means of Observed Fuzziness Overall before and after fill-up of assumed 
# non-active compounds
wilcoxStats <- wilcox.test(ofFillMean$x, ofOrigMean$x, paired = TRUE, alternative = "less")

lapply(wilcoxStats, write, opt$outfile, append=TRUE, ncolumns=1000)

# Avoid sending non-zero exit values on exit
quit(save = "no", status = 0, runLast = FALSE)