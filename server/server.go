package server

import (
	"context"
	"fmt"

	"cli-chat/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
)

// TODO: call room and create chatter etc

// PublicKey wras ssh.PublicKey
type PublicKey struct {
	key ssh.PublicKey
}

func (k PublicKey) String() string {
	return fmt.Sprintf("%s", k.key)
}

type Server struct {
	host       string
	port       int
	wishServer *ssh.Server
	room       *Room
}

func NewServer(keyPath, host string, port int) (*Server, error) {
	finish := make(chan string)
	mainRoom := NewRoom("0", finish)
	r := mainRoom
	s := &Server{
		host: host,
		port: port,
		room: r,
	}

	ws, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		ssh.PublicKeyAuth(publicKeyHandler),
		wish.WithHostKeyPath("./cli-chat"),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
			lm.Middleware(),
			chatMiddleware(s),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to start wish server: %v", err)
	}
	s.wishServer = ws
	return s, nil
}

func publicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

func (s *Server) Start() error {
	return s.wishServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.wishServer.Shutdown(ctx)
}

// attach ssh session info to the model
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		wish.Fatalln(s, "no active terminal, skipping")
		return nil, nil
	}
	m := tui.Model{
		Term:   pty.Term,
		Width:  pty.Window.Width,
		Height: pty.Window.Height,
	}
	m.SetupModel()
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
