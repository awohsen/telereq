package bot

import (
	s "github.com/awohsen/telereq/storage"
	db "github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo"
	tele "gopkg.in/telebot.v3"
	"strconv"
	"strings"
)

func (b Bot) onAdd(c tele.Context) error {
	args := c.Args()

	if len(args) >= 1 {
		for _, arg := range args {
			if len(args) > 1 {
				_ = c.Send(strings.Replace(b.Text(c, "chat.processing"), "%chat%", arg, 1))
			}

			chat, err := b.ChatByUsername(arg)

			if err != nil {
				switch err {
				case tele.ErrChatNotFound:
					_ = c.Reply(b.Text(c, "err.chat_not_found"))
				default:
					_ = c.Reply(b.Text(c, "err.database"))
				}
				continue
			}

			u, err := b.ChatMemberOf(chat, c.Sender())

			if err != nil {
				switch err {
				case tele.ErrChatNotFound:
					_ = c.Reply(b.Text(c, "err.chat_not_found"))
				default:
					_ = c.Reply(b.Text(c, "err.database"))
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
							return c.Reply(b.Text(c, "err.database"))
						}
						_ = c.Reply(b.Text(c, "add.succeed"))
					default:
						return c.Reply(b.Text(c, "err.database"))
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
									_ = c.Reply(b.Text(c, "err.chat_not_found"))
								default:
									_ = c.Reply(b.Text(c, "err.database"))
								}
								continue
							}
							// normal admins can't revoke owner or other admins access
							if oldAdmin.Role == tele.Creator || oldAdmin.Role == tele.Administrator {
								_ = c.Reply(b.Text(c, "err.del.not_enough_rights"))
								continue
							}
						}
					}
					err = s.DelChat(existingChat)
					if err != nil {
						return c.Reply(b.Text(c, "err.database"))
					}

					newChat := s.NewChat(chat.ID, c.Sender().ID)
					err = db.Coll(&s.Chat{}).Create(newChat)
					if err != nil {
						return c.Reply(b.Text(c, "err.database"))
					}

					_ = c.Reply(b.Text(c, "add.succeed"))
				} else {
					_ = c.Reply(b.Text(c, "err.add.chat_exist"))
					continue
				}
			} else {
				_ = c.Reply(b.Text(c, "err.add.not_enough_rights"))
			}
		}
	} else {
		return c.Reply(b.Text(c, "add"))
	}
	return nil
}
func (b Bot) onDel(c tele.Context) error {
	args := c.Args()

	if len(args) >= 1 {
		for _, arg := range args {
			if len(args) > 1 {
				_ = c.Send(strings.Replace(b.Text(c, "chat.processing"), "%chat%", arg, 1))
			}
			chat, err := b.ChatByUsername(arg)

			if err != nil {
				switch err {
				case tele.ErrChatNotFound:
					_ = c.Reply(b.Text(c, "err.chat_not_found"))
				default:
					_ = c.Reply(b.Text(c, "err.database"))
				}
				continue
			}

			existingChat := &s.Chat{}

			err = s.GetChat(existingChat, chat.ID)
			if err != nil {
				switch err {
				case mongo.ErrNoDocuments:
					_ = c.Reply(b.Text(c, "err.del.chat_not_found"))
				default:
					return c.Reply(b.Text(c, "err.database"))
				}
				continue
			}

			// someone trying to revoke access
			if s.USER+strconv.Itoa(int(c.Sender().ID)) != existingChat.Owner {
				u, err := b.ChatMemberOf(chat, c.Sender())

				if err != nil {
					switch err {
					case tele.ErrChatNotFound:
						_ = c.Reply(b.Text(c, "err.chat_not_found"))
					default:
						_ = c.Reply(b.Text(c, "err.database"))
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
								_ = c.Reply(b.Text(c, "err.chat_not_found"))
							default:
								_ = c.Reply(b.Text(c, "err.database"))
							}
							continue
						}
						// normal admins can't revoke owner or other admins access
						if oldAdmin.Role == tele.Creator || oldAdmin.Role == tele.Administrator {
							_ = c.Reply(b.Text(c, "err.del.not_enough_rights"))
							continue
						}
					}
				}
			}
			err = s.DelChat(existingChat)
			if err != nil {
				return c.Reply(b.Text(c, "err.database"))
			}
			_ = c.Reply(b.Text(c, "del.succeed"))
		}
	} else {
		return c.Reply(b.Text(c, "del"))
	}
	return nil
}
