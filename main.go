package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/samirkhanal52/go-todo/route"
)

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/", route.HandleIndex)

	srv := &http.Server{
		Addr:         os.Getenv("HOST"),
		Handler:      srvMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go func() {
		fmt.Println("Listening on port...", os.Getenv("HOST"))

		if err := http.ListenAndServe(os.Getenv("HOST"), srvMux); err != nil {
			log.Fatal(err)
		}
	}()

	<-stopChan
	log.Println("Server Shutting Down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	fmt.Println("Server Shut Down")
}
