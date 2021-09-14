package main

import (
	Fmt "fmt"
	Ws "github.com/gorilla/websocket"
	Log "log"
	Json "encoding/json"
)

func main() {
	collect()
	return
}

type KlinePayload struct {
	Symbol 		string `json:"s"`
	Kline  		Kline  `json:"k"`
}

type Kline struct {
	Time        int    `json:"t"`
	High        string `json:"h"`
	Trades      int    `json:"n"`
	VolumeBase  string `json:"v"`
	VolumeQuote string `json:"q"`
	VolumeV     string `json:"V"`
	VolumeQ     string `json:"Q"`
	Closed		bool   `json:"x"`
}

func collect() {
	const socket = "wss://stream.binance.com:9443/ws/btcusdt@kline_1m"

	var wsDialer Ws.Dialer // howto interfaces

	for {
		wsConn, _, err := wsDialer.Dial(socket, nil)
		if err != nil {
			Log.Println(err.Error())
		}

		for {
			msgType, resp, err := wsConn.ReadMessage()
			if err != nil {
				Log.Println(string(resp))
				Log.Println(err)
				break
			}

			if msgType != Ws.TextMessage {
				Log.Println(string(resp))
				Log.Println(msgType)
				continue
			}

			json := KlinePayload{}

			err = Json.Unmarshal(resp, &json)
			if err != nil {
				Log.Println(err)
				continue
			}

			k := json.Kline

			if json.Kline.Closed {
				Fmt.Printf("%v %v %v %v %v %v %v\n",
					k.Time / 1000,
					k.High,
					k.Trades,
					k.VolumeBase,
					k.VolumeQuote,
					k.VolumeV,
					k.VolumeQ,
				)
			}
		}
	}
}

