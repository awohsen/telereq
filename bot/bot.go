package bot

import (
	s "github.com/awohsen/telereq/storage"
	db "github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
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
	_, err := s.GetUser(c.Sender())
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			role := "normal"
			if s.IsManager(c.Sender().ID) {
				role = "manager"
			}

			err := s.NewUser(c.Sender(), role, "main")

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
	return c.Send(b.Text(c, "creator"))
}

func (b Bot) onLanguage(c tele.Context) error {
	args := c.Args()

	if len(args) == 1 {
		args[0] = strings.ToLower(args[0])
		switch args[0] {
		case "en", "fa":
			err := s.SetUserLocale(c.Sender(), args[0])
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

	err := db.Coll(u).FindByID(c.Sender().Recipient(), u)
	if err != nil {
		return c.Reply(b.Text(c, "err.database"))
	}

	err = db.Coll(u).Delete(u)
	if err != nil {
		return c.Reply(b.Text(c, "err.database"))
	}
	return c.Reply(b.Text(c, "del.succeed"))
}
