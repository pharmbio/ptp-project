
# Reproducible Probabilistic Target Profiles (aka PTP)

This is the source code with workflows and pipelines for producing the models
behind the paper:

Lampa, S., Alvarsson, J., Arvidsson Mc Shane, S., Berg, A., Ahlberg, E., &
Spjuth, O. (2018).
[Predicting Off-Target Binding Profiles With Confidence Using Conformal Prediction](https://doi.org/10.3389/fphar.2018.01256).
Frontiers in pharmacology, 9, 1256.

## Code structure

The computational experiments are found under the
[`exp`](https://github.com/pharmbio/ptp-project/tree/master/exp) folder.
The experiment that produced the final models is available in
[`exp/20180426-wo-drugbank`](https://github.com/pharmbio/ptp-project/tree/master/exp/20180426-wo-drugbank).

## Requirements

- Bash
- Awk
- [cURL](https://curl.se/)
- [Go 1.15+](https://golang.org/)
- [R](https://www.r-project.org/), with the package [getopt](https://cran.r-project.org/web/packages/getopt/index.html)
- Java
- [CPSign](https://arosbio.com/cpsign/download/), either version 0.6.14 or 1.5.0

## Misc links

- [CPSign Documentation](http://cpsign-docs.genettasoft.com/)
