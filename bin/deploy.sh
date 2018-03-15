#!/bin/bash
MODELFILE=$1
CATEGORY=$2
USERNAME=$3
PASSWORD=$4
if [[ -z $MODELFILE || -z $CATEGORY ]]; then
    echo "Usage: deploy.sh <modelfile.jar> <category> [username [password]]";
    exit 1
fi;

KEYCLOAK_URL=http://keycloak-test2-modweb.130.238.55.60.nip.io/auth/

if [[ -z $USERNAME ]]; then
    echo "User name:"
    read USERNAME
fi;
if [[ -z $PASSWORD ]]; then
    echo "Password:"
    read PASSWORD
fi;
MW_URL=http://modelingweb-test4-modweb.130.238.55.60.nip.io/
RESULT=`curl --data "grant_type=password&client_id=modelingweb&username=$USERNAME&password=$PASSWORD" ${KEYCLOAK_URL}realms/toxhq/protocol/openid-connect/token`
TOKEN=`echo $RESULT | jq -r '.access_token'`

RESULT=`curl \
-F "category=$CATEGORY" \
-F "filecontent=@$MODELFILE" \
--header "Authorization: bearer $TOKEN" \
${MW_URL}api/v1/models/fromFile`

echo $RESULT
