#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

offset=$1
begin=$((8080 + 10 * offset))
cluster=7000

export OPTIMIZELY_API_PORT=$((begin))
export OPTIMIZELY_ADMIN_PORT=$((begin + 1))
export OPTIMIZELY_WEBHOOK_PORT=$((begin + 2))
export OPTIMIZELY_CLUSTER_PORT=$((cluster + offset))

export OPTIMIZELY_CLUSTER_NODES=127.0.0.1:$cluster

export OPTIMIZELY_LOG_PRETTY=true
export OPTIMIZELY_API_ENABLEOVERRIDES=true

nohup ../bin/optimizely &>/dev/null &
#../bin/optimizely


open http://localhost:$((begin + 1))
