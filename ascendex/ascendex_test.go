package ascendex

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}
func TestConnection(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	u, err := url.Parse("http://127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	u.Scheme = "ws"
	api := APIClient{}
	api.wsURL = u.String()
	api.Connection()
	// Connect to the server
	ws := api.conn
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("%v", err)
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}
}
func TestDisconnect(t *testing.T) {

}
func TestSubscribeToChannel(t *testing.T) {

}
func TestReadMessagesFromChannel(t *testing.T) {

}
func TestWriteMessagesToChannel(t *testing.T) {

}
