package main

import (
	"TestWork/src/monitoring"
	"TestWork/src/telegram"
	"bytes"
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Services []Service `yaml:"services"`
	Telegram Telegram  `yaml:"telegram"`
}

type Service struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}
type Telegram struct {
	Token  string `yaml:"token"`
	ChatID int64  `yaml:"chat_id"`
}

var config Config

func main() {
	err := cleanenv.ReadConfig("./src/config/config.yaml", &config)
	if err != nil {
		log.Fatal("Error reading config file: " + err.Error())
	}

	bot := telegram.New(config.Telegram.Token)
	err = bot.SendMessage("Алибаба", config.Telegram.ChatID)
	if err != nil {
		log.Println("Error sending message: " + err.Error())
	}

	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("can't open log.txt file:", err)
	}

	errorsFile, err := os.OpenFile("errorsFile.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("can't open errorsFile.txt file:", err)
	}

	serviceList := map[string][3]int{}

	client := http.Client{
		Timeout: time.Second * 10,
	}

	for {

		for _, service := range config.Services {

			monitoring.AddServiceForMonitor(service.Name, serviceList)

			req, err := http.NewRequest("POST", service.URL, bytes.NewBuffer([]byte(`{}`)))
			if err != nil {
				log.Println(err)
			}
			req.Header.Set("Content-Type", "application/json")

			start := time.Now()
			resp, err := client.Do(req)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					log.Println("Timeout request for service: " + service.Name)
				}
				continue
			}
			resp.Body.Close()
			finish := time.Now().Sub(start)
			_, err = logFile.WriteString(time.Now().String() + " " + service.URL + " " + resp.Status + " " + finish.String() + "\n")
			if err != nil {
				log.Println(err)
			}

			monitoring.CheckCriticalErrors(service.Name, resp.Status, finish, serviceList)

			compare := serviceList[service.Name]
			if compare[0] == 1 || compare[1] >= 3 || compare[2] == 1 {
				message := "В сервисе " + service.Name + " произошла критическая ошибка " + resp.Status + " время ошибки " + start.String()
				err = bot.SendMessage(message, config.Telegram.ChatID)
				if err != nil {
					log.Println(err)
				}
				_, err = errorsFile.WriteString(time.Now().String() + " " + service.URL + " " + resp.Status + " " + finish.String() + "\n")
				if err != nil {
					log.Println(err)
				}
			}

		}
		time.Sleep(5 * time.Minute)
	}
}
