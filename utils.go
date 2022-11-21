package main

type ChatID string

// Recipient returns chat ID (see Recipient interface).
func (i ChatID) Recipient() string {
	return string(i)
}
