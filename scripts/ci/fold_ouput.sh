##!/usr/bin/env bash

name=$1
desc=$2
shift
shift

echo -e "travis_fold:start:$name\033[33;1m$desc\033[0m"
echo $($@)
echo -e "\ntravis_fold:end:$name\r"
