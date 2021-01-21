#!/bin/bash

set +x

KEYCLOAK_URL=http://keycloak-modelingweb.os.pharmb.io/auth
read -p "Username: " USERNAME
read -p "Password: " PASSWORD
echo "$USERNAME:$PASSWORD"
MW_URL=http://modelingweb.service.pharmb.io

RESULT=$(curl --data "grant_type=password&client_id=modelingweb&username=$USERNAME&password=$PASSWORD" "${KEYCLOAK_URL}/realms/toxhq/protocol/openid-connect/token")
echo "Keycloak result: "$RESULT
TOKEN=$(echo $RESULT | jq -r '.access_token')
echo "Keycloak token: "$TOKEN

sleep 5

for mdl in dat/final_models/*/r1/fill/*r1*jar; do
    echo "--------------------------------------------------------------------------------";
    echo "Trying to upload $mdl ..."
    echo "--------------------------------------------------------------------------------";
    RESULT=$(curl -F "category=ptp" -F "filecontent=@$mdl" --header "Authorization: bearer $TOKEN" ${MW_URL}"/api/v1/models/fromFile");
    echo $RESULT
    sleep 0.5
done;
