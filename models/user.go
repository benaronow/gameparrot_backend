package models

type User struct {
    UID string `bson:"uid" json:"uid"`
    Email string `bson:"email" json:"email"`
    Online bool `json:"online"`
}