# Without Drugbank Experiment - TODO and Journal

- [x] Copy Go files from fillup-vs-not experiment
- [x] Update with path fixes from fillup-vs-not experiment
- [x] Clean up copied Go files of stuff not relevant to this experiment
- [x] Implement a map for looking up target/gene-specific cost values (since
  this is extracted in a previous workflow, and we don't want to re-run the
  cost-search)
  - [x] Decide if we should do this by referring to the output of the
    previous experiment, and so get the full audit trail of that file into
    outputs of this workflow?
    - Went with just hard-coded values right now
- [>] Include relevant code (for figuring out drugbank active compounds) from
  the drugbank-vs-not experiment
  - [x] Add extraction and merge components from excapedb-vs-drugbank experiement
  - [ ] Create filtering component
    - Some hints on how to do it:
      https://stackoverflow.com/questions/14062402/awk-using-a-file-to-filter-another-one-out-tr
    - [x] Make the merge ID file have only one column
    - [ ] Extract only withdrawn, and approved only as to fill up to 1000 molecules
