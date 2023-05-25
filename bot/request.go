package bot

import (
	"fmt"
	s "github.com/awohsen/telereq/storage"
	db "github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	tele "gopkg.in/telebot.v3"
	"strconv"
	"time"
)

func (b Bot) onJoinRequest(c tele.Context) error {
	_, err := s.AppendRequest(c.ChatJoinRequest().Chat.ID, c.ChatJoinRequest().Sender.ID)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func (b Bot) onAccept(c tele.Context) error {
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
				return c.Reply(b.Text(c, "accept"))
			}
		}

		err := db.Coll(chat).FindByID(s.CHAT+strconv.Itoa(int(chatID)), chat, opt)

		if err != nil {
			switch err {
			case mongo.ErrNoDocuments:
				return c.Reply(b.Text(c, "err.accept.chat_not_found"))
			default:
				return c.Reply(b.Text(c, "err.database"))
			}
		}

		if s.USER+strconv.Itoa(int(c.Sender().ID)) != chat.Owner {
			return c.Reply(b.Text(c, "err.accept.not_enough_rights"))
		}

		if len(chat.Requests) >= 1 {
			count := len(chat.Requests)
			succeeded, failed := 0, 0
			targetChat := ChatID(args[0])
			start := time.Now()
			for _, user := range chat.Requests {
				err := b.ApproveJoinRequest(targetChat, &tele.User{ID: user})

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

			p := AnswerRequestResponse{
				Count:      count,
				Time:       time.Since(start).String(),
				Succeed:    succeeded,
				Failed:     failed,
				After:      0,
				Differance: 0,
				FailRatio:  failed / count * 100,
			}

			return c.Reply(b.Text(c, "request_answer.result", p))
		}

	} else {
		return c.Reply(b.Text(c, "accept"))
	}
	return nil
}
