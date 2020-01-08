# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.9.0] - January 6th, 2019
- Bump to 1.0.0-rc1@d1b332c of the Optimizely go-sdk
- Add metric visibility into event dispatcher
- Miscellaneous clean-up and of docs and openapi spec
- Add top level config package to consolidate configuration
- Incorporate OptimizelyConfig into feature and experiment models
- Add get experiment and list experiment endpoints
- Add user features endpoint for batched decision responses
- Add windows tooling
- Add credit section to README
- Improve service shutdown

## [0.8.1] - December 4th, 2019
- Bump to 1.0.0-rc1@973644b of the Optimizely go-sdk
- Update test harness with new interface

## [0.8.0] - November 18th, 2019
- Adds ability to limit the number of active api connections
- Allows SDK keys to be bootstrapped during startup
- Adds http server timeouts
- Adds graceful shutdown hooks
- Adds support for forced variation API
- Adds support for experimentation APIs

## [0.7.0] - November 7th, 2019
- Adds request timing metrics
- Allows config file location to be set
- Bumps go-sdk version to latest master

## [0.6.0] - October 31st, 2019
* Adds a few more debug logs
- Updates to latest master to resolve some targeting issues
- Update make clean to clean mod cache

### Bug Fixes
- Actually enable the impression tracking endpoint in the router

## [0.5.0] - October 24th, 2019
- Adds GET endpoint for user-based features
- Adds impression tracking for Feature Tests

## [0.4.0] - October 23rd, 2019
- Adds admin endpoints for health, info and metrics
- Adds requestId to logs and response header
- Improves log integration with Optimizely SDK
- Updates swagger spec to match current server implementation
- Updates dependency version for the Optimizely SDK
- Enhance webhook service configurability

## [0.3.0] - October 14th, 2019
- Adds user centric API routes
- Introduces spf13/viper based configuration
- Adds OptlyContext middleware
- Adds webhoook support for multiple concurrent SDK keys

## [0.2.0] - October 3rd, 2019
- Adds Optimizely webhook support
- Adds full Feature MGMT support
- Adds NSQ for UserEvent message transport
- Adds support for multiple concurrent SDK keys

## [0.1.0] - September 4th, 2019
This is the initial release which supported a basic web application and go-sdk integration.

