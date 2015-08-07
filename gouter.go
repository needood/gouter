package gouter

import (
    "net/http"
    "regexp"
    "log"
)


type route struct {
    pattern *regexp.Regexp
    handler http.Handler
}

type RegexpHandler struct {
    routes []*route
}

func (h *RegexpHandler) Handler(pattern string, handler http.Handler) {
    h.routes = append(h.routes, &route{regexp.MustCompile(pattern), handler})
}

func (h *RegexpHandler) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request, []string)) {
    reg := regexp.MustCompile("^"+pattern+"$")
    h.routes = append(h.routes, &route{reg, http.HandlerFunc(makeHandler(handler,reg))})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    for _, route := range h.routes {
        if route.pattern.MatchString(r.URL.Path) {
            route.handler.ServeHTTP(w, r)
            return
        }
    }
    // no pattern matched; send 404 response
    http.NotFound(w, r)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, []string),reg *regexp.Regexp) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := reg.FindStringSubmatch(r.URL.Path)
        log.Print(r.URL.Path)
        fn(w, r, m)
    }
}
