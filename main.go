package main

import (
	"log"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/nu7hatch/gouuid"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"errors"
	"github.com/gophergala/cheppirc/theme"
)

type SessionList struct {
	sync.RWMutex
	Sessions map[string]Session
}

type Session struct {
	Uuid string
	C *irc.Conn
	Data *theme.ThemeData
}

type chatHandler struct {
	sessionList *SessionList
}

type loginHandler struct {
}

type connectHandler struct {
	sessionList *SessionList
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
	nick := r.Form.Get("nick")
	channel := r.Form.Get("channel")
	server := r.Form.Get("server")
	port := r.Form.Get("port")

	session, err := newSession(nick, channel, server, port)
	if err != nil {
		w.Write([]byte("{\"success\": false, \"message\": \"" + err.Error() + "\"}"))
	}

	c.sessionList.Sessions[session.Uuid] = *session

	w.Write([]byte("{\"success\": true, \"message\": \"" + session.Uuid + "\"}"))
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

func newSession(nick, channel, server, port string) (*Session, error) {
	cfg := irc.NewConfig(nick)
	cfg.SSL = false
	cfg.Server = server + ":" + port
	cfg.NewNick = func(n string) string { return n + "^" }
	c := irc.Client(cfg)

	log.Println(c.String())
	id, _ := uuid.NewV4()
	session := &Session{id.String(), c, nil}
	session.Data = theme.NewThemeData()
	log.Println("UUID:", id.String())

	c.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) { 
			conn.Join(channel)
			session.Data.AddMessage(channel, "", "Now talking on " + channel)
		})

	c.HandleFunc("privmsg",
		func(conn *irc.Conn, line *irc.Line) { 
			log.Println("Raw:", line.Raw, "Nick:", line.Nick, "Src:", line.Src, "Args:", line.Args, "time:", line.Time) 
			session.Data.AddMessage(line.Args[0], line.Nick, line.Args[1])
		})


	if err := c.Connect(); err != nil {
		return nil, errors.New("Connection error: " + err.Error())
	}

	return session, nil
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
	log.Println("Messages:", s.Data.Messages)
	for key, m := range s.Data.Messages {
		output = output + "\n---\n" + "key: " + key + " m0=" + m[0].Text
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