package storage

import (
	"strconv"

	db "github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const CHAT = "CHAT_"

type Chat struct {
	db.DateFields `bson:",inline"`
	ID            string  `json:"id" bson:"_id,omitempty"`
	Owner         string  `json:"owner" bson:"owner"`
	Requests      []int64 `json:"requests" bson:"requests"`
	//Links          []string `json:"links" bson:"links"`
}

func (m *Chat) PrepareID(id interface{}) (interface{}, error) {
	return id, nil
}

func (m *Chat) GetID() interface{} {
	return m.ID
}

func (m *Chat) SetID(id interface{}) {
	m.ID = id.(string)
}

func NewChat(id int64, owner int64) *Chat {
	return &Chat{
		ID:       CHAT + strconv.Itoa(int(id)),
		Owner:    USER + strconv.Itoa(int(owner)),
		Requests: []int64{},
	}
}

func GetChat(c *Chat, id int64) error {
	return db.Coll(c).FindByID(CHAT+strconv.Itoa(int(id)), c)
}

func DelChat(c *Chat) error {
	return db.Coll(c).Delete(c)
}

func AppendRequest(c int64, u int64) (result *mongo.UpdateResult, err error) {
	ctx := db.Ctx()
	return db.Coll(&Chat{}).UpdateByID(
		ctx,
		CHAT+strconv.Itoa(int(c)),
		bson.M{operator.AddToSet: bson.M{"requests": u}},
	)
}

func RemoveRequest(c int64, u int64) (result *mongo.UpdateResult, err error) {
	ctx := db.Ctx()
	return db.Coll(&Chat{}).UpdateByID(
		ctx,
		CHAT+strconv.Itoa(int(c)),
		bson.M{operator.Pull: bson.M{"requests": u}},
	)
}
