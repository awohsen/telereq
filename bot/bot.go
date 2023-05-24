package bot

import (
	"fmt"
	s "github.com/awohsen/telereq/storage"
	db "github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"strconv"
	"strings"
)

type Bot struct {
	*tele.Bot
	*layout.Layout
}

func New(path string, setting tele.Settings) (*Bot, error) {
	lt, err := layout.New(path)
	if err != nil {
		return nil, err
	}

	b, err := tele.NewBot(setting)
	if err != nil {
		return nil, err
	}

	b.Use(lt.Middleware("en", s.GetUserLocale))

	return &Bot{
		Bot:    b,
		Layout: lt,
	}, nil
}

func (b *Bot) Start() {
	b.Handle("/start", b.onStart)
	b.Handle("/creator", b.onCreator)
	b.Handle("/language", b.onLanguage)
	b.Handle("/signout", b.onSignout)

	b.Handle("/add", b.onAdd)
	b.Handle("/del", b.onDel)

	b.Handle(tele.OnChatJoinRequest, b.onJoinRequest)
	b.Handle("/accept", b.onAccept)

	b.Bot.Start()
}

func (b Bot) onStart(c tele.Context) error {
	u := &s.User{}

	err := s.GetUser(u, c.Sender().ID)
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			u := s.NewUser(c.Sender().ID, "normal", "main")

			if s.IsManager(c.Sender().ID) {
				u.Role = "manager"
			}

			err = db.Coll(&s.User{}).Create(u)
			if err != nil {
				return c.Reply(b.Text(c, "err.database"))
			}
			return c.Reply(b.Text(c, "start"))

		default:
			return c.Reply(b.Text(c, "err.database"))
		}
	}
	return c.Reply(b.Text(c, "start"))
}

func (b Bot) onCreator(c tele.Context) error {

	var chats []s.Chat
	coll := db.Coll(&s.Chat{})

	_ = coll.SimpleFind(&chats, bson.M{})

	fmt.Println(chats)

	return c.Send(b.Text(c, "creator"))
}

func (b Bot) onLanguage(c tele.Context) error {
	args := c.Args()

	if len(args) == 1 {
		args[0] = strings.ToLower(args[0])
		switch args[0] {
		case "en", "fa":
			err := s.SetUserLocale(c.Sender().ID, args[0])
			if err != nil {
				return c.Reply(b.Text(c, "err.database"))
			} else {
				return c.Reply(b.Text(c, "language.succeed"))
			}
		default:
			return c.Reply(b.Text(c, "err.language.choose"))
		}
	} else {
		return c.Reply(b.Text(c, "language"))
	}
}

func (b Bot) onSignout(c tele.Context) error {
	u := &s.User{}

	err := db.Coll(u).FindByID(s.USER+strconv.Itoa(int(c.Sender().ID)), u)
	if err != nil {
		return c.Reply(b.Text(c, "err.database"))
	}

	err = db.Coll(u).Delete(u)
	if err != nil {
		return c.Reply(b.Text(c, "err.database"))
	}
	return c.Reply(b.Text(c, "del.succeed"))
}
