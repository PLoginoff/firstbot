package main

// собирать стаканы с N бирж, мониторить и записывать промежутки (gaps)
// в качестве интерфейса — бот в телеграм @tvconnector_bot
// сервер: 128mb за натом

// текущие недостатки:
//  - нужно было сразу использовать прослойку (но там мало доки) и сравнивать хотя бы 10 бирж
//  - меньше копипасты, больше интерфейсов

import (
	Json "encoding/json"
	Fmt "fmt"
	Gabs "github.com/Jeffail/gabs"
	Ws "github.com/gorilla/websocket"
	Deep "github.com/patrikeh/go-deep"
	Training "github.com/patrikeh/go-deep/training"
	Load "github.com/shirou/gopsutil/load"
	Mem "github.com/shirou/gopsutil/mem"
	Tg "gopkg.in/tucnak/telebot.v2"
	IOUtil "io/ioutil"
	Log "log"
	"math"
	Http "net/http"
	Os "os"
	"strconv"
	Time "time"
)

var usd float32

const ticker = "https://blockchain.info/ticker"
const binanceSocket = "wss://stream.binance.com:9443/ws/btcusdt@depth10@100ms"
const bitfinexSocket = "wss://api.bitfinex.com/ws/2"

const minGap = 0.23
const minGapFind = 0.1

func main() {
	go updateBook1()
	go updateBook2()
	go updateTicket()
	go findGap(minGapFind)

	token := Os.Getenv("TELEGRAM")

	b, err := Tg.NewBot(Tg.Settings{
		Token:     token,
		Poller:    &Tg.LongPoller{Timeout: 10 * Time.Second},
		ParseMode: "HTML",
	})

	if err != nil {
		Log.Fatal(err)
		return
	}

	b.Handle("/start", func(m *Tg.Message) {
		b.Send(m.Sender, "My first bot on go. Try to use /gap /gaps /btc /top")
	})

	b.Handle("/btc", func(m *Tg.Message) {
		b.Send(m.Sender, "BTCUSD: "+Fmt.Sprintf("%.2f", usd))
	})

	b.Handle("/gap", func(m *Tg.Message) {
		print1 := "<b>Bitfinex:</b>\n<pre>" + printBook(orderBook1) + "</pre>\n\n<b>Binance:</b>\n<pre>" + printBook(orderBook2) +
			"</pre>\n\n<b>Gap</b> (first version):\n" + Fmt.Sprintf("%.2f%%", calcGap(orderBook1, orderBook2))
		_, err := b.Send(m.Sender, print1)
		if err != nil {
			Log.Println(err.Error())
		}
	})

	// todo make `/gaps 0.222`
	b.Handle("/gaps", func(m *Tg.Message) {
		minGapParam := minGap

		if len(caughtGaps) > 0 {
			maxLines := 10
			b.Send(m.Sender, Fmt.Sprintf("I found %v gaps, show > %.2f: ", len(caughtGaps), minGapParam))
			for _, s := range caughtGaps {
				if s.Gap > minGapParam {
					show := Fmt.Sprintf(
						"Time: %v\nGap: %.2f\n<b>Bitfinex:</b>\n<pre>%v</pre>\n\n<b>Binance:</b>\n<pre>%v</pre>",
						s.Date.Format(Time.RFC3339),
						s.Gap,
						printBook(orderBook1),
						printBook(orderBook2),
					)
					b.Send(m.Sender, show)
					if maxLines--; maxLines < 0 {
						break
					}
				}
			}
		} else {
			b.Send(m.Sender, Fmt.Sprintf("No gaps for %.2f :-(", minGap))
		}
	})

	b.Handle("/top", func(m *Tg.Message) {
		load, _ := Load.Avg()
		mem, _ := Mem.VirtualMemory()
		top := Fmt.Sprintf("Load: %.2f %.2f %.2f", load.Load1, load.Load5, load.Load15) +
			Fmt.Sprintf("\nMem: %vm / %vm / %.0f%%", mem.Available/1024/1024, mem.Total/1024/1024, mem.UsedPercent)
		_, err := b.Send(m.Sender, top)
		if err != nil {
			Log.Println(top)
			Log.Fatal(err)
		}
	})

	b.Start()
}

type price struct {
	Symbol string `json:"symbol"`
	symbol string // todo hm... wtw?
	Last15 string // todo error... `json:"15m"`
	Buy    float32
	Sell   float32
	Last   float32
}

type Stack struct {
	Buy  [10]float64
	Sell [10]float64
}

func updateTicket() {
	for {
		Time.Sleep(Time.Second * 33)

		client := Http.Client{Timeout: Time.Second * 2}

		req, _ := Http.NewRequest(Http.MethodGet, ticker, nil)
		res, e := client.Do(req)

		if e != nil {
			Log.Println("Skip update price, GET error")
			continue
		}

		body, _ := IOUtil.ReadAll(res.Body)

		var result map[string]price
		err := Json.Unmarshal(body, &result)

		if err != nil {
			Log.Println("Skip update price, parse error")
			continue
		}

		usd = result["USD"].Last
	}
}

type Order struct {
	Price float64
	Size  float64
}
type OrderBook map[int8]Order

