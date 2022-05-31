package models

import "time"

type User struct {
	UUID     string `json:"uuid"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Auth struct {
	UUID      string     `json:"uuid"`
	UserUUID  string     `json:"user_uuid"`
	Token     string     `json:"token"`
	CreatedAt *time.Time `json:"created_at"`
}

const (
	TypeText  = "text"
	TypeFile  = "file"
	TypeLogin = "login"
	TypeCard  = "card"
)

type Record struct {
	UUID      string     `json:"uuid"`
	UserUUID  string     `json:"user_uuid"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Data      []byte     `json:"data"`
	Comment   string     `json:"comment"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type RecordField int

const (
	RecordFieldUnknown RecordField = iota
	RecordFieldName
	RecordFieldData
	RecordFieldComment
)

func (d RecordField) String() string {
	return [...]string{"Unknown", "Name", "Data", "Comment"}[d]
}

//const (
//	ActionCreate = "create"
//	ActionUpdate = "update"
//	ActionDelete = "delete"
//)

type Event struct {
	UUID       string    `json:"uuid"`
	UserUUID   string    `json:"user_uuid"`
	RecordUUID string    `json:"record_uuid"`
	Date       time.Time `json:"date"`
	Action     string    `json:"action"`
	Data       []byte    `json:"data"`
}

type FileDataType struct {
	Filename string
	Contents []byte
}

type LoginDataType struct {
	Login    string
	Password string
}

type BankCardDataType struct {
	Number   string
	Holder   string
	ExpMonth string
	ExpYear  string
	CVV      string
}
