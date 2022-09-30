package server

import (
	"fmt"
	"log"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gliderlabs/ssh"
)

var m tea.Model

type NoteMsg string

type Chatter struct {
	room    *Room
	session ssh.Session
	program *tea.Program
	key     PublicKey
	once    sync.Once
}

func (c *Chatter) String() string {
	u := c.session.User()
	return fmt.Sprintf("%s (%s)", u, c.room)
}

func (c *Chatter) Send(m tea.Msg) {
	if c.program != nil {
		c.program.Send(m)
	} else {
		// TODO: include some kind of username
		log.Printf("error sending message to chatter, program is nil")
	}
}

func (c *Chatter) Write(b []byte) (int, error) {
	return c.session.Write(b)
}

func (c *Chatter) WriteString(s string) (int, error) {
	return c.session.Write([]byte(s))
}

func (c *Chatter) Close() error {
	c.once.Do(func() {
		defer delete(c.room.chatters, c.key.String())
		if c.program != nil {
			c.program.Kill()
		}
		c.session.Close()
	})
	return nil
}

// StartChatting starts the Bubble Tea program
func (c *Chatter) StartChatting() {
	_, wchan, _ := c.session.Pty()
	errc := make(chan error, 1)
	go func() {
		select {
		case err := <-errc:
			log.Printf("error starting program %s", err)
		case w := <-wchan:
			if c.program != nil {
				c.program.Send(tea.WindowSizeMsg{Width: w.Width, Height: w.Height})
			}
		case <-c.session.Context().Done():
			c.Close()
		}
	}()

	defer c.room.SendMsg(NoteMsg(fmt.Sprintf("%s left the room", c)))
	var err error
	m, err = c.program.StartReturningModel()
	// TODO: idk what to do with this model yet
	errc <- err
	c.Close()
}
