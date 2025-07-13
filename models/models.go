package models

type Message struct {
	Message string `bson:"message" json:"message"`
	From string `bson:"from" json:"from"`
	To string `bson:"to" json:"to"`
}

type Interaction struct {
    UID string `bson:"uid" json:"uid"`
    Messages []Message `bson:"messages" json:"messages"`
}

type FriendRequest struct {
	From string `bson:"from" json:"from"`
	To   string `bson:"to" json:"to"`
}

type User struct {
    UID string `bson:"uid" json:"uid"`
    Email string `bson:"email" json:"email"`
    Interactions []Interaction `bson:"interactions" json:"interactions"`
	FriendRequests []FriendRequest `bson:"friend_requests" json:"friend_requests"`
    Online bool `json:"online"`
}

type UpdateType string

const (
	UpdateTypeMessage UpdateType = "message"
	UpdateTypeFriendRequest UpdateType = "friend_request"
	UpdateTypeFriendAccept UpdateType = "friend_accept"
)

type Update struct {
	Type UpdateType `json:"type"`
	From string `json:"from,omitempty"`
	To string `json:"to,omitempty"`
	Message string `json:"message,omitempty"`
	Status []User `json:"status,omitempty"`
}