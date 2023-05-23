package main

import (
	"fmt"
	"gopkg.in/telebot.v3/layout"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/joho/godotenv"
	db "github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	s "github.com/awohsen/telereq/storage"
)

var lt *layout.Layout

func main() {
	_ = godotenv.Load(".env")

	err := db.SetDefaultConfig(nil, "telereq", options.Client().ApplyURI("mongodb://127.0.0.1:27017"))

	if err != nil {
		fmt.Println(err)
	}

	pref := tele.Settings{
		Token:     os.Getenv("BOT_TOKEN"),
		Poller:    &tele.LongPoller{Timeout: 30 * time.Second},
		ParseMode: tele.ModeHTML,
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		fmt.Println(err)
		return
	}

	lt, err = layout.New("bot.yml")
	if err != nil {
		panic(err)
	}

	b.Use(lt.Middleware("en", s.GetUserLocale))

	b.Handle("/creator", func(c tele.Context) error {

		var chats []s.Chat
		coll := db.Coll(&s.Chat{})

		_ = coll.SimpleFind(&chats, bson.M{})

		fmt.Println(chats)

		return c.Send(lt.Text(c, "creator"))
	})

	b.Handle("/start", func(c tele.Context) error {
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
					return c.Reply(lt.Text(c, "err.database"))
				}
				return c.Reply(lt.Text(c, "start"))

			default:
				return c.Reply(lt.Text(c, "err.database"))
			}
		}
		return c.Reply(lt.Text(c, "start"))
	})

	b.Handle("/language", func(c tele.Context) error {
		args := c.Args()

		if len(args) == 1 {
			args[0] = strings.ToLower(args[0])
			switch args[0] {
			case "en", "fa":
				err := s.SetUserLocale(c.Sender().ID, args[0])
				if err != nil {
					return c.Reply(lt.Text(c, "err.database"))
				} else {
					return c.Reply(lt.Text(c, "language.succeed"))
				}
			default:
				return c.Reply(lt.Text(c, "err.language.choose"))
			}
		} else {
			return c.Reply(lt.Text(c, "language"))
		}
	})

	b.Handle("/signout", func(c tele.Context) error {
		u := &s.User{}

		err := db.Coll(u).FindByID(s.USER+strconv.Itoa(int(c.Sender().ID)), u)
		if err != nil {
			return c.Reply(lt.Text(c, "err.database"))
		}

		err = db.Coll(u).Delete(u)
		if err != nil {
			return c.Reply(lt.Text(c, "err.database"))
		}
		return c.Reply(lt.Text(c, "del.succeed"))
	})

	b.Handle("/add", func(c tele.Context) error {
		args := c.Args()

		if len(args) >= 1 {
			for _, arg := range args {
				if len(args) > 1 {
					_ = c.Send(strings.Replace(lt.Text(c, "chat.processing"), "%chat%", arg, 1))
				}

				chat, err := b.ChatByUsername(arg)

				if err != nil {
					switch err {
					case tele.ErrChatNotFound:
						_ = c.Reply(lt.Text(c, "err.chat_not_found"))
					default:
						_ = c.Reply(lt.Text(c, "err.database"))
					}
					continue
				}

				u, err := b.ChatMemberOf(chat, c.Sender())

				if err != nil {
					switch err {
					case tele.ErrChatNotFound:
						_ = c.Reply(lt.Text(c, "err.chat_not_found"))
					default:
						_ = c.Reply(lt.Text(c, "err.database"))
					}
					continue
				}

				if u.Role == tele.Creator || u.Role == tele.Administrator {
					existingChat := &s.Chat{}

					err := s.GetChat(existingChat, chat.ID)
					if err != nil {
						switch err {
						case mongo.ErrNoDocuments:
							newChat := s.NewChat(chat.ID, c.Sender().ID)
							err = db.Coll(&s.Chat{}).Create(newChat)
							if err != nil {
								return c.Reply(lt.Text(c, "err.database"))
							}
							_ = c.Reply(lt.Text(c, "add.succeed"))
						default:
							return c.Reply(lt.Text(c, "err.database"))
						}
						continue
					}

					if s.USER+strconv.Itoa(int(c.Sender().ID)) != existingChat.Owner {
						// it is ok for chat owner to revoke access of admins over bot
						if u.Role != tele.Creator {
							// other admin only can revoke if the user who submitted chat; is no more admin there
							if u.Role == tele.Administrator {
								oldAdminID := strings.TrimPrefix(existingChat.Owner, s.USER)
								oldAdmin, err := b.ChatMemberOf(chat, ChatID(oldAdminID))

								if err != nil {
									switch err {
									case tele.ErrChatNotFound:
										_ = c.Reply(lt.Text(c, "err.chat_not_found"))
									default:
										_ = c.Reply(lt.Text(c, "err.database"))
									}
									continue
								}
								// normal admins can't revoke owner or other admins access
								if oldAdmin.Role == tele.Creator || oldAdmin.Role == tele.Administrator {
									_ = c.Reply(lt.Text(c, "err.del.not_enough_rights"))
									continue
								}
							}
						}
						err = s.DelChat(existingChat)
						if err != nil {
							return c.Reply(lt.Text(c, "err.database"))
						}

						newChat := s.NewChat(chat.ID, c.Sender().ID)
						err = db.Coll(&s.Chat{}).Create(newChat)
						if err != nil {
							return c.Reply(lt.Text(c, "err.database"))
						}

						_ = c.Reply(lt.Text(c, "add.succeed"))
					} else {
						_ = c.Reply(lt.Text(c, "err.add.chat_exist"))
						continue
					}
				} else {
					_ = c.Reply(lt.Text(c, "err.add.not_enough_rights"))
				}
			}
		} else {
			return c.Reply(lt.Text(c, "add"))
		}
		return nil
	})
	b.Handle("/del", func(c tele.Context) error {
		args := c.Args()

		if len(args) >= 1 {
			for _, arg := range args {
				if len(args) > 1 {
					_ = c.Send(strings.Replace(lt.Text(c, "chat.processing"), "%chat%", arg, 1))
				}
				chat, err := b.ChatByUsername(arg)

				if err != nil {
					switch err {
					case tele.ErrChatNotFound:
						_ = c.Reply(lt.Text(c, "err.chat_not_found"))
					default:
						_ = c.Reply(lt.Text(c, "err.database"))
					}
					continue
				}

				existingChat := &s.Chat{}

				err = s.GetChat(existingChat, chat.ID)
				if err != nil {
					switch err {
					case mongo.ErrNoDocuments:
						_ = c.Reply(lt.Text(c, "err.del.chat_not_found"))
					default:
						return c.Reply(lt.Text(c, "err.database"))
					}
					continue
				}

				// someone trying to revoke access
				if s.USER+strconv.Itoa(int(c.Sender().ID)) != existingChat.Owner {
					u, err := b.ChatMemberOf(chat, c.Sender())

					if err != nil {
						switch err {
						case tele.ErrChatNotFound:
							_ = c.Reply(lt.Text(c, "err.chat_not_found"))
						default:
							_ = c.Reply(lt.Text(c, "err.database"))
						}
						continue
					}

					// it is ok for chat owner to revoke access of admins over bot
					if u.Role != tele.Creator {
						// other admin only can revoke if the user who submitted chat; is no more admin there
						if u.Role == tele.Administrator {
							oldAdminID := strings.TrimPrefix(existingChat.Owner, s.USER)
							oldAdmin, err := b.ChatMemberOf(chat, ChatID(oldAdminID))

							if err != nil {
								switch err {
								case tele.ErrChatNotFound:
									_ = c.Reply(lt.Text(c, "err.chat_not_found"))
								default:
									_ = c.Reply(lt.Text(c, "err.database"))
								}
								continue
							}
							// normal admins can't revoke owner or other admins access
							if oldAdmin.Role == tele.Creator || oldAdmin.Role == tele.Administrator {
								_ = c.Reply(lt.Text(c, "err.del.not_enough_rights"))
								continue
							}
						}
					}
				}
				err = s.DelChat(existingChat)
				if err != nil {
					return c.Reply(lt.Text(c, "err.database"))
				}
				_ = c.Reply(lt.Text(c, "del.succeed"))
			}
		} else {
			return c.Reply(lt.Text(c, "del"))
		}
		return nil
	})

	b.Handle(tele.OnChatJoinRequest, func(c tele.Context) error {
		_, err := s.AppendRequest(c.ChatJoinRequest().Chat.ID, c.ChatJoinRequest().Sender.ID)
		if err != nil {
			fmt.Println(err)
		}
		return nil
	})

	b.Handle("/accept", func(c tele.Context) error {
		args := c.Args()
		if len(args) == 2 {
			chatID, _ := strconv.ParseInt(args[0], 10, 64)

			chat := &s.Chat{}
			opt := &options.FindOneOptions{}

			switch args[1] {
			case "all", "al", "a":
			default:
				count, err := strconv.ParseInt(args[1], 10, 64)

				if err == nil {
					opt.Projection = bson.M{"requests": bson.M{operator.Slice: count}}
				} else {
					return c.Reply(lt.Text(c, "accept"))
				}
			}

			err := db.Coll(chat).FindByID(s.CHAT+strconv.Itoa(int(chatID)), chat, opt)

			if err != nil {
				switch err {
				case mongo.ErrNoDocuments:
					return c.Reply(lt.Text(c, "err.accept.chat_not_found"))
				default:
					return c.Reply(lt.Text(c, "err.database"))
				}
			}

			if s.USER+strconv.Itoa(int(c.Sender().ID)) != chat.Owner {
				return c.Reply(lt.Text(c, "err.accept.not_enough_rights"))
			}

			if len(chat.Requests) >= 1 {
				start := time.Now()
				succeeded, failed := 0, 0
				for _, user := range chat.Requests {
					err := b.ApproveJoinRequest(ChatID(args[0]), &tele.User{ID: user})

					if err != nil {
						failed++

						switch err.Error() {
						case ErrAlreadyParticipant.Error():
						case ErrChannelsTooMuch.Error():
						case ErrUserChannelsTooMuch.Error():
						case tele.ErrUserIsDeactivated.Error():
						case ErrHideRequesterMissing.Error():
						default:
							fmt.Println(err)
							continue
						}
					} else {
						succeeded++
					}

					_, _ = s.RemoveRequest(chatID, user) // we do want to save some as failed, but let keep it simple for now
				}

				return c.Reply("✅ به %cs% درخواست در %t% با موفقیت پاسخ داده شد.\n\n👤 اعضای چت: ca (+ca-cb)\n\n⚠️ از call درخواست، تعداد cf (%f) با خطا مواجه شد!" + time.Since(start).String())
			}

		} else {
			return c.Reply(lt.Text(c, "accept"))
		}
		return nil
	})

	b.Start()
}
