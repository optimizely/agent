# ODP Cache
Use a ODP Cache to persist segments and ensure they are sticky. 

## Out of Box Cache Usage

1. To use the in-memory `ODPCache`, update the `config.yaml` as shown below:
```
## configure optional ODP Cache
odpCache:
  default: "in-memory"
  services:
    in-memory: 
      ## 0 means no limit on capacity
      capacity: 0
      ## supports lifo/fifo
      storageStrategy: "fifo"
```

2. To use the redis `ODPCache`, update the `config.yaml` as shown below:
```
## configure optional User profile service
odpCache:
  default: "redis"
  services:
    redis: 
      host: "your_host"
      password: "your_password"
      database: 0 ## your database
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
odpCache:
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
