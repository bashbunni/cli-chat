package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cli-chat/server"
)

const (
	// TODO: eventually outgrow mock data
	host = "0.0.0.0"
	port = 23235
)

func main() {
	var key string

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)

	k := os.Getenv("CHAT_SERVER_KEY_PATH")
	if k != "" {
		log.Println("k not empty")
		key = k
	}
	//	h := os.Getenv("CHAT_SERVER_HOST")
	//	if h != "" {
	//		host = h
	//	}
	//	p := os.Getenv("CHAT_SERVER_PORT")
	//	if p != "" {
	//		port, _ = strconv.Atoi(p)
	//	}

	s, err := server.NewServer(key, host, port)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		//  start server until done?
		if err := s.Start(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
