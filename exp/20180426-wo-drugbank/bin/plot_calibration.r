#!/usr/bin/Rscript

# ------------------------------------------------------------------------
# Commandline parsing
# ------------------------------------------------------------------------
library(getopt)
optspec = matrix(c(
  'infile', 'i', 1, 'character',
  'outfile', 'o', 1, 'character',
  'format', 'f', 1, 'character',
  'gene', 'g', 1, 'character'
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
	pdf(opt$outfile, width=2.8, height=3.0);
}

d <- read.csv(opt$infile, sep = '\t', header = TRUE);

plot(d$confidence, d$accuracty, xlab="Confidence", ylab="Accuracy", xlim=c(0,1), ylim=c(0,1), axes=FALSE)
axis(side=1, at=c(0.0,0.5,1.0), labels=c("0", "0.5", "1.0"), tick=TRUE)
axis(side=2, at=c(0.0,0.5,1.0), labels=c("0", "0.5", "1.0"), tick=TRUE)
#box()
mtext(opt$gene)

dev.off()
# Avoid sending non-zero exit values on exit
quit(save = "no", status = 0, runLast = FALSE)
