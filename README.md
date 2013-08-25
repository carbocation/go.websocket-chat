Golang Websocket Chat
=====================
This is a websocket-based chat library with examples; it uses [Garyburd's websocket library](https://github.com/garyburd/go-websocket), and is deeply inspired by his examples. 
Please see http://gary.beagledreams.com/page/go-websocket-chat.html for more details.

Usage
=====
`Connections` represent websocket connections from individual users. Those users can register with `hubs`, which are, in essence, 
chat rooms. Messages sent to a hub will be received by every user subscribed to a hub (i.e., to every user who has connected via a 
websocket which has not timed out or been closed).

See the "examples" section for working usage examples.