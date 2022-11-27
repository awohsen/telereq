package main

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

func newUser(id int64, role string, state string) *User {
	return &User{
		ID:    USER + strconv.Itoa(int(id)),
		Role:  role,
		State: state,
	}
}

func getUser(u *User, id int64) error {
	return db.Coll(u).FindByID(USER+strconv.Itoa(int(id)), u)
}

func setUserLocale(id int64, locale string) error {
	u := &User{}

	err := getUser(u, id)
	if err != nil {
		return err
	}

	u.Locale = locale

	return db.Coll(u).Update(u)
}

func getUserLocale(r tele.Recipient) string {
	u := &User{}
	userID, _ := strconv.ParseInt(r.Recipient(), 10, 64)
	err := getUser(u, userID)
	if err != nil {
		return "en" // user does not exist or etc ...
	}
	return u.Locale
}

func isManager(id int64) bool {
	rawManagers := strings.Split(os.Getenv("MANAGERS"), ",")
	managers := make(map[int64]bool)

	for i := 0; i < len(rawManagers); i++ {
		manager, _ := strconv.ParseInt(rawManagers[i], 10, 64)
		managers[manager] = true
	}

	if managers[id] {
		return true
	}
	return false
}

func isAdmin(id int64) bool {
	u := &User{}

	err := getUser(u, id)
	if err != nil {
		return false
	}

	if u.Role == "manager" || u.Role == "admin" {
		return true
	}

	return false
}
