#!/usr/bin/env bash
set -e

# Attaches the generate_secret binaries to the GitHub release for $RELEASE_TAG.
# Uses `gh` (preinstalled on GitHub-hosted runners) instead of the deprecated `hub` CLI.
# `gh release upload` supports per-asset display labels via the "path#label" syntax.
gh release upload "$RELEASE_TAG" --clobber \
  "/tmp/output_packages/generate_secret-linux-amd64-$APP_VERSION.tar.gz#generate_secret $APP_VERSION for Linux 64-bit" \
  "/tmp/output_packages/generate_secret-darwin-amd64-$APP_VERSION.tar.gz#generate_secret $APP_VERSION for MacOS 64-bit" \
  "/tmp/output_packages/generate_secret-windows-amd64-$APP_VERSION.tar.gz#generate_secret $APP_VERSION for Windows 64-bit"
