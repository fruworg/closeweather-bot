package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
			desk := ""
			geo := ""
			arr := strings.Split(fmt.Sprintf("", w.Weather), " ")
			for i := 3; i < len(arr)-1; i++ {
				if i == 3 {
					desk = strings.Title(strings.ToLower(arr[i]))
				} else {
					desk = desk + " " + arr[i]
				}
			}

			res, err := http.Get("https://tesis.lebedev.ru/forecast_activity.html")
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
			}

			// Load the HTML document
			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			doc.Find(`td`).Each(func(i int, s *goquery.Selection) {
				if i == 48 {
					geo = fmt.Sprintf("Магнитная буря: %s\n", s.Text())
				}
				if i == 51 {
					geo = geo + fmt.Sprintf("Процентаж сильной: %s", s.Text())
				}
			})
			msg = fmt.Sprintf("%s %s %s\n\nТемпература: %.2f°\nОщущается как: %.2f°\nСкорость ветра: %.2f м/c\n\n%s",
			      w.Sys.Country, w.Name, desk, w.Main.Temp, w.Main.FeelsLike, w.Wind.Speed, geo)
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
	a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
}
