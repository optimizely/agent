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

1. To use the redis `UserProfileService`, update the `config.yaml` as shown below:
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
