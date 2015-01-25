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
	"errors"
	"github.com/gophergala/cheppirc/theme"
)

type SessionList struct {
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

	t := template.Must(template.ParseFiles("templates/chat.html"))
	session.Data.RLock()
	t.Execute(w, session.Data)
	session.Data.RUnlock()

	//w.Write([]byte("Hello IRC"))
	//w.Write(data)
}

func (c *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessMsg := fmt.Sprintf("%v %v from %v Headers: %+v", r.Method, r.RequestURI, r.RemoteAddr, r.Header)
	log.Println(accessMsg)

	t := template.Must(template.ParseFiles("templates/login.html"))
	t.Execute(w, "CHEPPIRC")
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
	log.Println("\nUUID:", id.String())
	log.Println("\nCFG:", cfg)

	c.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) { 
			log.Println("Connected to", line.Raw)
			conn.Join(channel)
			session.Data.AddMessage(channel, "", "Now talking on " + channel)
			conn.Who(channel)
		})

	c.HandleFunc("privmsg",
		func(conn *irc.Conn, line *irc.Line) { 
			log.Println("PRIVMSG - Raw:", line.Raw, "Nick:", line.Nick, "Src:", line.Src, "Args:", line.Args, "time:", line.Time)
			session.Data.AddMessage(line.Args[0], line.Nick, line.Args[1])
		})

	c.HandleFunc("352",
		func(conn *irc.Conn, line *irc.Line) {
			log.Println("352 - RAW:", line.Raw)
			session.Data.SetUsers(line.Args[1], line.Args[5], line.Args[3] + " " + line.Args[4])
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