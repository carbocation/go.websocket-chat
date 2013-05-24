package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"code.google.com/p/go.net/websocket"
)

//Main hub lives here
//Todo: turn this into a map with a mutex
var h *hub

var addr = flag.String("addr", ":8080", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))

func homeHandler(c http.ResponseWriter, req *http.Request) {
	homeTempl.Execute(c, req.Host)
}

func injectorHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Thanks for sending your message.")

	go func() {
		h.broadcast <- "A third party injected this message for fun or for profit."
	}()
}

func main() {
	fmt.Println("Launched")
	flag.Parse()
	http.HandleFunc("/", homeHandler)
	http.Handle("/ws", websocket.Handler(wsHandler))
	http.HandleFunc("/injector", injectorHandler)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
