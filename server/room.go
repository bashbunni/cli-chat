package server

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gliderlabs/ssh"
)

var idleTimeout = time.Minute * 3

type Room struct {
	id       string
	chatters map[string]*Chatter
	sync     chan tea.Msg
	done     chan struct{}
	finish   chan string
}

// String implements the Stringer interface
func (r *Room) String() string {
	return r.id
}

func NewRoom(id string, finish chan string) *Room {
	s := make(chan tea.Msg)
	r := &Room{
		id:       id,
		chatters: make(map[string]*Chatter, 0),
		sync:     s,
		done:     make(chan struct{}, 1),
		finish:   finish,
	}
	go func() {
		r.Listen()
	}()

	return r
}

func (r *Room) Close() {
	log.Printf("closing room %s", r)
	for _, c := range r.chatters {
		n, _ := c.WriteString("Idle timeout.\n")
		fmt.Println(n)
		c.Close()
	}

	r.done <- struct{}{}
	r.finish <- r.id
	close(r.sync)
	close(r.done)
}

// Listen listens for messages from players in the room and other events
func (r *Room) Listen() {
	for {
		select {
		case <-r.done:
			return
		case <-time.After(idleTimeout):
			log.Printf("idle timeout for room %s", r)
			r.Close()
		case m := <-r.sync:
			switch msg := m.(type) {
			// TODO: handle messages
			case tea.WindowSizeMsg:
				log.Println(msg.Height)
			}
		}
	}
}

// SendMsg sends a tea.Msg to all chatters in a room
func (r *Room) SendMsg(m tea.Msg) {
	go func() {
		for _, c := range r.chatters {
			c.Send(m)
		}
	}()
}

// NewChatter creates a new chatter for the given session
func (r *Room) NewChatter(s ssh.Session) *Chatter {
	chatter := &Chatter{
		room:    r,
		session: s,
		key:     PublicKey{key: s.PublicKey()},
	}
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithInput(s),
		tea.WithOutput(s),
	)

	chatter.program = p
	return chatter
}

func (r *Room) AddChatter(s ssh.Session) (*Chatter, error) {
	k := s.PublicKey()
	if k == nil {
		return nil, fmt.Errorf("no public key")
	}
	key := PublicKey{key: k}
	p, ok := r.chatters[key.String()]
	if ok {
		return nil, fmt.Errorf("Chatter %s is already in the room", p)
	}
	return r.NewChatter(s), nil
}
