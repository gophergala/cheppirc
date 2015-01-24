package main

import (
	"log"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"net/http"
	"strconv"
)

type Session struct {
	Uuid string
	c *irc.Conn
}

type chatHandler struct {
}
type loginHandler struct {
}

func (c *chatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessMsg := fmt.Sprintf("%v %v from %v Headers: %+v", r.Method, r.RequestURI, r.RemoteAddr, r.Header)
	log.Println(accessMsg)

	r.ParseForm()
	if true {
		http.Redirect(w, r, "login", 302)
	}
	w.Write([]byte("Hello IRC"))
}

func (c *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessMsg := fmt.Sprintf("%v %v from %v Headers: %+v", r.Method, r.RequestURI, r.RemoteAddr, r.Header)
	log.Println(accessMsg)

	w.Write([]byte("Hello LOGIN"))
}

func newChatHandler() *chatHandler {
	c := &chatHandler{}
	return c
}

func newLoginHandler() *loginHandler {
	c := &loginHandler{}
	return c
}

func main() {
	log.Println("Starting up server...")
	mux := http.NewServeMux()
	mux.Handle("/chat", newChatHandler())
	mux.Handle("/login", newLoginHandler())

	log.Println("Listening...")
	http.ListenAndServe(":"+strconv.Itoa(8081), mux)
}