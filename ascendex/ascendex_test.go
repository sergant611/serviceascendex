package ascendex

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
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

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	u.Scheme = "ws"
	api := APIClient{}
	api.wsURL = u.String()
	api.Connection()
	ws := api.conn
	defer ws.Close()

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
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	u.Scheme = "ws"
	api := APIClient{}
	api.wsURL = u.String()
	api.conn, _, err = websocket.DefaultDialer.Dial(api.wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	api.iscon = true
	api.done = make(chan struct{})
	ws := api.conn
	defer ws.Close()
	api.Disconnect()
	if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err == nil {
		t.Error("no close")
	}

}
func TestSubscribeToChannel(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	u.Scheme = "ws"
	api := APIClient{}
	api.wsURL = u.String()
	api.conn, _, err = websocket.DefaultDialer.Dial(api.wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	api.iscon = true
	api.done = make(chan struct{})
	ws := api.conn
	defer ws.Close()
	api.SubscribeToChannel("USDT_BTC")
	type message struct {
		Op string `json:"op"`
		Id string `json:"id"`
		Ch string `json:"ch"`
	}
	m := &message{}
	testm := &message{sub, "", strings.Join([]string{bbo, "BTC/USDT"}, ":")}
	err = ws.ReadJSON(m)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if diff := deep.Equal(m, testm); diff != nil {
		t.Error(diff)
	}
}
func TestReadMessagesFromChannel(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	u.Scheme = "ws"
	api := APIClient{}
	api.wsURL = u.String()
	api.conn, _, err = websocket.DefaultDialer.Dial(api.wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	api.iscon = true
	api.done = make(chan struct{})
	ws := api.conn
	defer ws.Close()
	ch := make(chan BestOrderBook)
	api.ReadMessagesFromChannel(ch)
	type data struct {
		Bid []string `json:"bid"`
		Ask []string `json:"ask"`
	}
	type message struct {
		M    string `json:"m"`
		Data data   `json:"data"`
	}
	m := &message{
		M: bbo,
		Data: data{
			Bid: []string{"20.0", "0.2"},
			Ask: []string{"10.0", "0.1"},
		},
	}
	wantbob := &BestOrderBook{
		Ask: Order{10.0, 0.1},
		Bid: Order{20.0, 0.2},
	}
	err = ws.WriteJSON(m)
	if err != nil {
		t.Fatalf("%v", err)
	}
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	select {
	case <-ticker.C:
		t.Error("timed out")
	case b := <-ch:
		if diff := deep.Equal(&b, wantbob); diff != nil {
			t.Error(diff)
		}

	}
}
func TestWriteMessagesToChannel(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	u.Scheme = "ws"
	api := APIClient{}
	api.wsURL = u.String()
	api.conn, _, err = websocket.DefaultDialer.Dial(api.wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	api.iscon = true
	api.done = make(chan struct{})
	ws := api.conn
	defer ws.Close()
	api.WriteMessagesToChannel()
	type message struct {
		Op string `json:"op"`
	}
	m := &message{}
	testm := &message{ping}
	err = ws.ReadJSON(m)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if diff := deep.Equal(m, testm); diff != nil {
		t.Error(diff)
	}
}
