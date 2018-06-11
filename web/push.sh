#!/bin/bash
oc rsync ./ $(oc get pod -l deploymentconfig=html -o custom-columns=NAME:.metadata.name --no-headers -n ptp):/usr/share/nginx/html/ -n ptp
