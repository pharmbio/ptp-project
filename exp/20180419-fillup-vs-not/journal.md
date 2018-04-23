# Experiment journal for experiment: Compare fillup vs not fillup

Main points:

- [x] Re-work select cost component to work on Overall Observed Fuzziness, not
  Class-averaged OF
- [x] Add seed (same seed within one replicate) to CPSign crossvalidate and train commands
- [x] Run full workflow *with* fillup (of assumed non-actives)
- [x] Run full workflow *without* fillup (of assumed non-actives)
- Had started this workflow overnight on messi/cihost
- In the morning (today, April 20) it had stopped upon a non-existent directory
  causing a problem for a pure Go component. [Fixed it in scipipe now](https://github.com/scipipe/scipipe/commit/05b6a8).
- [x] Perhaps should sort according to number of actives, not total number?
- [x] - Fix bug that didn't separate between orig/fill for counting actives/non-actives
- [x] Do some proper statistics to say with confideince which one is (significantly) better.
  - Paired t-tests, (between fillup/non-fillups, within each replicate?)
  - Wilcoxon says the difference is fine (meaning, obsfuzz overall is clearly
    better/smaller after fillup):

  ```r
  > wilcox.test(ofFillMean$x, ofOrigMean$x, paired = TRUE, alternative = "less")

      Wilcoxon signed rank test

  data:  ofFillMean$x and ofOrigMean$x
  V = 2, p-value = 1e-06
  alternative hypothesis: true location shift is less than 0
  ```

Some more remaining points:

- [ ] Add calibration plot
