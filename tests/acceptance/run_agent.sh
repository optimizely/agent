#!/usr/bin/env bash

# make sure that run_agent.sh file is next to agent dir
# then we can use $PWD/agent as the path

# if PWD is Agent then proceed
# if PWD is acceptance, then go two dirs up

#if [$PWD == "agent"]
#then
#  make run
#elif [$PWD == "acceptance"]
#then
#  cd ../../agent/
#  make run
#fi

#VAR=/home/me/mydir/file.c
#
#$ DIR=$(dirname "${VAR}")
#
#$ echo "${DIR}"
#/home/me/mydir
#
#$ basename "${VAR}"
#file.c


echo $PWD
#if [[ $PWD == * ]]
#then
#    echo true
#else
#    echo false
#fi
#
#
#echo $PWD
#cd ../../
#echo $PWD
#make run
