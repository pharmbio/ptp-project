#!/bin/bash

KEYCLOAK_URL=http://keycloak-test2-modweb.130.238.55.60.nip.io/auth
USERNAME="test"
PASSWORD="test"
MW_URL=http://modelingweb-test4-modweb.130.238.55.60.nip.io

RESULT=`curl --data "grant_type=password&client_id=modelingweb&username=$USERNAME&password=$PASSWORD" ${KEYCLOAK_URL}realms/toxhq/protocol/openid-connect/token`
TOKEN=`echo $RESULT | jq -r '.access_token'`

RESULT=`curl \
    -F "category=testcategory" \
    -F "filecontent=@sampleModel.jar" \
    --header "Authorization: bearer $TOKEN" \
    ${MW_URL}api/v1/models/fromFile`

echo $RESULT
