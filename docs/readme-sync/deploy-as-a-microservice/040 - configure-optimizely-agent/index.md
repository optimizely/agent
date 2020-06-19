---
title: "Configure Optimizely Agent"
excerpt: ""
slug: "configure-optimizely-agent"
hidden: false
metadata: 
  title: "Configure Agent microservice - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:27.173Z"
updatedAt: "2020-04-08T21:42:08.698Z"
---
By default Optimizely Agent uses the configuration file in the current active directory, e.g.,  `./config.yaml`. You can override the [default configuration](https://github.com/optimizely/agent/blob/master/config.yaml) by providing a yaml configuration file at runtime.

You can specify alternative configuration locations at runtime via an environment variable or command line flag:

```bash
OPTIMIZELY_CONFIG_FILENAME=config.yaml make run
```


Below is a comprehensive list of available configuration properties.

|Property Name|Env Variable|Description|
|---|---|---|
|admin.auth.clients|N/A|Credentials for requesting access tokens. See: [Authorization Guide](doc:authorization)|
|admin.auth.jwksURL|OPTIMIZELY_ADMIN_AUTH_JWKSURL|JWKS URL for validating access tokens. See: [Authorization Guide](doc:authorization)|
|admin.auth.jwksUpdateInterval|OPTIMIZELY_ADMIN_AUTH_JWKSUPDATEINTERVAL|JWKS Update Interval for caching the keys in the background. See: [Authorization Guide](doc:authorization)|
|admin.auth.hmacSecrets|OPTIMIZELY_ADMIN_AUTH_HMACSECRETS|Signing secret for issued access tokens. See: [Authorization Guide](doc:authorization)|
|admin.auth.ttl|OPTIMIZELY_ADMIN_AUTH_TTL|Time-to-live of issued access tokens. See: [Authorization Guide](doc:authorization)|
|admin.port|OPTIMIZELY_ADMIN_PORT|Admin listener port. Default: 8088|
|api.auth.clients|N/A|Credentials for requesting access tokens. See: [Authorization Guide](doc:authorization)|
|api.auth.hmacSecrets|OPTIMIZELY_API_AUTH_HMACSECRETS|Signing secret for issued access tokens. See: [Authorization Guide](doc:authorization)|
|api.auth.jwksURL|OPTIMIZELY_API_AUTH_JWKSURL|JWKS URL for validating access tokens. See: [Authorization Guide](doc:authorization)|
|api.auth.jwksUpdateInterval|OPTIMIZELY_API_AUTH_JWKSUPDATEINTERVAL|JWKS Update Interval for caching the keys in the background. See: [Authorization Guide](doc:authorization)|
|api.auth.ttl|OPTIMIZELY_API_AUTH_TTL|Time-to-live of issued access tokens. See: [Authorization Guide](doc:authorization)|
|api.port|OPTIMIZELY_API_PORT|Api listener port. Default: 8080|
|api.maxConns|OPTIMIZLEY_API_MAXCONNS|Maximum number of concurrent requests|
|author|OPTIMIZELY_AUTHOR|Agent author. Default: Optimizely Inc.|
|certfile|OPTIMIZELY_CERTFILE|Path to a certificate file, used to run Agent with HTTPS|
|client.batchSize|OPTIMIZELY_CLIENT_BATCHSIZE|The number of events in a batch. Default: 10|
|config.filename|OPTIMIZELY_CONFIG_FILENAME|Location of the configuration YAML file. Default: ./config.yaml|
|client.flushInterval|OPTIMIZELY_CLIENT_FLUSHINTERVAL|The maximum time between events being dispatched. Default: 30s|
|client.pollingInterval|OPTIMIZELY_CLIENT_POLLINGINTERVAL|The time between successive polls for updated project configuration. Default: 1m|
|client.queueSize|OPTIMIZELY_CLIENT_QUEUESIZE|The max number of events pending dispatch. Default: 1000|
|disabledCiphers|OPTIMIZELY_DISABLEDCIPHERS|List of TLS ciphers to disable when accepting HTTPS connections|
|keyfile|OPTIMIZELY_KEYFILE|Path to a key file, used to run Agent with HTTPS|
|log.level|OPTIMIZELY_LOG_LEVEL|The log [level](https://github.com/rs/zerolog#leveled-logging) for the agent. Default: info|
|log.pretty|OPTIMIZELY_LOG_PRETTY|Flag used to set colorized console output as opposed to structured json logs. Default: false|
|name|OPTIMIZELY_NAME|Agent name. Default: optimizely|
|version|OPTIMIZELY_VERSION|Agent version. Default: `git describe --tags`|
|sdkKeys|OPTIMIZELY_SDK_KEYS|List of SDK keys used to initialize on startup|
|server.readTimeout|OPTIMIZELY_SERVER_READTIMEOUT|The maximum duration for reading the entire body. Default: “5s”|
|server.writeTimeout|OPTIMIZELY_SERVER_WRITETIMEOUT|The maximum duration before timing out writes of the response. Default: “10s”|
|webhook.port|OPTIMIZELY_WEBHOOK_PORT|Webhook listener port: Default: 8085|
|webhook.projects.<*projectId*>.sdkKeys|N/A|Comma delimited list of SDK keys applicable to the respective projectId|
|webhook.projects.<*projectId*>.secret|N/A|Webhook secret used to validate webhook requests originating from the respective projectId|
|webhook.projects.<*projectId*>.skipSignatureCheck|N/A|Boolean to indicate whether the signature should be validated. TODO remove in favor of empty secret.|