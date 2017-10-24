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
par(mai=c(1.2,1.2,1.2,1.2))
# Plot active/nonactive compounds
bp <- barplot(counts, 
        names=d$Gene, 
        beside = FALSE, 
        col=c("white", "#dddddd"), 
        main = 'Active / Nonactive compounds count (log-scaled)', 
        log="y", 
        las=2,
        cex.names=0.8,
        ylim=c(1,1e8),
        legend = FALSE,
        xlab=NA)
mtext("Compounds", side=2, line=3.6);
legend("bottomright", 
       c("Active", "Nonactive"),
       fill=c("white", "#dddddd", NA, NA),
);

par(new=TRUE);
      
# Plot training time (minutes)
plot(bp,d$ExecTimeMS/(1000*60), type="b", col="red", axes=FALSE, log="y", ylab=NA, xlab=NA);
axis(2, las=2, col.axis="red", col.ticks="red");
mtext("Training time (m)", side=2, line=5, col="red")
par(new=TRUE)

# Plot 1-Efficiency
plot(bp,1-d$Efficiency, type="b", axes=FALSE, col="blue", col.axis="blue", las=2, ylab=NA, xlab=NA, ylim=c(0,1));
axis(4, las=2, col="blue", col.axis="blue")
mtext("1-Efficiency", side=4, line=4, col="blue")

# --------------------------------------------------------------------------------
# Alternative legend, with the line plots included:
# --------------------------------------------------------------------------------
#legend("bottomright", 
#       c("Active", "Nonactive", "Training time (m)", "1-Efficiency"), 
#       pch=c(NA, NA, 1, 1), 
#       col=c(NA, NA, "red", "blue"), 
#       fill=c("white", "#dddddd", NA, NA),
#       border=c("black", "black", NA, NA),
#   );
# --------------------------------------------------------------------------------

dev.off()
# Avoid sending non-zero exit values on exit
quit(save = "no", status = 0, runLast = FALSE)