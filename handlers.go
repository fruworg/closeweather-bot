package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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
	msg := "\n*Привет!* Для начала напиши город в чат.\nДалее введи необходимую *команду*:\n" +
		"Команда */today* - прогноз на сегодня.\nКоманда */week* - прогноз на 5 дней."
	a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
}

// Handle the msg command here
func (a *application) msgHandler(m *tbot.Message) {
	citycodes := map[string]string{
		"абакан":           "QYPM",
		"алматы":           "P8OF",
		"анадырь":          "SSWT",
		"архангельск":      "SRLE",
		"астана":           "QJNY",
		"астрахань":        "PQM0",
		"ашхабад":          "OCMV",
		"баку":             "OQM6",
		"барнаул":          "QWOZ",
		"белгород":         "QGL2",
		"биробиджан":       "Q5T3",
		"бишкек":           "P5O8",
		"благовещенск":     "QESN",
		"брянск":           "QWKV",
		"вильнюс":          "R4K4",
		"владивосток":      "P7T0",
		"владикавказ":      "P6LQ",
		"владимир":         "RDLD",
		"волгоград":        "Q4LP",
		"вологда":          "RVLC",
		"воронеж":          "QMLA",
		"горно-алтайск":    "QOP6",
		"грозный":          "P8LT",
		"душанбе":          "OFNQ",
		"екатеринбург":     "RHN2",
		"ереван":           "OPLQ",
		"иваново":          "RILF",
		"ижевск":           "RHMG",
		"иркутск":          "QQQP",
		"йошкар-ола":       "RGM0",
		"казань":           "RBM3",
		"калининград":      "R4JQ",
		"калуга":           "R3L1",
		"кемерово":         "R8P6",
		"киев":             "QFKK",
		"киров":            "RSM5",
		"кишинёв":          "PUKF",
		"кострома":         "RNLF",
		"краснодар":        "PIL9",
		"красноярск":       "RCPR",
		"курган":           "R9NG",
		"курск":            "QML1",
		"кызыл":            "QMPV",
		"липецк":           "QSLB",
		"магадан":          "RXUK",
		"майкоп":           "PGLC",
		"махачкала":        "P6LY",
		"минеральныеводы":  "PDLL",
		"минск":            "R0KB",
		"москва":           "RAL5",
		"мурманск":         "TIKR",
		"набережные челны": "RAMD",
		"назрань":          "P7LQ",
		"нальчик":          "P9LN",
		"нижний новгород":  "RELO",
		"новгород":         "RRKM",
		"новосибирск":      "R6OX",
		"омск":             "R6O4",
		"оренбург":         "QNML",
		"орёл":             "QUL0",
		"пенза":            "QVLR",
		"пермь":            "ROMP",
		"петрозаводск":     "SBKV",
		"петропавловск-камчатский": "QUV8",
		"псков":           "RNKD",
		"рига":            "RIK0",
		"ростов-на-дону":  "PVLB",
		"рязань":          "R4LB",
		"салехард":        "T3NK",
		"самара":          "QVM6",
		"санкт-петербург": "S0KJ",
		"саранск":         "R1LS",
		"саратов":         "QLLU",
		"симферополь":     "PIKU",
		"смоленск":        "R5KO",
		"сочи":            "PALB",
		"ставрополь":      "PILI",
		"станциявосток":   "53QX",
		"станциямирный":   "73PR",
		"сыктывкар":       "SAM8",
		"таллин":          "RXK2",
		"тамбов":          "QSLG",
		"ташкент":         "OWNS",
		"тбилиси":         "OYLQ",
		"тверь":           "REL0",
		"тольятти":        "QXM4",
		"томск":           "RFP3",
		"тула":            "R1L5",
		"тюмень":          "RJNH",
		"улан-удэ":        "QNQZ",
		"ульяновск":       "R2M1",
		"уфа":             "R4MO",
		"хабаровск":       "Q3T9",
		"ханты-мансийск":  "S6NR",
		"чебоксары":       "RDLY",
		"челябинск":       "R7N4",
		"череповец":       "RVL6",
		"черкесск":        "PDLI",
		"чита":            "QORH",
		"элиста":          "PQLP",
		"южно-сахалинск":  "PUTW",
		"якутск":          "SCST",
		"ярославль":       "RMLC"}
	a.client.SendChatAction(m.Chat.ID, tbot.ActionTyping)
	msg, datecheck, cityname, desc, url, urldate := "", 0, "", "", "", ""
	switch m.Text {
	case "/week", "/today":
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
					if m.Text == "/week" {
						cityname = " Прогноз на неделю"
					} else {
						cityname = " Прогноз на сегодня"
					}
					if len(val.City.Name) > 10 {
						cityname = val.City.Country + cityname + "\n" + val.City.Name
					} else {
						cityname = val.City.Country + " " + val.City.Name + cityname
					}
					cst := strings.Split(fmt.Sprintf("%s", val.List[0]), " ")
					cdt := strings.Split(cst[len(cst)-4], "-")
					cdate := fmt.Sprintf("%s-%s-%s", cdt[2], cdt[1], cdt[0])
					urldate = fmt.Sprintf("%s%s%s", cdt[0], cdt[1], cdt[2])
					for i := 0; i < len(val.List)-1; i++ {
						fl := strings.Split(fmt.Sprintf("%.2f", val.List[i:(i+1)]), " ")
						st := strings.Split(fmt.Sprintf("%s", val.List[i:(i+1)]), " ")
						dt := strings.Split(st[len(st)-4], "-")
						date := fmt.Sprintf("%s-%s-%s", dt[2], dt[1], dt[0])
						desc = strings.Title(strings.ToLower(st[11]))
						if len(st) == 26 {
							desc = desc + " " + st[12] + " " + st[13]
						}
						if len(st) == 25 {
							desc = desc + " " + st[12]
						}
						if m.Text == "/week" && date != cdate {
							if /*((st[len(st)-3] == "09:00:00" || st[len(st)-3] == "15:00:00" ||
							  st[len(st)-3] == "21:00:00") && datecheck == 0) ||*/
							st[len(st)-3] == "06:00:00" || st[len(st)-3] == "18:00:00" {
								if st[len(st)-3] == "06:00:00" || datecheck == 0 {
									/*st[len(st)-3] == "21:00:00" || st[len(st)-3] == "18:00:00"*/
									datecheck++
									msg = msg + "\n\n> Прогноз на " + date
								}
								msg = msg + fmt.Sprintf("\n\n%s - %s\nТемпература: %s°\nОщущается: %s°\nВетер: %s м/c\n%s.",
									st[len(st)-3], fadvice(fl[4]), strings.TrimLeft(fl[1], "{"), fl[4],
									strings.TrimLeft(fl[14], "{"), desc)
							}
						} else if m.Text == "/today" && date == cdate {
							msg = msg + fmt.Sprintf("\n\n%s - %s\nТемпература: %s°\nОщущается: %s°\nВетер: %s м/c\n%s.",
								st[len(st)-3], fadvice(fl[4]), strings.TrimLeft(fl[1], "{"), fl[4],
								strings.TrimLeft(fl[14], "{"), desc)
						}
					}
					if m.Text == "/today" {
						w, err := owm.NewCurrent("C", "ru", os.Getenv("OWM_API_KEY")) // fahrenheit (imperial) with Russian output
						if err != nil {
							log.Fatalln(err)
						}
						w.CurrentByName(city)
						arr := strings.Split(fmt.Sprintf("", w.Weather), " ")
						for i := 3; i < len(arr)-1; i++ {
							if i == 3 {
								desc = strings.Title(strings.ToLower(arr[i]))
							} else {
								desc = desc + " " + arr[i]
							}
						}
						msg = fmt.Sprintf("%s\n\nСейчас - %s\nТемпература: %.2f°\nОщущается как: %.2f°\nСкорость ветра: %.2f м/c\n%s.",
							cityname, fadvice(fl[4]), w.Main.Temp, w.Main.FeelsLike, w.Wind.Speed, desc) + msg
						if citycodes[strings.ToLower(city)] != "" {
							urldate = citycodes[strings.ToLower(city)] + "_" + urldate
						}
						url = "https://tesis.lebedev.ru/upload_test/files/kp_" + urldate + ".png?bg=1"
					} else {
						msg = cityname + msg
						url = "https://tesis.lebedev.ru/upload_test/files/fc_" + urldate + ".png"
						fmt.Println(msg, url)
					}
				}
			} else {
				msg = fmt.Sprintf("%v", len(val.List))
			}
		}

	default:
		m.Text = strings.TrimRight(m.Text, " .!")
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

func fadvice(stemp string) (advice string) {
	if temp, err := strconv.ParseFloat(stemp, 32); err == nil {
		if temp <= -50.00 {
			advice = "Сиди дома"
		}
		if temp <= -40.00 {
			advice = "-40"
		}
		if temp <= -30.00 {
			advice = "-30"
		}
		if temp <= -20.00 {
			advice = "-20"
		}
		if temp <= -10.00 {
			advice = "-10"
		}
		if temp <= 00.00 {
			advice = "00"
		}
		if temp >= 00.00 {
			advice = "0"
		}
		if temp >= 10.00 {
			advice = "10"
		}
		if temp >= 20.00 {
			advice = "20"
		}
		if temp >= 30.00 {
			advice = "Шорты + футболка"
		}
		if temp >= 40.00 {
			advice = "Одежда не нужна"
		}
		if temp >= 50.00 {
			advice = "Сиди дома"
		}
	}
	return
}
