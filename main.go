package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/carbocation/gotogether"
	"github.com/garyburd/go-websocket/websocket"
	"github.com/gorilla/mux"
)

//Main hub lives here
var r *mux.Router
var hubs *hubMap = &hubMap{m: map[string]*hub{}}

var addr = flag.String("addr", ":9997", "http service address")
var homeTempl = template.Must(gotogether.LoadTemplates(template.New("home.html"), "templates/home.html"))

func homeHandler(w http.ResponseWriter, req *http.Request) {
	path, _ := r.Get("chat").URLPath("id", "Room 1")
	http.Redirect(w, req, path.String(), http.StatusFound)
	return
	//homeTempl.Execute(w, data)
}

func chatHandler(w http.ResponseWriter, req *http.Request) {
	data := struct {
		Host string
		ID   string
	}{
		Host: req.Host,
		ID:   mux.Vars(req)["id"],
	}
	//gotogether.LoadTemplates(
	homeTempl.Execute(w, data)
}

func injectorHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Thanks for sending your message.")

	go func() {
		msg := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
		for i := 0; i< 20; i++ {
			msg = append(msg, []byte("abcdefghijklmnopqrstuvwxyz0123456789")...)
		}
		hubs.BroadcastAll(msg)
	}()
}

//TODO(james):
//When registering the handler, pull out the ID that they registered on.
//That will be the channel ID, which determines which hub they register upon.
//Each hub will exist as a separate goroutine. Thus, each page will have a
//different non-blocking hub, but all messages on a given page will be in order.
//TODO(james):
//When sending new data over the socket, look up all parent IDs in an
// ancestry cache; if that's empty then check DB. Finally, send the
// data to anyone registered to any ancestor channels.
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
	//If not, create it.
	id := mux.Vars(req)["id"]
	log.Printf("wsHandler: Websocket connected on %s\n", id)

	//Buffer up to 256 messages for this client
	c := &connection{send: make(chan []byte, 256), ws: ws}

	//Register this connection into the (global var) hub
	log.Printf("wsHandler: GOT HERE%s\n", id)
	h := GetHub(id)
	log.Printf("wsHandler: CREATED HUB%s\n", id)
	h.register <- c
	log.Printf("wsHandler: REGISTERED CONNECTION TO HUB%s\n", id)

	//If this deferred function gets called, it implies that
	// writer and reader already exited
	defer func() {
		h.unregister <- c
		log.Printf("conn.wsHandler: SOCKET CLOSED %v is completely closed SOCKET CLOSED\n", c)
	}()
	go c.writer()
	c.reader(h)
}

func main() {
	log.Println("Launched")
	log.Printf("Hubmap: %+v", hubs)
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
