package main

import (
	"fmt"
	"github.com/awohsen/telereq/bot"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/joho/godotenv"
	db "github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	_ = godotenv.Load()

	err := db.SetDefaultConfig(nil, "telereq", options.Client().ApplyURI("mongodb://127.0.0.1:27017"))

	if err != nil {
		fmt.Println(err)
	}

	pref := tele.Settings{
		Token:     os.Getenv("BOT_TOKEN"),
		Poller:    &tele.LongPoller{Timeout: 30 * time.Second},
		ParseMode: tele.ModeHTML,
	}

	b, err := bot.New("bot.yml", pref)
	if err != nil {
		panic(err)
	}

	b.Start()
}
