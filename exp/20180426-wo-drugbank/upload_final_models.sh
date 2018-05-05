#!/bin/bash

set +x

KEYCLOAK_URL=http://keycloak-test2-modweb.130.238.55.60.nip.io/auth
read -p "Username: " USERNAME
read -p "Password: " PASSWORD
echo "$USERNAME:$PASSWORD"
MW_URL=http://modelingweb-test4-modweb.130.238.55.60.nip.io

KC_COMMAND="curl --data grant_type=password&client_id=modelingweb&username=$USERNAME&password=$PASSWORD ${KEYCLOAK_URL}/realms/toxhq/protocol/openid-connect/token"
echo "Executing Keycloak command: $KC_COMMAND";
RESULT=$($KC_COMMAND)
echo "Keycloak result: "$RESULT
TOKEN=$(echo $RESULT | jq -r '.access_token')
echo "Keycloak token: "$TOKEN

for mdl in dat/final_models/*/r1/fill/*r1*jar; do
    echo "--------------------------------------------------------------------------------";
    echo "Trying to upload $mdl ..."
    echo "--------------------------------------------------------------------------------";
    RESULT=`curl -F "category=ptp-wo-drugbank" -F "filecontent=@$mdl" --header "Authorization: bearer $TOKEN" ${MW_URL}"/api/v1/models/fromFile"`;
    echo $RESULT
    sleep 0.5
done;
