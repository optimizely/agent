# User Profile Service
Use a User Profile Service to persist information about your users and ensure variation assignments are sticky.  
The User Profile Service implementation you provide will override Optimizely's default bucketing behavior in cases  
when an experiment assignment has been saved.

When implementing in a multi-server or stateless environment, we suggest using this interface with a backend like  
Cassandra or Redis. You can decide how long you want to keep your sticky bucketing around by configuring these services.

Implementing a User Profile Service is optional and is only necessary if you want to keep variation assignments sticky  
even when experiment conditions are changed while it is running (for example, audiences, attributes,  
variation pausing, and traffic distribution). Otherwise, the agent is stateless and relies on deterministic bucketing to return  
consistent assignments.

## UserProfileService EndPoints

1. To lookup a user profile, use agent's `POST /v1/lookup`:

```
In the request `application/json` body, include the `userId`. The full request looks like this:

```curl
curl --location --request POST 'http://localhost:8080/v1/lookup' \
--header 'X-Optimizely-SDK-Key: YOUR_SDK_KEY' \
--header 'Accept: text/event-stream' \
--header 'Content-Type: application/json' \
--data-raw '{
  "userId": "string"
}'
```

2. To save a user profile, use agent's `POST /v1/save`:

```
In the request `application/json` body, include the `userId` and `experimentBucketMap`. The full request looks like this:

```curl
curl --location --request POST 'http://localhost:8080/v1/save' \
--header 'X-Optimizely-SDK-Key: YOUR_SDK_KEY' \
--header 'Accept: text/event-stream' \
--header 'Content-Type: application/json' \
--data-raw '{
    "userId": "string",
    "experimentBucketMap": {
        "experiment_id_to_save": {
            "variation_id": "variation_id_to_save"
        }
    }
}'
```

## Out of Box UserProfileService Usage

1. To use the in-memory `UserProfileService`, update the `config.yaml` as shown below:
```
## configure optional User profile service
userProfileService:
      default: "in-memory"
      services:
        in-memory: 
          ## 0 means no limit on capacity
          capacity: 0
          ## supports lifo/fifo
          storageStrategy: "fifo"
```

2. To use the redis `UserProfileService`, update the `config.yaml` as shown below:
```
## configure optional User profile service
userProfileService:
      default: "redis"
      services:
        redis: 
          host: "your_host"
          password: "your_password"
          database: 0 ## your database
```

3. To use the rest `UserProfileService`, update the `config.yaml` as shown below:
```
## configure optional User profile service
userProfileService:
      default: "rest"
      services:
        rest:
          host: "your_host"
          lookupPath: "/lookup_endpoint"
          lookupMethod: "POST"
          savePath: "/save_endpoint"
          saveMethod: "POST"
          userIDKey: "user_id"
          headers: 
            "header_key": "header_value"
```

Implement 2 api's `/lookup_endpoint` and `/save_endpoint` on your `host`. Api methods will be `POST` by default but can be  
updated through `lookupMethod` and `saveMethod` properties. Similarly, request parameter key for `user_id` can also be updated  
using `userIDKey` property.
    
- `lookup_endpoint` should accept `user_id` in its json body or query (depending upon the method type) and if successful,   
return the status code `200` with json response (keep in mind that when sending response,   
`user_id` should be substituted with value of `userIDKey` from config.yaml):   

```
{
  "experiment_bucket_map": {
    "saved_experiment_id": {
      "variation_id": "saved_variation_id"
    }
  },
  "user_id": "saved_user_id"
}
```
- `save_endpoint` should accept the following parameters in its json body or query (depending upon the method type)   
and return the status code `200` if successful (keep in mind that `user_id` should be substituted   
with the value of`userIDKey` from config.yaml):  

```
{
  "experiment_bucket_map": {
    "experiment_id_to_save": {
      "variation_id": "variation_id_to_save"
    }
  },
  "user_id": "user_id_to_save"
}
```

## Custom UserProfileService Implementation

To implement a custom user profile service, followings steps need to be taken:
1. Create a struct that implements the `decision.UserProfileService` interface in `plugins/userprofileservice/services`.
2. Add a `init` method inside your UserProfileService file as shown below:
```
func init() {
	myUPSCreator := func() decision.UserProfileService {
		return &yourUserProfileServiceStruct{
		}
	}
	userprofileservice.Add("my_ups_name", myUPSCreator)
}
```
3. Update the `config.yaml` file with your `UserProfileService` config as shown below:

```
## configure optional User profile service
userProfileServices:
    default: "my_ups_name"
    services:
        my_ups_name: 
           ## Add those parameters here that need to be mapped to the UserProfileService
           ## For example, if the UPS struct has a json mappable property called `host`
           ## it can updated with value `abc.com` as shown
           host: “abc.com”
```
- If a user has created multiple `UserProfileServices` and wants to override the `default` `UserProfileService` for a specifc `sdkKey`, they can do so by providing the `UserProfileService` name in the request Header `X-Optimizely-UPS-Name`.
