---
title: "Webhooks"
excerpt: ""
slug: "webhooks-agent"
hidden: false
metadata: 
  title: "Agent microservice webhooks - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:26.981Z"
updatedAt: "2020-05-05T17:03:48.045Z"
---
Optimizely Agent implements a  webhook listener used to receive inbound [Webhook](doc:configure-webhooks) requests from optimizely.com. These webhooks enable PUSH style notifications triggering immediate project configuration updates.
The webhook listener is configured on its own port (default: 8085) since it can be configured to select traffic from the internet.

To accept webhook requests Agent must be configured by mapping an Optimizely Project Id to a set of SDK keys along
with the associated secret used for validating the inbound request. An example webhook configuration can be seen below, while the full example configuration can be found in the the provided [config.yaml](https://github.com/optimizely/agent/blob/master/config.yaml#L58).

```yaml
##
## webhook service receives update notifications to your Optimizely project. Receipt of the webhook will
## trigger an immediate download of the datafile from the CDN
##
webhook:
    ## http listener port
    port: "8089"
#    ## a map of Optimizely Projects to one or more SDK keys
#    projects:
#        ## <project-id>: Optimizely project id as an integer
#        <project-id>:
#            ## sdkKeys: a list of SDKs linked to this project
#            sdkKeys:
#                - <sdk-key-1>
#                - <sdk-key-1>
#            ## secret: webhook secret used the validate the notification
#            secret: <secret-10000>
#            ## skipSignatureCheck: override the signature check (not recommended for production)
#            skipSignatureCheck: true
```