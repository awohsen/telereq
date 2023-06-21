package storage

import (
	tele "gopkg.in/telebot.v3"
	"os"
	"strconv"
	"strings"

	db "github.com/kamva/mgm/v3"
)

const USER = "USER_"

type User struct {
	db.DateFields `bson:",inline"`
	ID            string `json:"id" bson:"_id,omitempty"`
	Role          string `json:"role" bson:"role"`
	State         string `json:"state" bson:"state"`
	Locale        string `json:"language_code" bson:"language_code"`
}

func (m *User) PrepareID(id interface{}) (interface{}, error) {
	return id, nil
}

func (m *User) GetID() interface{} {
	return m.ID
}

func (m *User) SetID(id interface{}) {
	m.ID = id.(string)
}

func NewUser(user *tele.User, role string, state string) error {
	u := &User{
		ID:    user.Recipient(),
		Role:  role,
		State: state,
	}

	return db.Coll(u).Create(u)
}

func GetUser(user tele.Recipient) (*User, error) {
	u := &User{}
	err := db.Coll(u).FindByID(user.Recipient(), u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func GetUserLocale(r tele.Recipient) string {
	u, err := GetUser(r)
	if err != nil {
		return "en"
	}

	return u.Locale
}

func SetUserLocale(user *tele.User, locale string) error {
	u, err := GetUser(user)
	if err != nil {
		return err
	}

	u.Locale = locale

	return db.Coll(u).Update(u)
}

func GetUserLocale(r tele.Recipient) string {
	u := &User{}
	userID, _ := strconv.ParseInt(r.Recipient(), 10, 64)
	err := GetUser(u, userID)
	if err != nil {
		return "en"
	}
	return u.Locale
}

func IsManager(id int64) bool {
	rawManagers := strings.Split(os.Getenv("MANAGERS"), ",")
	managers := make(map[int64]bool)

	for i := 0; i < len(rawManagers); i++ {
		manager, _ := strconv.ParseInt(rawManagers[i], 10, 64)
		managers[manager] = true
	}

	return managers[id]
}
