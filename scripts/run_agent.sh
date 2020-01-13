#!/usr/bin/env sh
#
# This script adds trusted certificates on Linux Debian.
# Copies certificate from certs/ directory.
# 
# This can be used by docker containers to verify self-signed certificates
# of our proxy servers.

cp certs/* /usr/local/share/ca-certificates
update-ca-certificates
./optimizely
