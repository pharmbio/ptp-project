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
  - [x] Make the merge ID file have only one column
  - [x] Extract only withdrawn, and approved only as to fill up to 1000 molecules
  - [x] Create filtering component
    - Some hints on how to do it:
      https://stackoverflow.com/questions/14062402/awk-using-a-file-to-filter-another-one-out-tr
  - [x] Fix bug that duplicate IDs of the same molecule can occur, because
    both CHEMBL and PubChem IDs are merged too simply right now (molecules need
    to be kept together when we want to select how many to pick, etc)
  - [x] Fix bug from the fact that approved/withdrawn status in drugbank raw
    data is not mutually exclusive
  - [x] Fix out of memory error from SLURM, on the remove_conflicting step
  - [x] Save excapeDB dataset port as a variable representing the excapedb
    dataset (new approach to easier workflow authoring I just realized)
  - [ ] Fix bug: We want to select drugbank molecules to remove that are
    available in ExcapeDB, otherwise our withdrawn molecules are too few (right
    now 918 instead of 1000)
    - Compared with:
      - `wc -l ../../raw/pubchem.chembl.dataset4publication_inchi_smiles.gisa.tsv`
      - `wc -l dat/drugbank_compids_to_remove.csv.onecol.csv`
      - Difference was: 70448221−70447303=918
    - Turns out: Maybe not(?!!). See chat log:
      > @channel: Vi måste välja ett annat antal än 1000 att dra bort från ExcapeDB.
      >
      > Bara 807 approved och 66 withdrawn, i DrugBank, som inte finns i ExcapeDB.
      >
      > ```bash
      > [dat]$ wc -l drugbank_withdrawn.csv.compids.csv{,.inexcapedb.csv}
      >  200 drugbank_withdrawn.csv.compids.csv
      >   66 drugbank_withdrawn.csv.compids.csv.inexcapedb.csv
      >  266 total
      > [dat]$ wc -l drugbank_approved.csv.compids.csv.uniq_appr.csv{,.inexcapedb.csv}
      >  2089 drugbank_approved.csv.compids.csv.uniq_appr.csv
      >   807 drugbank_approved.csv.compids.csv.uniq_appr.csv.inexcapedb.csv```
      > ```
      >
      > staffan [5:51 PM]
      > @saml varför måste alla finnas i ExcapeDB med? räcker det inte med att vi har 1000 som vi vet utfallet för och så tränar vi på resterande data? är väl inget egenvärde i att kunna plocka bort dem från ExcapeDB-data?
      >
      > saml [5:52 PM]
      > @staffan Hmm, det har du kanske rätt i ja :slightly_smiling_face: (edited)
      > Körde faktiskt så först, men fick sedan för mig att vi behövde plocka från dem som finns i excapedb ... hmm ...
      >
      > staffan [5:57 PM]
      > känns som att huvud-idéen är att vi inte tränar på samma compounds som vi använder i valideringen
      >
      > saml [5:57 PM]
      > Nu inser jag dessutom att jag inverterat uttrycket ...
      > Det är snarare så att 66 st withdrawn, och 807 approved *inte* finns med i excapedb ...
      > ... så om vi skulle validera mot bara dessa, så skulle vi inte ens behöva ta bort nåt från excapedb
      > Det ter sig ju nästan lockande :slightly_smiling_face:
      >
      > staffan [5:59 PM]
      > hur många har vi då som vi kan validera med? var det drugbank som hade typ 11k totalt?
      >
      > saml [6:00 PM]
      > Mjao, fast av dessa är bara 2550 approved small molecule drugs (edited)
      > Men vi skulle alltså kunna validera mot 873 st (807 approved + 66 withdrawn) drugbank compounds helt utan att göra nåt åt excapedb-datat (edited)
      > Man skulle kanske vilja åtminstone se till att börja med dessa i det man plockar ut ... och sedan komplettera med compounds som finns i excapedb, tills man når 1000. Men workflowet blir å andra sidan alltmer komplext då ... (edited)
      > (Man anar inte riktigt hur komplexa operationer det blir av att göra nåt som verkar så intuitivt på vita tavlan :thinking_face:) (edited)
      >
      > Alternativt så backar jag tillbaka till hur jag gjorde från början: D.v.s. bryr mig inte i om valda drugbank compounds finns i excapedb eller inte ... men tar bort dem därifrån ifall de finns där. (edited)