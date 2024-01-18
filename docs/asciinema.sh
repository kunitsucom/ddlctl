#!/bin/sh
suffix=$(TZ=UTC date +%Y%m%dT%H%M%SZ)
asciinema rec -c "ghostplay ./ghostplay.sh" "ddlctl_demo_${suffix}.cast"
agg "ddlctl_demo_${suffix}.cast" "ddlctl_demo_${suffix}.gif"
