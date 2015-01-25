package theme

import (
	"github.com/gophergala/cheppirc/message"
	"github.com/gophergala/cheppirc/user"
	"log"
)

type ThemeData struct {
	Messages map[string][]message.Message
	Users map[string][]user.User
}

func (d *ThemeData) AddMessage(target, sender, text string) {
	m := message.Message{sender, text}
	if _, ok := d.Messages[target]; !ok {
		log.Println("Target not found. Args:", target)
		d.Messages[target] = []message.Message{}
	}
	d.Messages[target] = append(d.Messages[target], m)
}

func NewThemeData() *ThemeData {
	d := &ThemeData{}
	d.Messages = make(map[string][]message.Message)
	d.Users = make(map[string][]user.User)
	return d
}

