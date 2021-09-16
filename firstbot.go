package main

// добавтиь цикл для переподключения сокетов загрузки стаканов


// в параллельном процессе без блокировок обновлять цену каждую секунду
// цену записывать в кэш, в некое хранилишие по типу рэдиса
// структура под запись разных значений... как лучше всего парсить JSON сразу в структуру?
// но рэдис слишком дорог, нужна своя штука тут
// прикрутить хуки телеги

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

func main() {
	const minGap = 0.23

	go updateBook1()
	go updateBook2()
	go updateTicket()
	go findGap(minGap)
	// ml()
	// return
	// printBooks()

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
		b.Send(m.Sender, "My first bot on go. Try to use /gap /btc /book")
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

	b.Handle("/gaps", func(m *Tg.Message) {
		if len(caughtGaps) > 0 {
			max := 10
			b.Send(m.Sender, "I found "+Fmt.Sprintf("%v", len(caughtGaps))+" gaps: ")
			for i, s := range caughtGaps {
				time := Time.Unix(int64(i), 0)
				show := time.Format(Time.RFC3339) + "\n" + s
				b.Send(m.Sender, show)
				if max--; max < 0 {
					break
				}
			}
		} else {
			b.Send(m.Sender, "No gaps :-(")
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

	b.Handle("/book", func(m *Tg.Message) {
		gap1 := (orderBook1[1].Price/orderBook1[-1].Price)*100 - 100.0
		log := Fmt.Sprintf("%.0f %.0f %.0f > %.2f < %.0f %.0f %.0f",
			orderBook1[-3].Price,
			orderBook1[-2].Price,
			orderBook1[-1].Price,
			gap1,
			orderBook1[1].Price,
			orderBook1[2].Price,
			orderBook1[3].Price)
		b.Send(m.Sender, log)
	})

	b.Start()
}

type price struct {
	Symbol string `json:"symbol"`
	symbol string // hm... wtw?
	Last15 string // error... `json:"15m"`
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
		// Log.Println("Updated price", usd)
	}
}

type Order struct {
	Price float64
	Size  float64
}
type OrderBook map[int8]Order

var orderBook1 = make(map[int8]Order)   // bitfinex
var orderBook2 = make(map[int8]Order)   // binance
var caughtGaps = make(map[int32]string) // unix:event

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

			log := Fmt.Sprintf("Gap: %.2f%%", gap) + "\nBitfinex:\n" + printBook(orderBook1) + "\nBinance:\n" + printBook(orderBook2)
			caughtGaps[int32(currentTime.Unix())] = log
		}
	}
}

func printBooks() {
	for {
		Time.Sleep(Time.Second * 5)

		/*
			gap1 := math.Abs((orderBook1[1].Price / orderBook1[-1].Price) * 100 - 100.0)
			Log.Printf("Bitfinex:   %.2f %.2f %.2f %.2f %.2f > %.2f < %.2f %.2f %.2f %.2f %.2f",
				orderBook1[-5].Price,
				orderBook1[-4].Price,
				orderBook1[-3].Price,
				orderBook1[-2].Price,
				orderBook1[-1].Price,
				gap1,
				orderBook1[1].Price,
				orderBook1[2].Price,
				orderBook1[3].Price,
				orderBook1[4].Price,
				orderBook1[5].Price)
		*/

		println("Bitfinex:")
		println(printBook(orderBook1))

		println("Binance:")
		println(printBook(orderBook2))

		println("Gap (first version!):")
		println(Fmt.Sprintf("%.2f%%", calcGap(orderBook1, orderBook2)))
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
	for { // redial todo как это сделать более красиво с этом wsDealler
		wsConn, _, err := wsDialer.Dial("wss://api.bitfinex.com/ws/2", nil)
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
	const socket = "wss://stream.binance.com:9443/ws/btcusdt@depth10@100ms"

	var wsDialer Ws.Dialer // howto interfaces

	for { // redial
		wsConn, _, err := wsDialer.Dial(socket, nil)
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

/*
пока алго такой вот:
 - 20points * 5values = 100 inputs - 1 output
 - 100, 100, 100, 100

зачем нормализовать?
 - чтобы можно было применять нейронку к разным парам



*/

func ml() {
	n := Deep.NewNeural(&Deep.Config{
		/* Input dimensionality */
		Inputs: 100,
		/* Two hidden layers consisting of two neurons each, and a single output */
		Layout: []int{100, 100, 1}, // 1000*1000 -> 1
		/* Activation functions: Sigmoid, Tanh, ReLU, Linear */
		Activation: Deep.ActivationSigmoid,
		/* Determines output layer activation & loss function:
		ModeRegression: linear outputs with MSE loss
		ModeMultiClass: softmax output with Cross Entropy loss
		ModeMultiLabel: sigmoid output with Cross Entropy loss
		ModeBinary: sigmoid output with binary CE loss */
		Mode: Deep.ModeBinary,
		/* Weight initializers: {deep.NewNormal(μ, σ), deep.NewUniform(μ, σ)} */
		Weight: Deep.NewNormal(1.0, 0.0),
		/* Apply bias */
		Bias: true,
	})

	// params: learning rate, momentum, alpha decay, nesterov
	optimizer := Training.NewSGD(0.05, 0.1, 1e-6, true)
	// params: optimizer, verbosity (print stats at every 50th iteration)
	trainer := Training.NewTrainer(optimizer, 100)
	// trainer := Training.NewBatchTrainer(optimizer, 50, 200, 4)

	Time.Sleep(Time.Second * 5) // wait data...

	prev := make(OrderBook) // prev version
	max := 25 // points
	examples := make(Training.Examples, max)

	// train cycle

	Log.Println("Collect data...")

	for k, v := range orderBook2 {
		prev[k] = v
	}

	i := 0
	for {
		Time.Sleep(Time.Millisecond * 1000)
		nextPrice := calcPrice(orderBook2)
		examples[i] = Training.Example{Input: formatBook(prev), Response: []float64{nextPrice}}
		if i++; i >= max {
			break
		}
		Log.Println("Diff:", nextPrice - calcPrice(prev))
		for k, v := range orderBook2 {
			prev[k] = v
		}
	}

	training, heldout := examples.Split(0.5)
	trainer.Train(n, training, heldout, 1000) // training, validation, iterations

	i = 0
	for {
		Time.Sleep(Time.Millisecond * 1000)
		println(printBook(orderBook2))
		println("Price: ", calcPrice(orderBook2))
		predict := n.Predict(formatBook(orderBook2))
		println("Predict: ", predict[0])

		if i++; i >= 5 {
			break
		}
	}
}

// todo func formatBook(b OrderBook) ([20]float64 plain) {
func formatBook(b OrderBook) []float64 {
	plain := make([]float64, 10)
	for pos, i := range []int8{-5, -4, -3, -2, -1, 1, 2, 3, 4, 5} {
		plain[pos] = b[i].Price
		// plain[pos + 1] = b[i].Size
	}
	return plain
}

func calcPrice(b OrderBook) float64 { // todo or ref?
	return (b[-1].Price + b[1].Price) / 2
}