package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	owm "github.com/briandowns/openweathermap"
	"github.com/go-redis/redis"
	"github.com/yanzay/tbot/v2"
)

var opt, err = redis.ParseURL(os.Getenv("REDIS_URL"))
var client = redis.NewClient(&redis.Options{
	Addr:     opt.Addr,
	Password: opt.Password,
	DB:       opt.DB,
})

type Author struct {
	City string `json:"city"`
}

func (a *application) startHandler(m *tbot.Message) {
	msg := "test"
	a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
}

// Handle the msg command here
func (a *application) msgHandler(m *tbot.Message) {
	a.client.SendChatAction(m.Chat.ID, tbot.ActionTyping)
	msg := "Ты сделал что-то не так!"
	switch m.Text {
	case "/today":
		city, err := client.Get(m.Chat.ID).Result()
		if err == redis.Nil {
			msg = "Сначала выбери город!\nКоманда /start в помощь."
		} else {
			w, err := owm.NewCurrent("C", "ru", os.Getenv("OWM_API_KEY")) // fahrenheit (imperial) with Russian output
			if err != nil {
				log.Fatalln(err)
			}
			city = strings.TrimLeft(city, `{"city":"`)
			city = strings.TrimRight(city, `"}`)
			w.CurrentByName(city)
			msg = fmt.Sprintf("%s", w.Main.Temp)
		}
	case "/week":
		city, err := client.Get(m.Chat.ID).Result()
		if err == redis.Nil {
			msg = "Сначала выбери город!\nКоманда /start в помощь."
		} else {
			w, err := owm.NewCurrent("C", "ru", os.Getenv("OWM_API_KEY")) // fahrenheit (imperial) with Russian output
			if err != nil {
				log.Fatalln(err)
			}
			city = strings.TrimLeft(city, `{"city":"`)
			city = strings.TrimRight(city, `"}`)
			w.CurrentByName(city)
			msg = fmt.Sprintf("%s", w.Main.Temp)
		}
	default:
		w, err := owm.NewCurrent("C", "ru", os.Getenv("OWM_API_KEY")) // fahrenheit (imperial) with Russian output
		if err != nil {
			log.Fatalln(err)
		}
		w.CurrentByName(m.Text)
		if w.Cod == 200 {
			msg = "Город не найден!"
		} else {
			json, err := json.Marshal(OWM{City: m.Text})
			if err != nil {
				fmt.Println(err)
			}
			err = client.Set(m.Chat.ID, json, 0).Err()
			if err != nil {
				fmt.Println(err)
			}
			msg = "Город изменён на" + m.Text + "."
		}
	}
	a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
}
