#!/usr/bin/env bash

# This conditional is to be able to run the Agent when tests are in "agent" repo as well as
# when tests are in "acceptance" repo. Because paths are different.
#echo $PWD
#if [[ $PWD == */agent/tests/acceptance ]]
#then
#    cd ../../
#fi
#
#make run


echo $PWD
cd ../../
make run