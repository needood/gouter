package gouter

import (
	"net/http"
	"regexp"
	"strings"
)

func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
	methods []string
}
type RegexpHandler struct {
	routes []*route
}
type Params map[interface{}]string

// methodMatcher matches the request against HTTP methods.

func (r *route) Method(methods ...string) {
	r.methods = r.methods[:0]
	for _, v := range methods {
		r.methods = append(r.methods, strings.ToUpper(v))
	}
}

func (h *RegexpHandler) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request, *Params)) *route {
	r := regexp.MustCompile("{(\\w+):([^{}]+(?:\\{\\d*,?\\d*\\})?)+}")
	pattern = r.ReplaceAllString(pattern, "(?P<$1>:$2)")
	reg := regexp.MustCompile("^" + pattern + "$")
	subRoute := &route{reg, http.HandlerFunc(makeHandler(handler, reg)), []string{"GET", "POST"}}
	h.routes = append(h.routes, subRoute)
	return subRoute
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) && matchInArray(route.methods, r.Method) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *Params), reg *regexp.Regexp) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := reg.FindStringSubmatch(r.URL.Path)
		n := reg.SubexpNames()
		params := Params{}
		for i := range m {
			params[i] = m[i]
			if n[i] != "" {
				params[n[i]] = m[i]
			}
		}
		fn(w, r, &params)
	}
}
