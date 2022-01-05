package main

import (
	"log"
	"os"

	"github.com/yanzay/tbot/v2"
)

type application struct {
	client *tbot.Client
}

var (
	app   application
	bot   *tbot.Server
	token string
)

func main() {
	bot = tbot.New(os.Getenv("TELEGRAM_TOKEN"), tbot.WithWebhook("https://magicdeer-bot.herokuapp.com", ":"+os.Getenv("PORT")))
	app.client = bot.Client()
	bot.HandleMessage("/start", app.startHandler)
	bot.HandleMessage(".+", app.msgHandler)
	log.Fatal(bot.Start())
}
