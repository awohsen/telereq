package main

import (
	"fmt"
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
)

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

	b.Handle("/creator", func(c tele.Context) error {

		var chats []Chat
		coll := db.Coll(&Chat{})

		_ = coll.SimpleFind(&chats, bson.M{})

		fmt.Println(chats)

		return c.Send("🤐 @awohsen")
	})

	b.Handle("/start", func(c tele.Context) error {
		u := &User{}

		err := getUser(u, c.Sender().ID)
		if err != nil {
			switch err {
			case mongo.ErrNoDocuments:
				u := newUser(c.Sender().ID, "normal", "main")

				if isManager(c.Sender().ID) {
					u.Role = "manager"
				}

				err = db.Coll(&User{}).Create(u)
				if err != nil {
					return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
				}
				return c.Reply("🤠 Welcome to robot!")

			default:
				//todo: report error to developers
				return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
			}
		}
		return c.Reply("\U0001FAE5")
	})
	b.Handle("/signout", func(c tele.Context) error {
		u := &User{}

		err := db.Coll(u).FindByID(USER+strconv.Itoa(int(c.Sender().ID)), u)
		if err != nil {
			return c.Reply("❌")
		}

		err = db.Coll(u).Delete(u)
		if err != nil {
			return c.Reply("❌")
		}
		return c.Reply("☑️")
	})

	b.Handle("/add", func(c tele.Context) error {
		args := c.Args()

		if len(args) >= 1 {
			for _, arg := range args {
				if len(args) > 1 {
					_ = c.Send("💬 Processing chat \"<code>" + arg + "</code>\":")
				}

				chat, err := b.ChatByUsername(arg)

				if err != nil {
					switch err {
					case tele.ErrChatNotFound:
						_ = c.Reply("💬 Chat not found! may you check for typos or check if bot is joined to chat or not...")
					default:
						_ = c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
					}
					continue
				}

				u, err := b.ChatMemberOf(chat, c.Sender())

				if err != nil {
					switch err {
					case tele.ErrChatNotFound:
						_ = c.Reply("💬 Chat not found! may you check for typos or check if bot is joined to chat or not...")
					default:
						_ = c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
					}
					continue
				}

				if u.Role == tele.Creator || u.Role == tele.Administrator {
					existingChat := &Chat{}

					err := getChat(existingChat, chat.ID)
					if err != nil {
						switch err {
						case mongo.ErrNoDocuments:
							newchat := newChat(chat.ID, c.Sender().ID)
							err = db.Coll(&Chat{}).Create(newchat)
							if err != nil {
								return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
							}

							_ = c.Reply("✅")
						default:
							return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
						}
						continue
					}

					if USER+strconv.Itoa(int(c.Sender().ID)) != existingChat.Owner {
						// it is ok for chat owner to revoke access of admins over bot
						if u.Role != tele.Creator {
							// other admin only can revoke if the user who submitted chat; is no more admin there
							if u.Role == tele.Administrator {
								oldAdminID := strings.TrimPrefix(existingChat.Owner, USER)
								oldAdmin, err := b.ChatMemberOf(chat, ChatID(oldAdminID))

								if err != nil {
									switch err {
									case tele.ErrChatNotFound:
										_ = c.Reply("💬 Chat not found! may you check for typos or check if bot is joined to chat or not...")
									default:
										_ = c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
									}
									continue
								}
								// normal admins can't revoke owner or other admins access
								if oldAdmin.Role == tele.Creator || oldAdmin.Role == tele.Administrator {
									_ = c.Reply("💬 This chat was registered to another admin, ask the owner to revoke it for you!")
									continue
								}
							}
						}
						err = delChat(existingChat)
						if err != nil {
							return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
						}

						newchat := newChat(chat.ID, c.Sender().ID)
						err = db.Coll(&Chat{}).Create(newchat)
						if err != nil {
							return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
						}

						_ = c.Reply("✅")
					} else {
						_ = c.Reply("💬 This chat was registered before!")
						continue
					}
				} else {
					_ = c.Reply("💬 You don't have the right to register this chat!")
				}
			}
		} else {
			return c.Reply(`💬 By using this command and placing your chat identifier(username or chat id), you can add chats in which you are an admin for further management.

<code>/add {chat}...</code>

❕Remember, to perform this command bot should have required administrator permissions on that chat. 

🔘Examples:
<code>/add -1001234567890</code>
<code>/add @username</code>
👆 Both works

<code>/add -1001234567890 @Durov @TelegramTips</code>
👆 You can place all you're chat at once as well`)
		}
		return nil
	})
	b.Handle("/del", func(c tele.Context) error {
		args := c.Args()

		if len(args) >= 1 {
			for _, arg := range args {
				if len(args) > 1 {
					_ = c.Send("💬 Processing chat \"<code>" + arg + "</code>\":")
				}
				chat, err := b.ChatByUsername(arg)

				if err != nil {
					switch err {
					case tele.ErrChatNotFound:
						_ = c.Reply("💬 Chat not found! may you check for typos or check if bot is joined to chat or not...")
					default:
						_ = c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
					}
					continue
				}

				existingChat := &Chat{}

				err = getChat(existingChat, chat.ID)
				if err != nil {
					switch err {
					case mongo.ErrNoDocuments:
						_ = c.Reply("💬 This chat has not yet registered!")
					default:
						return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
					}
					continue
				}

				// someone trying to revoke access
				if USER+strconv.Itoa(int(c.Sender().ID)) != existingChat.Owner {
					u, err := b.ChatMemberOf(chat, c.Sender())

					if err != nil {
						switch err {
						case tele.ErrChatNotFound:
							_ = c.Reply("💬 Chat not found! may you check for typos or check if bot is joined to chat or not...")
						default:
							_ = c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
						}
						continue
					}

					// it is ok for chat owner to revoke access of admins over bot
					if u.Role != tele.Creator {
						// other admin only can revoke if the user who submitted chat; is no more admin there
						if u.Role == tele.Administrator {
							oldAdminID := strings.TrimPrefix(existingChat.Owner, USER)
							oldAdmin, err := b.ChatMemberOf(chat, ChatID(oldAdminID))

							if err != nil {
								switch err {
								case tele.ErrChatNotFound:
									_ = c.Reply("💬 Chat not found! may you check for typos or check if bot is joined to chat or not...")
								default:
									_ = c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
								}
								continue
							}
							// normal admins can't revoke owner or other admins access
							if oldAdmin.Role == tele.Creator || oldAdmin.Role == tele.Administrator {
								_ = c.Reply("💬 This chat was registered to another admin, ask the owner to revoke it for you!")
								continue
							}
						}
					}
				}
				err = delChat(existingChat)
				if err != nil {
					return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
				}
				_ = c.Reply("☑️")
			}
		} else {
			return c.Reply(`💬 By using this command and placing your chat identifier(username or chat id), you can remove chats that you don't need anymore.

<code>/del {chat}...</code>

❕Remember, by performing this command all your chat settings would get wiped out!

🔘Examples:
<code>/del -1001234567890</code>
<code>/del @username</code>
👆 Both works

<code>/del -1001234567890 @Durov @TelegramTips</code>
👆 You can place all you're chat at once as well`)
		}
		return nil
	})

	b.Handle(tele.OnChatJoinRequest, func(c tele.Context) error {
		_, err := appendRequest(c.ChatJoinRequest().Chat.ID, c.ChatJoinRequest().Sender.ID)
		if err != nil {
			fmt.Println(err)
		}
		return nil
	})

	b.Handle("/accept", func(c tele.Context) error {
		args := c.Args()
		if len(args) == 2 {
			chatID, _ := strconv.ParseInt(args[0], 10, 64)

			chat := &Chat{}
			opt := &options.FindOneOptions{}

			switch args[1] {
			case "all", "al", "a":
			default:
				count, err := strconv.ParseInt(args[1], 10, 64)

				if err == nil {
					opt.Projection = bson.M{"requests": bson.M{operator.Slice: count}}
				} else {
					return c.Reply("waiting for tele.layout") // fixme instruction message
				}
			}

			err := db.Coll(chat).FindByID(CHAT+strconv.Itoa(int(chatID)), chat, opt)

			if err != nil {
				switch err {
				case mongo.ErrNoDocuments:
					return c.Reply("💬 This chat hasn't been added to the bot yet!")
				default:
					return c.Reply("🤕 Error! There was problem in executing your command.\n\n☑️ Please try again later; this was reported to developers...")
				}
			}

			if USER+strconv.Itoa(int(c.Sender().ID)) != chat.Owner {
				return c.Reply("💬 You don't have the right to do that!")
			}

			if len(chat.Requests) >= 1 {
				for _, user := range chat.Requests {
					err := b.ApproveJoinRequest(ChatID(args[0]), &tele.User{ID: user})

					if err != nil {
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

						_, _ = removeRequest(chatID, user) // we do want to save some as failed, but let keep it simple for now
					}
				}
			}
		} else {
			return c.Reply(`💬 By using this command and placing the desired request amount beside your chat identifier(username or chat id), you can accept their join requests to that specified chat.

<code>/accept {chat} {amount}</code>

❕Remember, to perform this command bot should have required administrator permissions on that chat. 

🔘Examples:
<code>/accept -1001234567890 10</code>
👆Accepts 10 join requests in the chat with id <code>-1001234567890.</code>

<code>/accept @username all</code>
👆 Accepts all join requests sent to @username chat.`)
		}
		return nil
	})

	b.Start()
}
