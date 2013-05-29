package wshub

import (
"encoding/json"
"errors"
)

//Packet provides structure for applying json callbacks
type Packet struct {
	Event string
	Data  interface{}
}

//Packetize takes in an event name and arbitrary data and returns
// a stringified json object converted to a bite slice.
func Packetize(event string, data interface{}) ([]byte, error) {
	p := Packet{Event: event, Data: data}
	jsondata, err := json.Marshal(p)
	if err != nil {
		return []byte(""), errors.New("wshub.Packetize: Could not marshal the provided data.") 
	}
	
	return jsondata, nil
}