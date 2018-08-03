#!/bin/bash

name="flottbot"

for n in *$name*
do
  mv "$n" "$name"
  platform=$(echo "$n" | awk -F[=_] '{print $2}')
  arch=$(echo "$n" | awk -F[=_] '{print $3}')
  tar czf "${name}"-"${platform}"-"${arch}".tgz "${name}"
  rm -f ${name}
done
