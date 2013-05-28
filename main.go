package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/carbocation/gotogether"
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
		hubs.BroadcastAll("A third party injected this message for fun or for profit.")
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
func wsHandler(ws *websocket.Conn) {
	//When we try to handle this, see if the hub exists.
	//If not, create it.
	id := mux.Vars(ws.Request())["id"]
	fmt.Printf("wsHandler: Websocket connected on %s\n", id)

	//Buffer up to 256 messages for this client
	c := &connection{send: make(chan string, 256), ws: ws}

	//Register this connection into the (global var) hub
	fmt.Printf("wsHandler: GOT HERE%s\n", id)
	h := GetHub(id)
	fmt.Printf("wsHandler: CREATED HUB%s\n", id)
	h.register <- c
	fmt.Printf("wsHandler: REGISTERED CONNECTION TO HUB%s\n", id)

	//If this deferred function gets called, it implies that
	// writer and reader already exited
	defer func() {
		h.unregister <- c
		fmt.Printf("conn.wsHandler: SOCKET CLOSED %v is completely closed SOCKET CLOSED\n", c)
	}()
	go c.writer()
	c.reader(h)
}

func main() {
	fmt.Println("Launched")
	fmt.Printf("Hubmap: %+v", hubs)
	flag.Parse()

	r = mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/injector", injectorHandler)
	r.HandleFunc(`/{id:[_!.,+\- a-zA-Z0-9]+}`, chatHandler).Name("chat")
	r.Handle(`/ws/{id:[_!.,+\- a-zA-Z0-9]+}`, websocket.Handler(wsHandler))
	r.HandleFunc("/loaderio-766fe6ab96dba175477e09ae8baf291c/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "loaderio-766fe6ab96dba175477e09ae8baf291c")
	})

	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
