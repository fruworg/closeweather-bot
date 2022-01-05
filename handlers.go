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

type OWM struct {
	City string `json:"city"`
}

func (a *application) startHandler(m *tbot.Message) {
	msg := "\n*Привет!* Чтобы начать, напиши город в чат.\nДалее введи необходимую *команду*:\n" +
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
			msg = fmt.Sprintf("%s %s Прогноз на сейчас\n\nТемпература: %.2f°\nОщущается как: %.2f°\nСкорость ветра: %.2f м/c\n%s.",
				w.Sys.Country, w.Name, w.Main.Temp, w.Main.FeelsLike, w.Wind.Speed, desk)
		}
	case "/week":
		city, err := client.Get(m.Chat.ID).Result()
		if err == redis.Nil {
			msg = "Сначала выбери город!\nКоманда /start в помощь."
		} else {
			w, err := owm.NewForecast("5", "C", "ru", os.Getenv("OWM_API_KEY")) // fahrenheit (imperial) with Russian output
			if err != nil {
				log.Fatalln(err)
			}
			city = strings.TrimLeft(city, `{"city":"`)
			city = strings.TrimRight(city, `"}`)
			w.DailyByName(city, 0)
			if val, ok := w.ForecastWeatherJson.(*owm.Forecast5WeatherData); ok {
				if len(val.List) != 0 {
					msg = fmt.Sprintf("%s %s Прогноз на неделю", val.City.Country, val.City.Name)
					for i := 0; i < 39; i++ {
						fl := strings.Split(fmt.Sprintf("%.2f", val.List[i:(i+1)]), " ")
						st := strings.Split(fmt.Sprintf("%s", val.List[i:(i+1)]), " ")
						dt := strings.Split(st[len(st)-4], "-")
						date := fmt.Sprintf("%s-%s-%s", dt[2], dt[1], dt[0])
						desc := strings.Title(strings.ToLower(st[11]))
						if len(st) == 26 {
							desc = desc + " " + st[12] + " " + st[13]
						}
						if len(st) == 25 {
							desc = desc + " " + st[12]
						}
						if st[len(st)-3] == "06:00:00" || st[len(st)-3] == "12:00:00" || st[len(st)-3] == "18:00:00" {
						msg = msg + fmt.Sprintf("\n\n%s %s\nТемпература: %s°\nОщущается: %s°\nВетер: %s м/c\n%s.",
							date, st[len(st)-3], strings.TrimLeft(fl[1], "{"), fl[4],
								  strings.TrimLeft(fl[14], "{"), desc)}
					}
				} else {
					msg = fmt.Sprintf("%v", len(val.List))
				}
			}
			url = "https://tesis.lebedev.ru/upload_test/files/fc_20220105.png"
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
	if url == "" {
		a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
	} else {
		a.client.SendPhoto(m.Chat.ID, url, tbot.OptCaption(msg))
	}
}
