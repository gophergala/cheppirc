package theme

import (
	"encoding/json"
	"github.com/gophergala/cheppirc/message"
	"github.com/gophergala/cheppirc/user"
	"log"
	"sync"
)

type ThemeData struct {
	Messages map[string][]message.Message
	Users map[string]map[string]user.User
	Uuid string
	sync.RWMutex
}

func (d *ThemeData) AddMessage(target, sender, text string, updater chan []byte) {
	log.Println("ADDMESSAGE:", text, "DEBUG USERS:", d.Users)
	d.Lock()
	m := message.Message{sender, text}
	if _, ok := d.Messages[target]; !ok {
		log.Println("ADDMESSAGE: Target not found. Target:", target)
		d.Messages[target] = []message.Message{}
	}
	d.Messages[target] = append(d.Messages[target], m)
	d.Unlock()
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("Error marshalling message:", err.Error())
	}
	updater <- b
}

func (d *ThemeData) SetUsers(target, nick, info string) {
	d.Lock()
	if _, ok := d.Users[target]; !ok {
		log.Println("SETUSERS: Target not found. Target:", target)
		d.Users[target] = make(map[string]user.User)
	}
	d.Users[target][nick] = user.User{nick, info}
	d.Unlock()
}

func NewThemeData() *ThemeData {
	d := &ThemeData{}
	d.Messages = make(map[string][]message.Message)
	d.Users = make(map[string]map[string]user.User)
	return d
}

