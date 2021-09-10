package main

// в параллельном процессе без блокировок обновлять цену каждую секунду
// цену записывать в кэш, в некое хранилишие по типу рэдиса
// структура под запись разных значений... как лучше всего парсить JSON сразу в структуру?
// но рэдис слишком дорог, нужна своя штука тут
// прикрутить хуки телеги

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"os"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
)

var usd float32
const ticker = "https://blockchain.info/ticker"

func main() {
	go updateTicket()

	token := os.Getenv("TOKEN")

	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, "BTCUSD: " + fmt.Sprint(usd))
	})

	b.Handle("/btc", func(m *tb.Message) {
		b.Send(m.Sender, "BTCUSD: " + fmt.Sprint(usd))
	})

	b.Start()
}

type price struct {
	Symbol string
	symbol string // hm... wtw?
	// 15m how?
	Buy float32
	Sell float32
	Last float32
}

func updateTicket() {
	for {
		client := http.Client{Timeout: time.Second * 2}

		req, _ := http.NewRequest(http.MethodGet, ticker, nil)
		res, e := client.Do(req)

		if e != nil {
			log.Println("Skip update price")
			time.Sleep(time.Second * 10)
			continue
		}

		body, _ := ioutil.ReadAll(res.Body)

		var result map[string]price
		json.Unmarshal(body, &result)
		usd = result["USD"].Last
		log.Println("Updated price", usd)
		time.Sleep(time.Second * 10)
	}
}
