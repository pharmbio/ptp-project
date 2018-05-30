#!/usr/bin/Rscript

# ------------------------------------------------------------------------
# Commandline parsing
# ------------------------------------------------------------------------
library(getopt)
optspec = matrix(c(
  'infile', 'i', 1, 'character',
  'outfile', 'o', 1, 'character',
  'format', 'f', 1, 'character',
  'gene', 'g', 2, 'character'
), byrow=TRUE, ncol=4);
opt = getopt(optspec);

# if help was asked for print a friendly message
# and exit with a non-zero error code
if ( is.null(opt$format) || is.null(opt$infile) || is.null(opt$outfile) || is.null(opt$gene) ) {
  cat('Usage: Rscript plot_heatmap.r -i infile -o outfile -f (png|pdf) -g genename\n');
  q(status=1);
}

# ------------------------------------------------------------------------
# Create and plot the heatmap
# ------------------------------------------------------------------------
# Set output format
if (opt$format == 'png') {
  png(opt$outfile, width=320, height=280, units="px")
} else if (opt$format =='pdf') {
  pdf(opt$outfile, width=4, height=4);
}

d <- read.csv(opt$infile, sep = '\t', header = TRUE);
#d <- read.csv("res/validation/htr2a/htr2a.valstats.tsv", sep = '\t', header = TRUE);
#setwd("~/mnt/ptp/exp/20180426-wo-drugbank/")

rownames(d) = d[,1] # Set rownames from first column
colnames(d) = c("Orig Label", "None", "Active", "Non-active", "Both")
dplot <- as.matrix(d[,2:5]) # Don't include first col in matrix, and make into matrix
barplot(dplot)
legend("topright", c("Orig Active", "Orig Non-active"), fill=c("black", "grey"))
mtext(paste("Class membership change for (", opt$gene, ")", sep=""))
dev.off()
# Avoid sending non-zero exit values on exit
quit(save = "no", status = 0, runLast = FALSE)
