# Notifications Guide

Agent provides an endpoint that sends notifications to subscribers via [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events). This is Agent's equivalent of Notification Listeners found in Optimizely SDKs.

For details on the notification types, what causes them to be triggered, and the data they provide, see the [Notification Listeners documentation](https://docs.developers.optimizely.com/full-stack/docs/set-up-notification-listener-go).

## Configuration

By default, the notifications endpoint is disabled. To enable it, change config.yaml:
```yaml
# config.yaml
api:
    enableNotifications: true
```
Or, enable it by setting an environment variable:
```shell script
export OPTIMIZELY_API_ENABLENOTIFICATIONS=1
```

## Usage
Send a `GET` request to `/v1/notifications/event-stream` to subscribe:
```shell script
curl -N -H "Accept:text/event-stream" -H "X-Optimizely-Sdk-Key:9LCprAQyd1bs1BBXZ3nVji"\
  http://localhost:8080/v1/notifications/event-stream
```
This connection will remain open, and any notifications triggered by other requests received by Agent are pushed as events to this stream. Try sending requests to `/v1/activate` or `/v1/track` to see notifications being triggered.


### Filtering
To subscribe only to a particular category of notifications, add a `filter` query parameter. For example, to subscribe only to Decision notifications:
```shell script
curl -N -H "Accept:text/event-stream" -H "X-Optimizely-Sdk-Key:9LCprAQyd1bs1BBXZ3nVji"\
  http://localhost:8080/v1/notifications/event-stream?filter=decision
```


## Example
A runnable Python example can be found in [`examples/notifications.py`](../examples/notifications.py).
