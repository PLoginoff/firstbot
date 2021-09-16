# TODO
 - добавить coinbase => все же начать юзать библиотеки с врапперами... но возможно нужно подлкючать более редкие биржи, нужно уметь самому писать такие врапперы
 - добавить битмекс
 - твоя первая тыша строк на Go, карго культ TypeScript

https://github.com/preichenberger/go-coinbasepro/blob/master/fees.go
как работает наследование в Go? по сравнению со структурами TS
сколько жрет node? всё же у ноды больше врапперов... но нет трендов(
мир крипты все же крутится вокруг node
выбор с точки зрения бабла... где больше подработкок?
взять подработку на го как джун?!
чем плоха node - смешивание стилей разных разрабов, большая разница, типа php

статья на vc список где искать работу


Алго такой:
 - следим за балансом на 1 и 2
 - если есть балансы для соврешения сделки для нового гэпа - делаем, иначе скипаем
 - после выполнения ордеров обновляем балансы в боте

Маржинальный арбитраж... получается с плечом или нет?

Внутрибиржевой арбитраж:
  У тебя есть 100$. На 100 ты покупаешь 1 бтс.
  1 бтс обмениваешь на 10 eth
  10 eth обратно в доллары. По итогу у тебя будет не 100$, а 105$ допустим

Exchange Rest And WebSocket API For Golang Wrapper support okcoin,okex,huobi,hbdm,bitmex,coinex,poloniex,bitfinex,bitstamp,binance,kraken,bithumb,zb,hitbtc,fcoin, coinbene
https://github.com/nntaoli-project/goex

A golang implementation of a console-based trading bot for cryptocurrency exchanges
https://github.com/saniales/golang-crypto-trading-bot
есть стратегии! но что-то заумно

просто бифеникс
https://github.com/bitfinexcom/bitfinex-api-go
https://github.com/bitfinexcom/bitfinex-api-go

A cryptocurrency trading bot and framework supporting multiple exchanges written in Golang.
https://github.com/thrasher-corp/gocryptotrader

https://github.com/robaho/go-trader

https://github.com/thierryung/gocoin

TODO:
 - common mistackes Go https://yourbasic.org/golang/gotcha/
 - язык компилируемый, сторой типизацией, но при этом с большим числом сахара, пусть компилятор будет сложным
 - добавить sqlite, https://gorm.io/docs/models.html https://github.com/xo/xo
 - подключить универсальный враппер
 - постоянно подтягивать стаканы с 10 бирж
 - записывать стаканы глупиной X 
 - расчитывать текущую цену на основе бирж под какой-то объем... мол сколько можно унести
 - сделать команду `/arbitrate bitfinex VALUE` - выводит diff всех бирж с первой и возможный профит при попытке торгануть VALUE
 
# Read
 - https://gobyexample.com/rate-limiting
 - https://github.com/thrasher-corp/gocryptotrader/blob/master/gctscript/README.md
 - https://hussachai.medium.com/error-handling-in-go-a-quick-opinionated-guide-9199dd7c7f76
