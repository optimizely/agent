#!/bin/bash

SERVER_AGENT_VERSION=2.2.3
SERVER_AGENT_DOWNLOAD_URL https://github.com/undera/perfmon-agent/releases/download/${SERVER_AGENT_VERSION}/ServerAgent-${SERVER_AGENT_VERSION}.zip
SERVER_AGENT_HOME=/usr/local/ServerAgent-${SERVER_AGENT_VERSION}
PATH=${SERVER_AGENT_HOME}:${PATH}

mkdir -p /tmp/serverzip  \
	&& curl -L --silent ${SERVER_AGENT_DOWNLOAD_URL} >  /tmp/serverzip/ServerAgent-${SERVER_AGENT_VERSION}.zip  \
	&& unzip -q /tmp/serverzip/ServerAgent-${SERVER_AGENT_VERSION}.zip -d /usr/local/ServerAgent-${SERVER_AGENT_VERSION}  \
	&& mv ${SERVER_AGENT_HOME}/lib/libsigar-amd64-linux.so ${SERVER_AGENT_HOME}/lib/libsigar-amd64-linux   \
	&& rm -r /tmp/serverzip

echo "running perfmon"
bash ${SERVER_AGENT_HOME}/startAgent.sh




