package ascendex

import (
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	ascendexHost string = "ascendex.github.io" //"ascendex.com"
	bbo          string = "bbo"
	sub          string = "sub"
	wsPath       string = "/api/pro/v1/stream"
	ping         string = "ping"
	pong         string = "pong"
)

var ErrorNoConnection = errors.New("not connection")

type APIClient struct {
	conn          *websocket.Conn
	wsURL         string
	convertSubSym func(string) string
	iscon         bool
	subGroup      sync.WaitGroup
	done          chan struct{}
	doneSub       chan struct{}
	mu            sync.RWMutex
	subChan       chan<- BestOrderBook
}

func GetAPIClient() *APIClient {
	return &APIClient{}
}

/*
Implement a websocket connection function
*/
func (api *APIClient) Connection() error {
	api.mu.Lock()
	defer api.mu.Unlock()
	if api.iscon {
		return nil
	}
	if api.done == nil {
		api.done = make(chan struct{})
	}
	api.iscon = true
	if api.wsURL == "" {
		api.wsURL = (&url.URL{Scheme: "ws", Host: ascendexHost, Path: wsPath}).String()
	}
	var err error
	api.conn, _, err = websocket.DefaultDialer.Dial(api.wsURL, nil)
	return errors.Wrapf(err, "ws connect %s", api.wsURL)
}

/*
Implement a disconnect function from websocket
*/
func (api *APIClient) Disconnect() {
	api.mu.Lock()
	defer api.mu.Unlock()
	if !api.iscon {
		return
	}
	api.iscon = false
	close(api.done)
	api.conn.Close()
}

/*
Implement a function that will subscribe to updates
of BBO for a given symbol
*/
func (api *APIClient) SubscribeToChannel(symbol string) error {
	api.mu.RLock()
	defer api.mu.RUnlock()
	if !api.iscon {
		return errors.Wrap(ErrorNoConnection, "subscribe error")
	}
	if api.convertSubSym == nil {
		api.convertSubSym = func(s string) string {
			strs := strings.Split(s, "_")
			itemCount := len(strs)
			for i := 0; i < itemCount/2; i++ {
				mirrorIdx := itemCount - i - 1
				strs[i], strs[mirrorIdx] = strs[mirrorIdx], strs[i]
			}
			return strings.Join(strs, "/")
		}
	}
	sym := api.convertSubSym(symbol)
	type message struct {
		Op string `json:"op"`
		Id string `json:"id"`
		Ch string `json:"ch"`
	}
	m := message{sub, "", strings.Join([]string{bbo, sym}, ":")}
	err := api.conn.WriteJSON(m)
	return errors.Wrap(err, "can't write JSON message")
}

/*
Implement a function that will write the data that
we receive from the exchange websocket to the channel
*/
func (api *APIClient) ReadMessagesFromChannel(ch chan<- BestOrderBook) {

	api.mu.RLock()
	defer api.mu.RUnlock()
	if !api.iscon {
		return
	}
	if ch == nil || api.subChan == ch {
		return
	}
	if api.doneSub == nil {
		api.doneSub = make(chan struct{})
	} else {
		close(api.doneSub)
		api.subGroup.Wait()
		api.doneSub = make(chan struct{})
	}
	type data struct {
		Bid []string `json:"bid"`
		Ask []string `json:"ask"`
	}
	type message struct {
		M    string `json:"m"`
		Data data   `json:"data"`
	}
	api.subGroup.Add(1)
	go func() {
		defer func() {
			close(ch)
			api.subGroup.Done()
		}()
		for {
			select {
			case <-api.done:
				return
			case <-api.doneSub:
				return
			default:
			}
			m := &message{}
			err := api.conn.ReadJSON(m)
			if err != nil {
				return
			}
			if strings.EqualFold(m.M, bbo) && len(m.Data.Ask) == 2 && len(m.Data.Bid) == 2 {

				amountAsk, err := strconv.ParseFloat(m.Data.Ask[0], 64)
				if err != nil {
					continue
				}
				priceAsk, err := strconv.ParseFloat(m.Data.Ask[1], 64)
				if err != nil {
					continue
				}
				amountBid, err := strconv.ParseFloat(m.Data.Bid[0], 64)
				if err != nil {
					continue
				}
				priceBid, err := strconv.ParseFloat(m.Data.Bid[1], 64)
				if err != nil {
					continue
				}
				bob := BestOrderBook{
					Ask: Order{amountAsk, priceAsk},
					Bid: Order{amountBid, priceBid},
				}
				ch <- bob
			}

		}

	}()
}

/*
Implement a function that will support connecting to a websocket
*/
func (api *APIClient) WriteMessagesToChannel() {
	api.mu.RLock()
	defer api.mu.RUnlock()
	if !api.iscon {
		return
	}
	type message struct {
		Op string `json:"op"`
	}
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()
		m := message{ping}
		for {
			select {
			case <-api.done:
				return
			case <-ticker.C:
				err := api.conn.WriteJSON(m)
				if err != nil {
					return
				}
			}
		}

	}()
}

// BestOrderBook struct
type BestOrderBook struct {
	Ask Order `json:"ask"`
	Bid Order `json:"bid"`
}

// Order struct
type Order struct {
	Amount float64 `json:"amount"`
	Price  float64 `json:"price"`
}
