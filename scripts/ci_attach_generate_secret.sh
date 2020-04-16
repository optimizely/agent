#!/usr/bin/env bash
set -e

hub release edit "$TRAVIS_TAG" -m "" \
  -a "/tmp/output_packages/generate_secret-linux-amd64-$APP_VERSION#generate_secret $APP_VERSION for Linux 64-bit" \
  -a "/tmp/output_packages/generate_secret-darwin-amd64-$APP_VERSION#generate_secret $APP_VERSION for MacOS 64-bit" \
  -a "/tmp/output_packages/generate_secret-windows-amd64-$APP_VERSION.exe#generate_secret $APP_VERSION for Windows 64-bit"
