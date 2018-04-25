
# 2018-01-03

Turns out we have a number of duplicates:

```bash
sqlite3 excapedb.db 'SELECT Gene_Symbol,SMILES,COUNT(*) FROM excapedb GROUP BY Gene_Symbol,SMILES HAVING COUNT(*) > 1' | tee duplicates.tsv

$ wc -l duplicates.tsv                                                                                                                                                                                                              
403949 duplicates.tsv

$ grep "|3" duplicates.tsv  | wc -l
297

$ grep "|4" duplicates.tsv  | wc -l                                                                          
1

$ grep "|5" duplicates.tsv  | wc -l                                                                          
0
```

... but not in terms of CID (Original_Entry_ID) and Entrez ID:

```
$ time sqlite3 excapedb.db 'SELECT Entrez_ID,Original_Entry_ID,COUNT(*) FROM excapedb GROUP BY Entrez_ID,Original_Entry_ID HAVING COUNT(*) > 1' | tee cid_duplicates.tsv                                                            

real    5m15.117s
user    3m31.975s
sys     0m34.021s

$ lltr
total 28393100
-rw-r--r-- 1 samuel samuel 29050791936  3 jan 15.10 excapedb.db
-rw-rw-r-- 1 samuel samuel    23733275  3 jan 15.18 duplicates.tsv
-rw-rw-r-- 1 samuel samuel           0  3 jan 15.39 cid_duplicates.tsv

$ wc -l cid_duplicates.tsv 
0 cid_duplicates.tsv
```
