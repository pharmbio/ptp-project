# Experiment journal for experiment: Compare fillup vs not fillup

Main points:

- [x] Re-work select cost component to work on Overall Observed Fuzziness, not
  Class-averaged OF
- [x] Add seed (same seed within one replicate) to CPSign crossvalidate and train commands
- [>] Run full workflow *with* fillup (of assumed non-actives)
- [>] Run full workflow *without* fillup (of assumed non-actives)
- Had started this workflow overnight on messi/cihost
- In the morning (today, April 20) it had stopped upon a non-existent directory
  causing a problem for a pure Go component. [Fixed it in scipipe now](https://github.com/scipipe/scipipe/commit/05b6a8).
- [x] Perhaps should sort according to number of actives, not total number?
- [ ] Do some proper statistics to say with confideince which one is (significantly) better.
  - Paired t-tests, (between fillup/non-fillups, within each replicate?)
  - ANOVA?

Some more remaining points:

- [ ] Add calibration plot
