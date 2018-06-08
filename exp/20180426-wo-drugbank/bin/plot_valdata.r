#!/usr/bin/Rscript

# ------------------------------------------------------------------------
# Commandline parsing
# ------------------------------------------------------------------------
library(getopt)
optspec = matrix(c(
  'infile', 'i', 1, 'character',
  'outfile', 'o', 1, 'character',
  'format', 'f', 1, 'character',
  'gene', 'g', 2, 'character',
  'confidence', 'c', 2, 'character'
), byrow=TRUE, ncol=4);
opt = getopt(optspec);

# if help was asked for print a friendly message
# and exit with a non-zero error code
if ( is.null(opt$format) || is.null(opt$infile) || is.null(opt$outfile) || is.null(opt$gene) || is.null(opt$confidence) ) {
  cat('Usage: Rscript plot_heatmap.r -i infile -o outfile -f (png|pdf) -g genename -c confidence_level\n');
  q(status=1);
}

# ------------------------------------------------------------------------
# Create and plot the heatmap
# ------------------------------------------------------------------------
# Set output format
if (opt$format == 'png') {
  png(opt$outfile, width=320, height=280, units="px")
} else if (opt$format =='pdf') {
    pdf(opt$outfile, width=3, height=3.6);
}

d <- read.csv(opt$infile, sep = '\t', header = TRUE);
#d <- read.csv("res/validation/htr2a/htr2a.valstats.tsv", sep = '\t', header = TRUE);
#setwd("~/mnt/ptp/exp/20180426-wo-drugbank/")

#rownames(d) = d[,1] # Set rownames from first column
#colnames(d) = c("Orig Label", "Both", "A", "N", "None")
#dplot <- as.data.frame(d[,2:5]) # Don't include first col in matrix, and make into matrix
dplot <- as.data.frame(c(d[1,2],d[1,3],d[1,4],d[1,5],d[2,2],d[2,3],d[2,4],d[2,5]))

# ------------------------------------------------------------------------
# START: BALL PLOT
# ------------------------------------------------------------------------
library(ggplot2)
# SET FONTS
m<-20
label_font <- element_text(family="Helvetica",size=16) # family="Arial" - did not work to save with Arial
axis_font <- element_text(family="Helvetica", size=16, margin = margin(t = m,r = m,b = m,l = m))

# PARAMTERS

if (opt$gene == "all targets") {
    chart_title <- paste("Confidence:", opt$confidence, sep=" ")
} else {
    chart_title <- opt$gene
}
circle_max_size <- 18
xlabs <- c("A", "A", "A", "A", "N", "N", "N", "N")
ylabs <- c("Both", "A", "N", "None", "Both", "A", "N", "None")
# PLOT
ggplot(dplot, aes(xlabs, factor(ylabs, levels=ylabs, ordered=TRUE)))+
  geom_point(aes(size=dplot), shape=21, fill="#dddddd")+
  #scale_size_identity()+
  scale_size_area(max_size = circle_max_size, guide=FALSE)+ # I think the area option looks best (easiest to see both large and small circles)
  #scale_size(range=c(0,30) ,guide = FALSE)+
  labs(title=chart_title, x = "Observed", y= "Predicted",font=label_font)+

  # THEMES
  theme_light()+ # gray grid
  #theme_linedraw()+ # Black hard line rounding drawing area, gray grid
  #theme_void()+ # No lines, all white background

  # SETTINGS
  theme(
    axis.text = axis_font,
    axis.title = label_font,
    title = label_font,
    plot.title = element_text(hjust = 0.5) # Center chart title
  )
#ggsave(file_location, height=3.6,width=3)
# ------------------------------------------------------------------------
# END: BALL PLOT
# ------------------------------------------------------------------------


#if (opt$gene == "all targets") {
#    barplot(dplot, beside=TRUE, col=c("white", "#dddddd"), ylim=c(0,3000), axes=FALSE)
#    axis(side=2, at=c(0,1000,2000,3000), labels=c("0", "1000", "", "3000"), tick=TRUE)
#    mtext(paste("Confidence:", opt$confidence, sep=" "))
#} else {
#    barplot(dplot, beside=TRUE, col=c("white", "#dddddd"), ylim=range(pretty(c(0, dplot))))
#    mtext(opt$gene)
#}
#legend("topright", c("Orig A", "Orig N"), fill=c("black", "grey"))
dev.off()
# Avoid sending non-zero exit values on exit
quit(save = "no", status = 0, runLast = FALSE)
