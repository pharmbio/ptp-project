Experiment Journal
==================

- [x] Copy the workflow from the previous fillup folder
- [x] Use new flag(s) for proper training set in CPSign
  - (`... crossvalidate --proper-trainfile`)
- [x] Fill up *to* 3x#actives, instead of *adding* 2x#actives new assumed negatives
- [x] Remove "target info" from names when training
- [x] Test re-run locally with 3 smallest target
- [x] Fix bug that makes the plot show the unfilled amount of non-actives in the plot
  - [x] Split up fill-up in two components: One that extracts the fill-up
    lines, and one that merges them with the existing data.
  - Note: Seems this might possibly have been caused by having multiple
    workflow runs going on at the same time. At least it worked better this
    time when running the small targets on cihost. 21 targets x 3 replicates,
    and no empty fillup files this time.
- [x] Fix minor thing making some file names contain two consecutive dots (`..`).
- [ ] Figure out why validity is `0.000` in the latest run, and effiency and
  validity a bit too good.
  staffan said:
  > måste ju vara något som inte stämmer. hur kan efficiency och fuzz vara
  > så bra?
  - Found it now: Had forgot to rename from "Validity" to "Accuracy" after
    name change in CPSign.
- [x] Implement date/time field in audit log.
- [x] Fix Validity->Accuracy name change
- [x] Re-run small datasets downstream of "extract_assumed_n_*" component,
      after fixing validity->accuracy name change
- [ ] Run full workflow *with* fillup (of assumed non-actives)
- [ ] Check up more about PROV and if we should adapt to it
- [ ] Run full workflow *without* fillup (of assumed non-actives)
