#!/usr/bin/Rscript

# ------------------------------------------------------------------------
# Commandline parsing
# ------------------------------------------------------------------------
library(getopt)
optspec = matrix(c(
  'infile', 'i', 1, 'character',
  'outfile', 'o', 1, 'character',
  'format', 'f', 1, 'character'
), byrow=TRUE, ncol=4);
opt = getopt(optspec);

# if help was asked for print a friendly message
# and exit with a non-zero error code
if ( is.null(opt$format) || is.null(opt$infile) || is.null(opt$outfile) ) {
  cat('Usage: Rscript plot_heatmap.r -i infile -o outfile -f (png|pdf)\n');
  q(status=1);
}

# ------------------------------------------------------------------------
# Create and plot the heatmap
# ------------------------------------------------------------------------

# Set output format
if (opt$format == 'png') {
	png(opt$outfile, width=640, height=1024, units="px")
} else if (opt$format =='pdf') {
	pdf(opt$outfile);
}

d <- read.csv(opt$infile, sep = '\t', header = TRUE);

par(mfrow=c(4,1));

counts <- table(d$Active, d$Nonactive)

barplot(d$DataSetSize,names=counts, col=c("lightblue", "darkyellow"), main = 'Active / Nonactive compounds');
barplot(d$Efficiency,names=d$Gene, ylim=c(0,1), main = 'Efficiency');
barplot(d$Validity,names=d$Gene, ylim=c(0,1), main = 'Validity');
barplot(d$ModelFileSize,names=d$Gene, main = 'Model file size (bytes)');

dev.off()

quit(save = "no", status = 0, runLast = FALSE)