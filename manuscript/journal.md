# Manuscript Journal / TODO

## To do

- [x] Add a few representative calibration plots
- [x] Add all calibration plots in supplementary material
- [x] Add result from fillup vs not fillup
- [x] Fix Norinder style plot
  - [x] For 0.9 confidence level
      - [x] One for all datasets
      - [x] One per target (Unless too few predicted compounds per target)
  - [x] For 0.8 confidence level
      - [x] One for all datasets
      - [x] One per target (Unless too few predicted compounds per target)
- [x] Check if terbutalin is in the training dataset (Update: It's not, it's
  among the withheld ones)
- [x] Create workflow picture
- [x] Add detailed workflow plot for fillup-vs-not
- [x] Beef up introduction (Jonalv? Saml?)
- [x] Write up some text in the results
- [x] Write up some text in the discussion
- [x] Add detailed workflow plot for wo-drugbank
- [x] Write Methods section about fillup-vs-not (Samuel)
- [x] Class membership plots: Use same order of plot labels as in Norinder 2015
- [x] Class membership plots: Use same colors as in fillup plots
- [x] Add 0, 0.5 and 1 as ticks in calibration plots
- [x] Include plot of large targets (wo-drugbank) (Samuel)
  - [x] Plot the large targets as a separate plot
  - [x] Maybe do 1, 2 and 3 as subplots
- [x] Extend discussion (Ola)
- [x] Describe "Original CP Efficiency" (Staffan?)
- [x] Try making class membership plots into bubble plots (Staffan)
- [x] Rename class membership plots to "predicted versus observed" or similar,
  in the text (Samuel)
- [x] Write something about "Credibility" in discussion (Staffan?)
- [x] Original class: Observed
- [x] Update plots about "M criterion" (Samuel)
- [x] Reword "fill-up" to something like "adding assumed non-actives" (Samuel)
- [x] Create table of all the datasets (Samuel)
- [x] Write conclusion (All)
- [x] Upload models to Zenodo, and cite
- [x] Write about separate treatment of assumed negatives, and explain what
  this means (Staffan?, Samuel?)
- [x] Press more on what we provide and what is new, and its implications (All)
- [x] Add URL to reference page (Jonathan?)
- [x] Make a short nice URL via pharmb.io and logo (Jonathan?)

## Fixes for revision 1:

- [x] Update figure (explaining A and N labels)
- [x] Add more references (Chembench, OCHEM etc)
- [-] Possibly add more general references
- [x] Explain role of ExCAPE-DB more.
- [x] Update manuscript about imbalanced-ness
- [x] Add refs about CP for unbalanced datasets
- [x] Clarify what the A label means (in terms of Molar conc)
- [x] Create dataset figure
- [x] Add dataset figure
- [x] Refer to dataset figure from table caption
- [x] Split methods part of `External validation' subsection and move into Methods section
- [x] Play down role of imbalanced-ness (don't mention that particular term in most cases)
- [>] Guide the user more regarding Conformal Prediction
- [>] Extend discussion about model quality, based on plots in figure 4 (prev fig 3)
- [ ] Height 3 -> Height 1-3
- [ ] Mention that downloadable models need a license and software
- [ ] Add dataset behind the “ball plots” as .TSV file in the supplement
- [ ] Add table with selectivity and specificity for all targets?
- [ ] Add numbers in the figure as well(?)

## EAs suggestions

- [?] Add Bender CP refs
- [x] Add line of unity in calibration plots
- [x] Describe graph and colors in fig 5a
- [x] Write about color scale in fig 5b
- [x] Describe colors in fig 6

## Project pre-publication todos (not tied to manuscript)

- [ ] Rerun with DrugBank data included
  - [ ] Publish updated models on Zenodo
- [ ] Make a packaged and "re-usable" pipeline?

## On Hold

- [ ] Change to use \url for URLs
- [ ] Put long URLs in footnotes
