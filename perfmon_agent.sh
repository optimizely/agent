#!/bin/sh

SERVER_AGENT_VERSION=2.2.1
SERVER_AGENT_DOWNLOAD_URL=http://jmeter-plugins.org/downloads/file/ServerAgent-${SERVER_AGENT_VERSION}.zip
SERVER_AGENT_HOME=/usr/local/ServerAgent-${SERVER_AGENT_VERSION}

mkdir -p /usr/local/ServerAgent-${SERVER_AGENT_VERSION}
wget http://jmeter-plugins.org/downloads/file/ServerAgent-${SERVER_AGENT_VERSION}.zip
unzip ServerAgent-${SERVER_AGENT_VERSION}.zip -d /usr/local/ServerAgent-${SERVER_AGENT_VERSION}
mv ${SERVER_AGENT_HOME}/lib/libsigar-amd64-linux.so ${SERVER_AGENT_HOME}/lib/libsigar-amd64-linux 
rm -rf ServerAgent-${SERVER_AGENT_VERSION}.zip \
			${SERVER_AGENT_HOME}/startAgent.bat \
			${SERVER_AGENT_HOME}/lib/*.dylib \
			${SERVER_AGENT_HOME}/lib/*.dll \
			${SERVER_AGENT_HOME}/lib/*.lib \
			${SERVER_AGENT_HOME}/lib/*.sl \
			${SERVER_AGENT_HOME}/lib/*.so

mv ${SERVER_AGENT_HOME}/lib/libsigar-amd64-linux ${SERVER_AGENT_HOME}/lib/libsigar-amd64-linux.so
rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
	
echo "running perfmon"
${SERVER_AGENT_HOME}/startAgent.sh
