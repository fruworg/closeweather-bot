package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"encoding/json"

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

type OWM struct {
	City string `json:"city"`
}

func (a *application) startHandler(m *tbot.Message) {
	msg := "\n*Привет!* Чтобы начать, напиши город в чат.\nДалее введи необходимую *команду*:\n"+
	"Команда */today* - прогноз на сегодня.\nКоманда */week* - прогноз на 5 дней."
	a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
}

// Handle the msg command here
func (a *application) msgHandler(m *tbot.Message) {
	a.client.SendChatAction(m.Chat.ID, tbot.ActionTyping)
	msg := "Ты сделал что-то не так!"
	url := ""
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
			desk := ""
			arr := strings.Split(fmt.Sprintf("", w.Weather), " ")
			for i := 3; i < len(arr)-1; i++ {
				if i == 3 {
					desk = strings.Title(strings.ToLower(arr[i]))
				} else {
					desk = desk + " " + arr[i]
				}
			}
			
			url = "https://tesis.lebedev.ru/magnetic_storms.html?date=20220105"
			msg = fmt.Sprintf("%s %s Прогноз погоды\n\nТемпература: %.2f°\nОщущается как: %.2f°\nСкорость ветра: %.2f м/c\n%s",
				w.Sys.Country, w.Name, w.Main.Temp, w.Main.FeelsLike, w.Wind.Speed, desk)
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
		if w.Cod != 200 {
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
			msg = "Город изменён - " + m.Text + "."
		}
	}
	if url == ""{
		a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)}else{
	a.client.SendPhoto(m.Chat.ID, url, tbot.OptCaption(msg))
	}
}
