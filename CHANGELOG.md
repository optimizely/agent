# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
- Reject requests with invalid hosts, and introduce `server.allowedHosts` configuration property

## [1.3.0] - July 7th, 2020
- Upgrade to use go-sdk v1.3.0. This adds support for JSON feature variables
- Add /debug/pprof endpoints to the admin service
- Run docker container as non root user
- Log warnings when HTTPS and authorization are not enabled via configuration

## [1.2.0] - June 18th, 2020
- Expose event dispatch URL as a config parameter
- Return experimentKey and variationKey with feature test decisions
- Expose health endpoint for all listeners
- Update API docs
- Streamline CI stages

## [1.1.0] - May 21st, 2020
- Upgrade to use go-sdk 1.2.0. This adds support for multi-rule rollouts.

## [1.0.2] - April 26th, 2020
- Add datafileURLTemplate configuration option

## [1.0.1] - April 22, 2020
- Update to use go-sdk 1.1.3. This has a fix to the batch event processor was creating a dispatcher without a logger.

## [1.0.0] - March 26th, 2020
- Update documentation and examples
- Add response body for override and track
- Add userId in /activate response
- Require userId in /activate request
- Add python integration test suite
- Add route handler and serve /openapi.yaml
- Improve logging from the SDK

## [0.14.0] - March 12th, 2020
- Update windows build script
- Remove pre-v1 api references
- Allow unknown keys in /activate request
- Improve API token issuer and auth client config
- Expand on example python scripts and general documentation

## [0.13.0] - March 5th, 2020
- Add ability to blacklist TLS ciphers
- Disable notifications and overrides by default
- Update misc items in swagger spec
- Improve README and add auth examples
- Update SDK client config namespace
- Add support for JWKS URI
- Add OAuth to swagger spec
- Add client secret creation tool

## [0.12.0] - February 14th, 2020
- Add support for typed user attributes
- Add event-stream API for notifications
- Add basic client credentials grant flow
- Add native TLS support via configuration
- Refactor to a POST based action API

## [0.11.0] - January 17th, 2020
- Bump to 1.0.0-rc1@d1b332c of the Optimizely go-sdk
- Improve build tooling
- Add standard metrics registry
- Update documentation
- Return 404 when feature or experiment are not found
- Update event payloads to include Agent name and version

## [0.10.0] - January 9th, 2020
- Rename repo to optimizely/agent and update imports
- Improve CI builds for Windows and SourceClear
- Major /pkg refactoring
- Exclude vulnerabilities identified through SourceClear

## [0.9.0] - January 8th, 2020
- Capture response time metrics in milliseconds
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
- Adds webhook support for multiple concurrent SDK keys

## [0.2.0] - October 3rd, 2019
- Adds Optimizely webhook support
- Adds full Feature MGMT support
- Adds NSQ for UserEvent message transport
- Adds support for multiple concurrent SDK keys

## [0.1.0] - September 4th, 2019
This is the initial release which supported a basic web application and go-sdk integration.

