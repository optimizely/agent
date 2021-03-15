---
title:  Agent Notifications
excerpt: ""
slug: "agent-notifications"
hidden: false
metadata: 
  title: "Agent notifications - Optimizely Full Stack"
createdAt: "2020-05-21T20:35:58.387Z"
updatedAt: "2021-03-15T23:02:34.056Z"
---

Agent provides an endpoint that sends notifications to subscribers via [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events). This is Agent's equivalent of Notification Listeners found in Optimizely SDKs.

For details on the notification types, what causes them to be triggered, and the data they provide, see the [Notification Listeners documentation](https://docs.developers.optimizely.com/full-stack/docs/set-up-notification-listener-go).

## Configuration

By default, the notifications endpoint is disabled. To enable it, change config.yaml:

```
# config.yaml
api:
    enableNotifications: true
```

Or, enable it by setting an environment variable:

```
# set an env. variable
export OPTIMIZELY_API_ENABLENOTIFICATIONS=1
```

## Usage

Send a `GET` request to `/v1/notifications/event-stream` to subscribe:

```
curl -N -H "Accept:text/event-stream" -H "X-Optimizely-Sdk-Key:<YOUR SDK KEY>"\
  http://localhost:8080/v1/notifications/event-stream
```

This connection will remain open, and any notifications triggered by other requests received by Agent are pushed as events to this stream. Try sending requests to `/v1/activate` or `/v1/track` to see notifications being triggered.


### Filtering

To subscribe only to a particular category of notifications, add a `filter` query parameter. For example, to subscribe only to Decision notifications:

```
# filter on decision notifications
curl -N -H "Accept:text/event-stream" -H "X-Optimizely-Sdk-Key:<YOUR SDK KEY>"\
  http://localhost:8080/v1/notifications/event-stream?filter=decision
```

## Example

For a runnable Python example, see [examples/notifications.py](https://github.com/optimizely/agent/tree/master/examples).
