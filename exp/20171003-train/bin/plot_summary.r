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
	png(opt$outfile, width=1200, height=640, units="px")
} else if (opt$format =='pdf') {
	pdf(opt$outfile);
}

d <- read.csv(opt$infile, sep = '\t', header = TRUE);
# --------------------------------------------------------------------------------
# Read in file manually for debugging:
# --------------------------------------------------------------------------------
# dev.off()
# setwd(dir = "/media/samuel/SAMUELLAMPA/proj/ptp/exp/20171003-train/")
# d <- read.csv("dat/final_models/summary.sorted.tsv", sep = '\t', header = TRUE);
# --------------------------------------------------------------------------------

counts <- rbind(d$Active, d$Nonactive)

# Force to avoid scientific numerical format (sci-penalty)
options(scipen=1, digits="0");

# Set margins (in inches, thus mai), for the whole plot
par(mai=c(1.2,1.2,1.2,1.4))

# Plot active/nonactive compounds
bp <- barplot(counts,
        names=d$Gene,
        beside = FALSE,
        col=c("white", "#dddddd"),
        main = "Compound counts, training time and efficiency per target",
        las=2,
        cex.names=0.8,
        ylim=c(0,10000),
        legend = FALSE,
        xlab=NA,
        ylab=NA,
        axes=FALSE);
axis(2, las=2, col.axis="black", at=c(0, 1000, 5000, 100000, 200000, 300000, 400000), labels=c("0", "1 k", "5 k", "100 k", "200 k", "300 k", "400 k"));
mtext("Compounds",
      side=2,
      line=3.6);
legend("bottomright",
       c("Active", "Nonactive"),
       fill=c("white", "#dddddd"));

par(new=TRUE);

# Plot training time (minutes)
plot(bp,d$ExecTimeMS/(1000*60), type="b", col="red", axes=FALSE, log="y", ylab=NA, xlab=NA);
axis(4, las=2, col="white", col.axis="red", col.ticks="red", at=c(1,30,60), labels=c("1 min", "30 min", "1 h"));
mtext("Training time (min)", side=4, line=3.6, col="red")
par(new=TRUE)

# Plot 1-Efficiency
plot(bp,1-d$Efficiency, type="b", axes=FALSE, col="blue", col.axis="blue", las=2, ylab=NA, xlab=NA, ylim=c(0,1));
axis(4, las=2, col="blue", col.axis="blue", at=c(0,0.5,1), labels=c("1", "0.5", "0"));
mtext("Efficiency", side=4, line=4.8, col="blue")

# --------------------------------------------------------------------------------
# Alternative legend, with the line plots included:
# --------------------------------------------------------------------------------
#legend("bottomright",
#       c("Active", "Nonactive", "Training time (min)", "1-Efficiency"),
#       pch=c(NA, NA, 1, 1),
#       col=c(NA, NA, "red", "blue"),
#       fill=c("white", "#dddddd", NA, NA),
#       border=c("black", "black", NA, NA),
#   );
# --------------------------------------------------------------------------------

dev.off()
# Avoid sending non-zero exit values on exit
quit(save = "no", status = 0, runLast = FALSE)
