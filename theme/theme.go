package theme

import (
	"github.com/gophergala/cheppirc/message"
	"log"
)

type ThemeData struct {
	Messages map[string][]message.Message
}

func (d *ThemeData) AddMessage(target, sender, text string) {
	m := message.Message{sender, target}
	if _, ok := d.Messages[target]; !ok {
		log.Println("Target not found. Args:", target)
		d.Messages[target] = []message.Message{}
	}
	d.Messages[target] = append(d.Messages[target], m)
}

