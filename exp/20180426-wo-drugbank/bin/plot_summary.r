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
	pdf(opt$outfile, width=5.75, height=6);
}

d <- read.csv(opt$infile, sep = '\t', header = TRUE);
# --------------------------------------------------------------------------------
# Read in file manually for debugging:
# --------------------------------------------------------------------------------
# dev.off()
# setwd(dir = "/media/samuel/SAMUELLAMPA/proj/ptp/exp/20171003-train/")
# d <- read.csv("res/final_models_summary.sorted.tsv", sep = '\t', dec=".", header = TRUE, quote="")
# --------------------------------------------------------------------------------

drepl <- split(d, d$Replicate)

invert <- function(x) (
  return(1-x)
)

counts <- rbind(drepl$r1$ActiveCnt, drepl$r1$NonactiveCnt)

# Force to avoid scientific numerical format (sci-penalty)
options(scipen=1, digits="3");

# Set margins (in inches, thus mai), for the whole plot
par(mai=c(1.2,1.2,1.2,1.6))

# Plot active/nonactive compounds
bplt <- barplot(counts,
        names=drepl$r1$Gene,
        beside = FALSE,
        col=c("white", "#dddddd"),
        #main = "Compound counts, training time and observed fuzziness per target",
        las=2,
        cex.names=0.8,
        legend = FALSE,
        ylim=c(0,400000),
        xlab=NA,
        ylab=NA,
        axes=FALSE);
#axis(2, las=2, col.axis="black");
axis(2, las=2, col.axis="black", at=c(0, 100000, 200000, 300000, 400000), labels=c("0", "100 k", "200 k", "300 k", "400 k"));
mtext("Compounds",
      side=2,
      line=3.6);
legend("right",
       c("Active", "Nonactive"),
       fill=c("white", "#dddddd"),
       bg="white");

# Ugly hack to get the sorting right: Get a list of total counts, that is sorted
# by alphabetic sort of gene names. This will work well to get sorting by total
# counts, on another vector (ofca_median for example, in this case) that is
# sorted alphabetically by gene name:
sort_vector_totcounts <- aggregate(d$TotalCnt, by=list(Gene = d$Gene), FUN=median)

# ------------------------------------------------------------------------
# Set up colors
# ------------------------------------------------------------------------
col_eff <- "#368645"
col_of <- "#167391"
col_caof <- "#673BA8"
#col_eff <- "#339933"
#col_of <- "#006699"
#col_caof <- "#996600"

