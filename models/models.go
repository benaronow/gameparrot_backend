package models

type Message struct {
	Message string `bson:"message" json:"message"`
	From string `bson:"from" json:"from"`
	To string `bson:"to" json:"to"`
}

type Friend struct {
    UID string `bson:"uid" json:"uid"`
    Messages []Message `bson:"messages" json:"messages"`
}

type User struct {
    UID string `bson:"uid" json:"uid"`
    Email string `bson:"email" json:"email"`
    Friends []Friend `bson:"friends" json:"friends"`
    Online bool `json:"online"`
}

type UpdateType string

const (
	UpdateTypeMessage UpdateType = "message"
	UpdateTypeFriend UpdateType = "friend"
)

type Update struct {
	Type UpdateType `json:"type"`
	From string `json:"from,omitempty"`
	To string `json:"to,omitempty"`
	Message string `json:"message,omitempty"`
	Status []User `json:"status,omitempty"`
}