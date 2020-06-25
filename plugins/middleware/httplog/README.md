## HTTPLog Middleware Plugin

The HTTPLog plugin provides HTTP request logging for Agent through the use of the [go-chi/httplog](https://github.com/go-chi/httplog) middleware.

### Example Request Log
```json
{
  "level": "info",
  "httpRequest": {
    "header": {
      "accept": "*/*",
      "user-agent": "curl/7.64.1"
    },
    "proto": "HTTP/1.1",
    "remoteIP": "[::1]:61291",
    "requestMethod": "GET",
    "requestPath": "/config",
    "requestURL": "http://localhost:8088/config",
    "scheme": "http"
  },
  "time": "2020-06-29T16:14:32-07:00",
  "message": "Request: GET /config"
}
```

### Example Response Log
```json
{
  "level": "info",
  "httpRequest": {
    "proto": "HTTP/1.1",
    "remoteIP": "[::1]:61291",
    "requestMethod": "GET",
    "requestPath": "/config",
    "requestURL": "http://localhost:8088/config"
  },
  "httpResponse": {
    "bytes": 854,
    "elapsed": 0.371279,
    "header": {
      "app-name": "optimizely",
      "app-version": "v1.2.0-10-gc7dd5bc",
      "author": "Optimizely Inc.",
      "content-type": "application/json; charset=utf-8"
    },
    "status": 200
  },
  "time": "2020-06-29T16:14:32-07:00",
  "message": "Response: 200 OK"
}
```