# --------------------------------------------------------------------------------
# Plot Efficiency
# --------------------------------------------------------------------------------
#par(new=TRUE)
#plot(bplt, 1-drepl$r1$Efficiency, type="p", axes=FALSE, col=col_eff, col.axis=col_eff, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, 1-drepl$r2$Efficiency, type="p", axes=FALSE, col=col_eff, col.axis=col_eff, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, 1-drepl$r3$Efficiency, type="p", axes=FALSE, col=col_eff, col.axis=col_eff, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
par(new=TRUE);
plot(bplt, 1-drepl$r1$Efficiency, type="p", axes=FALSE, col=col_eff, col.axis=col_eff, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
eff_median <- aggregate(d$Efficiency, by=list(Gene = d$Gene), FUN=median)
eff_median <- eff_median[order(sort_vector_totcounts$x),]
par(new=TRUE);
plot(bplt, 1-eff_median$x, type="l", axes=FALSE, col=col_eff, col.axis=col_eff, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
axis(4, las=2, col="black", col.axis="black", at=c(0,0.5,1), labels=c("1", "0.5", "0"));
mtext("Efficiency (prediction set size)", side=4, line=3.6, col=col_eff)
# --------------------------------------------------------------------------------


# --------------------------------------------------------------------------------
# Plot observed fuzziness (Overall)
# --------------------------------------------------------------------------------
#par(new=TRUE)
#plot(bplt, 1-drepl$r1$ObsFuzzOverall, type="p", axes=FALSE, col=col_of, col.axis=col_of, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, 1-drepl$r2$ObsFuzzOverall, type="p", axes=FALSE, col=col_of, col.axis=col_of, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, 1-drepl$r3$ObsFuzzOverall, type="p", axes=FALSE, col=col_of, col.axis=col_of, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
par(new=TRUE);
plot(bplt, 1-drepl$r1$ObsFuzzOverall, type="p", axes=FALSE, col=col_of, col.axis=col_of, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
ofca_median <- aggregate(d$ObsFuzzOverall, by=list(Gene = d$Gene), FUN=median)
ofca_median <- ofca_median[order(sort_vector_totcounts$x),]
par(new=TRUE);
plot(bplt, 1-ofca_median$x, type="l", axes=FALSE, col=col_of, col.axis=col_of, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#axis(4, las=2, col=col_of, col.axis=col_of, at=c(0,0.5,1), labels=c("1", "0.5", "0"));
mtext("Observed Fuzziness (OF)", side=4, line=4.8, col=col_of)
# --------------------------------------------------------------------------------


# --------------------------------------------------------------------------------
# Plot observed fuzziness (Class-Averaged)
# --------------------------------------------------------------------------------
#par(new=TRUE)
#plot(bplt, 1-drepl$r1$ObsFuzzClassAvg, type="p", axes=FALSE, col=col_caof, col.axis=col_caof, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, 1-drepl$r2$ObsFuzzClassAvg, type="p", axes=FALSE, col=col_caof, col.axis=col_caof, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, 1-drepl$r3$ObsFuzzClassAvg, type="p", axes=FALSE, col=col_caof, col.axis=col_caof, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
par(new=TRUE);
plot(bplt, 1-drepl$r1$ObsFuzzClassAvg, type="p", axes=FALSE, col=col_caof, col.axis=col_caof, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
ofca_median <- aggregate(d$ObsFuzzClassAvg, by=list(Gene = d$Gene), FUN=median)
ofca_median <- ofca_median[order(sort_vector_totcounts$x),]
par(new=TRUE);
plot(bplt, 1-ofca_median$x, type="l", axes=FALSE, col=col_caof, col.axis=col_caof, las=2, ylab=NA, xlab=NA, ylim=c(0,1));
#axis(4, las=2, col=col_caof, col.axis=col_caof, at=c(0,0.5,1), labels=c("1", "0.5", "0"));
mtext("Class-averaged Observed Fuzziness (CAOF)", side=4, line=6.0, col=col_caof)
# --------------------------------------------------------------------------------


# --------------------------------------------------------------------------------
# Plot training time (minutes)
# --------------------------------------------------------------------------------
#par(new=TRUE);
#plot(bplt, drepl$r1$ExecTimeMS/(1000*60), type="p", col="red", axes=FALSE, log="y", ylab=NA, xlab=NA);
#par(new=TRUE);
#plot(bplt, drepl$r2$ExecTimeMS/(1000*60), type="p", col="red", axes=FALSE, log="y", ylab=NA, xlab=NA);
#par(new=TRUE);
#plot(bplt, drepl$r3$ExecTimeMS/(1000*60), type="p", col="red", axes=FALSE, log="y", ylab=NA, xlab=NA);
#exectime_median <- aggregate(d$ExecTimeMS, by=list(Gene = d$Gene), FUN=median)
#exectime_median <- exectime_median[order(sort_vector_totcounts$x),]
#par(new=TRUE);
#plot(bplt, exectime_median$x/(1000*60), type="l", col="red", axes=FALSE, log="y", ylab=NA, xlab=NA);
#axis(4, las=2, col="white", col.axis="red", col.ticks="red", at=c(1,30,60), labels=c("1 min", "30 min", "1 h"));
#mtext("Training time (min)", side=4, line=2.4, col="red")
# --------------------------------------------------------------------------------


# --------------------------------------------------------------------------------
# Plot accuracy
# --------------------------------------------------------------------------------
#par(new=TRUE);
#plot(bplt, drepl$r1$Accuracy, type="p", col="purple2", axes=FALSE, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, drepl$r2$Accuracy, type="p", col="purple2", axes=FALSE, ylab=NA, xlab=NA, ylim=c(0,1));
#par(new=TRUE);
#plot(bplt, drepl$r3$Accuracy, type="p", col="purple2", axes=FALSE, ylab=NA, xlab=NA, ylim=c(0,1));
#accuracy_median <- aggregate(d$Accuracy, by=list(Gene = d$Gene), FUN=median)
#accuracy_median <- accuracy_median[order(sort_vector_totcounts$x),]
#par(new=TRUE);
#plot(bplt, accuracy_median$x, type="l", col="purple2", axes=FALSE, ylab=NA, xlab=NA, ylim=c(0,1));
# --------------------------------------------------------------------------------


# --------------------------------------------------------------------------------
# Alternative legend, with the line plots included:
# --------------------------------------------------------------------------------
#legend("bottomright",
#       c("Active", "Nonactive", "Training time (min)", "1-ClassAvgObsFuzz"),
#       pch=c(NA, NA, 1, 1),
#       col=c(NA, NA, "red", "blue"),
#       fill=c("white", "#dddddd", NA, NA),
#       border=c("black", "black", NA, NA),
#   );
# --------------------------------------------------------------------------------

dev.off()
# Avoid sending non-zero exit values on exit
quit(save = "no", status = 0, runLast = FALSE)
