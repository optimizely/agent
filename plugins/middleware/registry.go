package middleware

import "net/http"

type Creator func() Middleware

var Middlewares = map[string]Creator{}

type Middleware interface {
	Handler() func(http.Handler) http.Handler
}

func Add(name string, creator Creator) {
	Middlewares[name] = creator
}
