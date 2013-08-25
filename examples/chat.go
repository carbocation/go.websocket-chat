package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/carbocation/go.websocket-chat"
	"github.com/garyburd/go-websocket/websocket"
	"github.com/gorilla/mux"
)

func init() {
	wshub.Initialize(10*time.Second,
		60*time.Second,
		60*time.Second*9/10,
		4096,
		256)
}

var r *mux.Router
var addr = flag.String("addr", ":9997", "http service address")
var homeTempl = template.Must(template.New("home.html").Parse(mainTemplate))

func homeHandler(w http.ResponseWriter, req *http.Request) {
	path, _ := r.Get("chat").URLPath("id", "Room 1")
	http.Redirect(w, req, path.String(), http.StatusFound)
}

func chatHandler(w http.ResponseWriter, req *http.Request) {
	data := struct {
		Host string
		ID   string
	}{
		Host: req.Host,
		ID:   mux.Vars(req)["id"],
	}

	homeTempl.Execute(w, data)
}

func injectorHandler(w http.ResponseWriter, req *http.Request) {
	go func() {
		msg := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
		for i := 0; i < 20; i++ {
			msg = append(msg, []byte("abcdefghijklmnopqrstuvwxyz0123456789")...)
		}
		wshub.BroadcastAll(msg)
	}()
}

func wsHandler(w http.ResponseWriter, req *http.Request) {
	ws, err := websocket.Upgrade(w, req.Header, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	//When we try to handle this, see if the hub exists.
	id := mux.Vars(req)["id"]
	wshub.Launch(ws, id)
}

func main() {
	log.Println("Launched")
	flag.Parse()

	r = mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/injector", injectorHandler)
	r.HandleFunc(`/{id:[_!.,+\- a-zA-Z0-9]+}`, chatHandler).Name("chat")
	r.HandleFunc(`/ws/{id:[_!.,+\- a-zA-Z0-9]+}`, wsHandler)
	r.HandleFunc("/loaderio-766fe6ab96dba175477e09ae8baf291c/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "loaderio-766fe6ab96dba175477e09ae8baf291c")
	})

	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
