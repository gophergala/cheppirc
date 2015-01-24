package main

import (
	"log"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type SessionList struct {
	sync.RWMutex
	Sessions map[string]Session
}

type Session struct {
	Uuid string
	c *irc.Conn
	data *ThemeData
}

type chatHandler struct {
	sessionList *SessionList
}

type loginHandler struct {
}

type connectHandler struct {
	sessionList *SessionList
}

type ThemeData struct {
	Messages map[string]Message
}

type Message struct {
	Text string
}

func (c *chatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessMsg := fmt.Sprintf("%v %v from %v Headers: %+v", r.Method, r.RequestURI, r.RemoteAddr, r.Header)
	log.Println(accessMsg)

	r.ParseForm()
	session := getSession(r.Form, c.sessionList)
	//TODO: validate if uuid exists
	if session == nil {
		http.Redirect(w, r, "login", 302)
		return
	}

	data := displayChat(session)

	//w.Write([]byte("Hello IRC"))
	w.Write(data)
}

func (c *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessMsg := fmt.Sprintf("%v %v from %v Headers: %+v", r.Method, r.RequestURI, r.RemoteAddr, r.Header)
	log.Println(accessMsg)

	t := template.Must(template.ParseFiles("templates/login.html"))
	t.Execute(w, "CHEPPIRC")

	w.Write([]byte("Hello LOGIN"))
}

func (c *connectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Println(r.Form)
}

func newChatHandler(s *SessionList) *chatHandler {
	c := &chatHandler{s}
	return c
}

func newLoginHandler() *loginHandler {
	c := &loginHandler{}
	return c
}

func newConnectHandler(s *SessionList) *connectHandler {
	c := &connectHandler{s}
	return c
}

func getSession(values url.Values, sessionList *SessionList) *Session {
	uuid := values.Get("session")
	if len(uuid) < 1 {
		return nil
	}

	if s, ok := sessionList.Sessions[uuid]; ok {
		return &s
	}

	return nil
}

func displayChat(s *Session) []byte {
	var output string
	output = "uuid:" + s.Uuid + "\n **** \n"
	for m := range s.data.Messages {
		output = output + "\n---\n" + m
	}

	return []byte(output)
}

func main() {
	log.Println("Starting up server...")
	sessionList := new(SessionList)
	sessionList.Sessions = make(map[string]Session)

	mux := http.NewServeMux()
	mux.Handle("/chat", newChatHandler(sessionList))
	mux.Handle("/login", newLoginHandler())
	mux.Handle("/connect", newConnectHandler(sessionList))

	log.Println("Listening...")

	http.ListenAndServe(":"+strconv.Itoa(8081), mux)
}