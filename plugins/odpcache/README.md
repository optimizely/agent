# ODP Cache
Use a ODP Cache to cache user segments fetched from the ODP server.

## Out of Box Cache Usage

1. To use the in-memory `ODPCache`, update the `config.yaml` as shown below:
```
## configure optional ODP Cache
odp:
  ## If no segmentsCache is defined (or no default is defined), we will use the default in-memory with default size and timeout
  segmentsCache:
    default: "in-memory"
    services:
      in-memory: 
        ## 0 means cache will be disabled
        size: 0
        ## timeout after which the cached item will become invalid.
        ## 0 means records will never be deleted
        timeout: 0s
```

2. To use the redis `ODPCache`, update the `config.yaml` as shown below:
```
## configure optional ODP Cache
odp:
  ## If no segmentsCache is defined (or no default is defined), we will use the default in-memory with default size and timeout
  segmentsCache:
    default: "redis"
    services:
      redis: 
        host: "your_host"
        password: "your_password"
        database: 0 ## your database
        timeout: 0s
```

## Custom ODPCache Implementation

To implement a custom odp cache, followings steps need to be taken:
1. Create a struct that implements the `cache.Cache` interface in `plugins/odpcache/services`.
2. Add a `init` method inside your ODPCache file as shown below:
```
func init() {
	myCacheCreator := func() cache.Cache {
		return &yourCacheStruct{
		}
	}
	odpcache.Add("my_cache_name", myCacheCreator)
}
```
3. Update the `config.yaml` file with your `ODPCache` config as shown below:

```
## configure optional ODPCache
odp
  segmentsCache:
    default: "my_cache_name"
    services:
      my_cache_name: 
        ## Add those parameters here that need to be mapped to the ODPCache
        ## For example, if the ODP struct has a json mappable property called `host`
        ## it can updated with value `abc.com` as shown
        host: “abc.com”
```
- If a user has created multiple `ODPCache` services and wants to override the `default` `ODPCache` for a specific `sdkKey`, they can do so by providing the `ODPCache` name in the request Header `X-Optimizely-ODP-Cache-Name`.

- Whenever a request is made with a unique `sdkKey`, The agent node handling that request creates and caches a new `ODPCache`. To keep the `ODPCache` type consistent among all nodes in a cluster, it is recommended to send the request Header `X-Optimizely-ODP-Cache-Name` in every request made.
