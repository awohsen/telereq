package main

import tele "gopkg.in/telebot.v3"

type ChatID string

// Recipient returns chat ID (see Recipient interface).
func (i ChatID) Recipient() string {
	return string(i)
}

// made pr for this but till then we should implement it
var (
	ErrAlreadyParticipant  = tele.NewError(400, "Bad Request: USER_ALREADY_PARTICIPANT", "Bad Request: USER_ALREADY_PARTICIPANT")
	ErrJoinedChannelsLimit = tele.NewError(400, "Bad Request: CHANNELS_TOO_MUCH", "Bad Request: CHANNELS_TOO_MUCH")
)
