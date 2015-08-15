package gouter

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"testing"
)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

func LoginHandler(rw http.ResponseWriter, req *http.Request, params *Params) {

	re := regexp.MustCompile("(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)")
	fmt.Println(re.MatchString("Alan Turing"))
	fmt.Printf("%q\n", re.SubexpNames())
	reversed := fmt.Sprintf("${%s} ${%s}", re.SubexpNames()[2], re.SubexpNames()[1])
	fmt.Println(reversed)

	fmt.Fprintf(rw, "<h1>%s</h1>%s", "Get", re.ReplaceAllString("Alan Turing", reversed))
}
func LoginPostHandler(rw http.ResponseWriter, req *http.Request, params *Params) {
	fmt.Fprintf(rw, "<h1>%s</h1>", "Post")
}
func Test_lalala(*testing.T) {

	flag.Parse()
	r := &RegexpHandler{}
	r.HandleFunc("/login", LoginHandler).Method("GET")
	r.HandleFunc("/login", LoginPostHandler).Method("POST")
	http.HandleFunc("/", r.ServeHTTP)
	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}
	http.ListenAndServe(":8080", nil)
}
