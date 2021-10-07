# TODO
 - добавить coinbase => все же начать юзать библиотеки с врапперами... но возможно нужно подлкючать более редкие биржи, нужно уметь самому писать такие врапперы
 - добавить битмекс
 - маржинальный арбитраж... получается с плечом или нет?
 - common mistackes Go https://yourbasic.org/golang/gotcha/
 - добавить sqlite, https://gorm.io/docs/models.html https://github.com/xo/xo
 - подключить универсальный враппер
 - постоянно подтягивать стаканы с 10 бирж
 - записывать стаканы глубиной X
 - рассчитывать текущую цену на основе бирж под какой-то объем... мол сколько можно унести

# пред. алго для торговли
 - следим за балансом на 1 и 2
 - если есть балансы для соврешения сделки для нового гэпа - делаем, иначе скипаем
 - после выполнения ордеров обновляем балансы в боте
 
# Libs

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

# записки на манжетах 
 - язык компилируемый, строгой типизации, но при этом с большим числом сахара, пусть компилятор будет сложным
 - карго культ TypeScript
 - https://github.com/preichenberger/go-coinbasepro/blob/master/fees.go
 - как работает наследование в Go? по сравнению со структурами TS
 - сколько жрет node? всё же у ноды больше врапперов... но нет трендов(
 - чем плоха node - смешивание стилей разных разрабов, большая разница, типа php

# Read
 - https://gobyexample.com/rate-limiting
 - https://github.com/thrasher-corp/gocryptotrader/blob/master/gctscript/README.md
 - https://hussachai.medium.com/error-handling-in-go-a-quick-opinionated-guide-9199dd7c7f76
