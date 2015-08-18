package gouter

import (
	"errors"
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

type HandlerFunc func(http.ResponseWriter, *http.Request, *Params)
type route struct {
	pattern *regexp.Regexp
	handler http.Handler
	methods []string
}

var GouterHandler = &RegexpHandler{}

func initRout(reg *regexp.Regexp, handler http.HandlerFunc) (*route, error) {
	return &route{reg, http.HandlerFunc(handler), []string{"GET", "POST"}}, nil
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) appendRoute(reg *regexp.Regexp, handler HandlerFunc) *route {
	handlerFunc := makeHandler(handler, reg, &h.routes)
	subRoute, _ := initRout(reg, handlerFunc)
	h.routes = append(h.routes, subRoute)
	return subRoute
}

type Params struct {
	paramsInt    []string
	paramsString map[string]int
	flags        map[string]int
	routes       *[]*route
}

func InitParam() *Params {
	params := new(Params)
	params.flags = make(map[string]int)
	params.paramsString = make(map[string]int)
	return params
}
func (p *Params) SetParam(m, n []string) error {
	if len(m) != len(n) {
		return errors.New("params's length is not equal")
	}
	p.paramsInt = m
	p.paramsString = make(map[string]int)
	for i := range n {
		if n[i] != "" {
			p.paramsString[n[i]] = i
		}
	}
	return nil
}
func (p *Params) SetByIndex(index int, value string) {
	p.paramsInt[index] = value
}
func (p *Params) Set(key, value string) {
	p.paramsInt[p.paramsString[key]] = value
}
func (p *Params) Get(key string) string {
	return p.paramsInt[p.paramsString[key]]
}
func (p *Params) GetByIndex(index int) string {
	return p.paramsInt[index]
}
func (p *Params) SetFlag(key string, value int) {
	p.flags[key] = value
}
func (p *Params) GetFlag(key string) int {
	return p.flags[key]
}
func (p *Params) Next(w http.ResponseWriter, r *http.Request) {
	routes := *p.routes
	for _, route := range routes[p.GetFlag("next"):] {
		if route.pattern.MatchString(r.URL.Path) && matchInArray(route.methods, r.Method) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

// methodMatcher matches the request against HTTP methods.

func (r *route) Method(methods ...string) {
	r.methods = r.methods[:0]
	for _, v := range methods {
		r.methods = append(r.methods, strings.ToUpper(v))
	}
}

func (h *RegexpHandler) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request, *Params)) *route {
	r := regexp.MustCompile("{([a-zA-Z]+\\d*):([^{}]*(?:\\\\\\{)?(?:\\\\\\{)?(?:p\\{\\w+\\})?(?:\\{\\w*,?\\d*\\})?)+}")
	pattern = r.ReplaceAllString(pattern, "(?P<$1>$2)")
	r2 := regexp.MustCompile("{([a-zA-Z]+\\d*)}")
	pattern = r2.ReplaceAllString(pattern, "(?P<$1>[^/]+)")
	reg := regexp.MustCompile("^" + pattern + "$")
	subRoute := h.appendRoute(reg, handler)
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

func makeHandler(fn HandlerFunc, reg *regexp.Regexp, routes *[]*route) http.HandlerFunc {

	params := InitParam()
	params.SetFlag("next", len(*routes)+1)
	params.routes = routes

	return func(w http.ResponseWriter, r *http.Request) {
		m := reg.FindStringSubmatch(r.URL.Path)
		n := reg.SubexpNames()
		params.SetParam(m, n)
		fn(w, r, params)
	}
}
