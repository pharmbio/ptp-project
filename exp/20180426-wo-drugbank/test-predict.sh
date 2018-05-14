id=$1
smiles=$2
if [[ -z $id || -z $smiles ]]; then
    echo "Usage: ./test-predict.sh <id> <smiles>";
else
    for mdl in $(ls -d dat/final_models/*); do
        tgt=${mdl#dat/final_models/};
        java -jar ../../bin/cpsign-0.6.14.jar \
            predict --license ../../bin/cpsign.lic \
            -c 1 \
            -im \
            -if $id"-"$tgt \
            -m dat/final_models/$tgt/r1/fill/$tgt.r1.fill.liblin_c*_nrmdl10.mdl.jar \
            -sm "$smiles";
    done;
fi;
