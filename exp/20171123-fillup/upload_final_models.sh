#!/bin/bash

KEYCLOAK_URL=http://keycloak-test2-modweb.130.238.55.60.nip.io/auth
read -p "Username: " USERNAME
read -p "Password: " PASSWORD
echo "$USERNAME:$PASSWORD"
MW_URL=http://modelingweb-test4-modweb.130.238.55.60.nip.io

RESULT=$(curl --data "grant_type=password&client_id=modelingweb&username=$USERNAME&password=$PASSWORD" ${KEYCLOAK_URL}realms/toxhq/protocol/openid-connect/token)
echo "Keycloak result: \n"$RESULT
TOKEN=$(echo $RESULT | jq -r '.access_token')

for mdl in dat/final_models/*/*r1*jar; do
    echo "--------------------------------------------------------------------------------";
    echo "Trying to upload $mdl ..."
    echo "--------------------------------------------------------------------------------";
    COMMAND="curl -F \"category=ptptest\" -F \"filecontent=@$mdl\" --header \"Authorization: bearer $TOKEN\" ${MW_URL}api/v1/models/fromFile";
    echo "Executing command: "$COMMAND
    RESULT=$($COMMAND)
    echo "Result from command: "
    echo $RESULT
    sleep 0.5
done;
