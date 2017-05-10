package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	wat "github.com/joeledstrom/wat-app/wat-client-api-lib"
)

var (
	nick = flag.String("nick", "SMHI-WeatherBot", "Nickname")
	host = flag.String("host", "", "Hostname or IP")
	port = flag.Int("port", 9595, "Port")
)

type WeatherProvider interface {
	GetCurrentTemperature(lat, lon float64) (float64, string, error)
}

func main() {
	flag.Parse()

	if *host == "" {
		flag.PrintDefaults()
		return
	}

	for {
		err := messageRecvLoop()

		if err != nil {
			if _, ok := err.(*wat.NickAlreadyInUse); ok {
				fmt.Println("Nick Already in use. Try another nick")
				return
			}

		}

		log.Printf("Error: %s\n", err)
		log.Println("Retrying/reconnecting in 5")
		time.Sleep(5 * time.Second)
	}
}

func messageRecvLoop() error {
	client := wat.NewClient()
	defer client.Close()
	err := client.Connect(fmt.Sprintf("%s:%d", *host, *port), *nick)

	if err != nil {
		return err
	}

	log.Println("wat-weather-bot listening for messages")

	for {
		msg, err := client.RecvMessage()
		if err != nil {
			return err
		}

		if strings.HasPrefix(msg.Content, "!weather") {

			tempMsg := getTemperatureMessage(msg.Location)

			err = client.SendMessage(wat.ClientMessage{tempMsg})

			if err != nil {
				return err
			}
		}
	}
}

func getTemperatureMessage(loc *wat.Location) string {
	parts := strings.Split(loc.Loc, ",")

	if len(parts) == 2 {
		lat, err := strconv.ParseFloat(parts[0], 64)
		lon, err := strconv.ParseFloat(parts[1], 64)

		if err == nil {
			temp, unit, err := NewSmhiProvider().GetCurrentTemperature(lat, lon)

			if err == nil {
				format := "Current temperature in %s: %.1f °%s"
				return fmt.Sprintf(format, loc.City, temp, unit[:1])
			} else {
				log.Println("Weather fetch failed: ", err)
			}
		}
	}

	return "Failed to fetch weather data. Please try again."
}
