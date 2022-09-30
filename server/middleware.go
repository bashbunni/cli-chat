package server

import (
	"log"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
	"github.com/muesli/termenv"
)

func chatMiddleware(srv *Server) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		lipgloss.SetColorProfile(termenv.ANSI256)

		return func(s ssh.Session) {
			_, _, active := s.Pty()
			if !active {
				s.Write([]byte("No TTY"))
				s.Exit(1)
				return
			}
			c, err := srv.room.AddChatter(s)
			if err != nil {
				s.Write([]byte([]byte(err.Error() + "\n")))
				s.Exit(1)
				return
			}
			log.Printf("%s joined the room", s.User())
			c.StartChatting()
			log.Printf("%s left the room", s.User())
			sh(s)
		}
	}
}