var orderBook1 = make(map[int8]Order)   // bitfinex
var orderBook2 = make(map[int8]Order)   // binance
type CaughtGap struct {
	Gap    float64
	Date   Time.Time
	First  OrderBook
	Second OrderBook
}
var caughtGaps = make(map[int32]CaughtGap) // unix:event

func findGap(minGap float64) {
	if minGap < 0.001 {
		minGap = 0.1
	}

	for {
		Time.Sleep(Time.Millisecond * 100)
		currentTime := Time.Now()
		gap := calcGap(orderBook1, orderBook2)

		if gap >= minGap && math.Abs(gap) < 20 {
			println(currentTime.String())

			println("Bitfinex:")
			println(printBook(orderBook1))

			println("Binance:")
			println(printBook(orderBook2))

			println("Gap (first version!):")
			println(Fmt.Sprintf("%.2f%%", gap))

			// todo how to copy struct?
			caughtGaps[int32(currentTime.Unix())] = CaughtGap{gap, currentTime, orderBook1, orderBook2}
		}
	}
}

func printBook(orderBook map[int8]Order) string {
	return Fmt.Sprintf(
		"+%.2f x%6.3f\n+%.2f x%6.3f\n+%.2f x%6.3f\n+%.2f x%6.3f\n+%.2f x%6.3f\n-%.2f x%6.3f\n-%.2f x%6.3f\n-%.2f x%6.3f\n-%.2f x%6.3f\n-%.2f x%6.3f",
		orderBook[5].Price, orderBook[5].Size,
		orderBook[4].Price, orderBook[4].Size,
		orderBook[3].Price, orderBook[3].Size,
		orderBook[2].Price, orderBook[2].Size,
		orderBook[1].Price, orderBook[1].Size,
		orderBook[-1].Price, orderBook[-1].Size,
		orderBook[-2].Price, orderBook[-2].Size,
		orderBook[-3].Price, orderBook[-3].Size,
		orderBook[-4].Price, orderBook[-4].Size,
		orderBook[-5].Price, orderBook[-5].Size,
	)
}

// first version for gap
func calcGap(book1 map[int8]Order, book2 map[int8]Order) float64 {
	gap := (book1[1].Price/book2[-1].Price)*100 - 100.0
	return gap
}

// bitfinex
func updateBook1() {
	var wsDialer Ws.Dialer
	for { // todo redial как это сделать более красиво с этом wsDealler
		wsConn, _, err := wsDialer.Dial(bitfinexSocket, nil)
		if err != nil {
			println(err.Error())
			continue
		}

		subscribe := map[string]string{
			"event":   "subscribe",
			"channel": "book",
			"symbol":  "tBTCUSD",
			"prec":    "P0",
			"freq":    "F0",
		}
		if err := wsConn.WriteJSON(subscribe); err != nil {
			println(err.Error())
			continue
		}

		for {
			msgType, resp, err := wsConn.ReadMessage()
			if err != nil {
				Fmt.Println(err)
				break
			}

			if msgType != Ws.TextMessage {
				continue
			}

			jsonParsed, err := Gabs.ParseJSON(resp)
			if err != nil {
				Log.Println(err)
				continue
			}

			exists := jsonParsed.Exists("event")
			if exists {
				continue
			}

			row, _ := jsonParsed.Index(1).Children()
			count := len(row)
			if count == 3 {
				row = []*Gabs.Container{jsonParsed.Index(1)}
			}

			for _, order := range row {
				if c, _ := order.ArrayCount(); c < 3 {
					continue
				}
				price := order.Index(0).Data().(float64)
				numOrder := order.Index(1).Data().(float64)
				size := order.Index(2).Data().(float64)
				id := int8(numOrder) // 1..5

				if size < 0 { // sell
					size = 0 - size
					id = 0 - id
				}

				orderBook1[id] = Order{Size: size, Price: price}
			}
		}
	}
}

// binance
func updateBook2() {
	var wsDialer Ws.Dialer // todo howto interfaces

	for { // todo redial
		wsConn, _, err := wsDialer.Dial(binanceSocket, nil)
		if err != nil {
			println(err.Error())
			continue
		}

		for {
			msgType, resp, err := wsConn.ReadMessage()
			if err != nil {
				Fmt.Println(err)
				break
			}

			if msgType != Ws.TextMessage {
				Log.Println(msgType)
				continue
			}

			jsonParsed, err := Gabs.ParseJSON(resp)
			if err != nil {
				Log.Println(err)
				continue
			}

			list, err := jsonParsed.Path("bids").Children() // buy
			if err != nil {
				Log.Println(string(resp))
				Log.Println(err.Error())
				break
			}

			id := int8(1)
			for _, order := range list {
				price, _ := strconv.ParseFloat(order.Index(0).Data().(string), 64)
				size, _ := strconv.ParseFloat(order.Index(1).Data().(string), 64)
				orderBook2[id] = Order{Size: size, Price: price}
				id = id + 1
			}

			bids, _ := jsonParsed.Path("asks").Children() // sell
			id = -1
			for _, order := range bids {
				price, _ := strconv.ParseFloat(order.Index(0).Data().(string), 64)
				size, _ := strconv.ParseFloat(order.Index(1).Data().(string), 64)
				orderBook2[id] = Order{Size: size, Price: price}
				id = id - 1
			}
		}
	}
}
