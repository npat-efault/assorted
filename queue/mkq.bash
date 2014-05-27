#!/bin/bash

if [ $# -ne 3 ]; then 
   echo "Usage: `basename $0` <package> <q-name> <elment type>" 1>&2
   exit 1
fi

pkg="$1"
qn="$2"
et="$3"

if echo "$qn" | grep -q '^[A-Z]'; then
   nn="New$qn"
else
   nn=new${qn^}
fi

gofmt -r "PACKAGE -> $pkg" \
| gofmt -r "NewQ -> $nn" \
| gofmt -r "Q -> $qn" \
| gofmt -r "ELTYPE -> $et"
