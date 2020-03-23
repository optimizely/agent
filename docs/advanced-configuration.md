# Advanced Configuration Notes

## Setting Configuration Values

Configuration can be provided to Agent via the following methods:
1. Reading from environment variables
2. Reading from a YAML configuration file

When a configuration option is specified through both methods, environment variables take precedence over the config file.

Internally, Optimizely Agent uses the [Viper](https://github.com/spf13/viper) library's support for configuration files and environment variables.

## Config File Location

The default location of the config file is `config.yaml` in the root directory. If you want to specify another location, use the `OPTIMIZELY_CONFIG_FILENAME` environment variable:
```bash
OPTIMIZELY_CONFIG_FILENAME=/path/to/other_config_file.yaml make run
```

## Nested Configuration Options
When setting the value of "nested" configuration options using environment variables, underscores denote deeper access. The following examples are equivalent ways of setting the client polling interval:
```yaml
# Setting a nested value in a .yaml file:
client:
    pollingInterval: 120s
```

```shell script
// Set environment variable for pollingInterval, nested inside client
export OPTIMIZELY_CLIENT_POLLINGINTERVAL=120s
```

## Unsupported Environment Variable Options
Some options can only be set via config file, and not environment variable (for details on these options, see the Configuration Options table in the [main README](../README.md)):
- `admin.auth.clients`
- `api.auth.clients`
- Options under`webhook.projects`
