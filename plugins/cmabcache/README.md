# CMAB Cache

Use a CMAB Cache to cache Contextual Multi-Armed Bandit (CMAB) decisions fetched from the CMAB prediction service.

## Out of Box Cache Usage

1. To use the in-memory `CMABCache`, update the `config.yaml` as shown below:

```yaml
## configure optional CMAB Cache
client:
  cmab:
    ## If no cache is defined (or no default is defined), we will use the default in-memory with default size and timeout
    cache:
      default: "in-memory"
      services:
        in-memory:
          ## maximum number of entries for in-memory cache
          size: 10000
          ## timeout after which the cached item will become invalid.
          timeout: 30m
```

2. To use the redis `CMABCache`, update the `config.yaml` as shown below:

```yaml
## configure optional CMAB Cache
client:
  cmab:
    ## If no cache is defined (or no default is defined), we will use the default in-memory with default size and timeout
    cache:
      default: "redis"
      services:
        redis:
          host: "your_host"
          password: "your_password"
          database: 0 ## your database
          timeout: 30m
```

## Custom CMABCache Implementation

To implement a custom CMAB cache, the following steps need to be taken:

1. Create a struct that implements the `cache.Cache` interface in `plugins/cmabcache/services`.

2. Add an `init` method inside your CMABCache file as shown below:

```go
func init() {
	myCacheCreator := func() cache.Cache {
		return &yourCacheStruct{
		}
	}
	cmabcache.Add("my_cache_name", myCacheCreator)
}
```

3. Update the `config.yaml` file with your `CMABCache` config as shown below:

```yaml
## configure optional CMABCache
client:
  cmab:
    cache:
      default: "my_cache_name"
      services:
        my_cache_name:
          ## Add those parameters here that need to be mapped to the CMABCache
          ## For example, if the CMAB cache struct has a json mappable property called `host`
          ## it can be updated with value `abc.com` as shown
          host: "abc.com"
```

- If a user has created multiple `CMABCache` services and wants to override the `default` `CMABCache` for a specific `sdkKey`, they can do so by providing the `CMABCache` name in the request Header `X-Optimizely-CMAB-Cache-Name`.

- Whenever a request is made with a unique `sdkKey`, the agent node handling that request creates and caches a new `CMABCache`. To keep the `CMABCache` type consistent among all nodes in a cluster, it is recommended to send the request Header `X-Optimizely-CMAB-Cache-Name` in every request made.
