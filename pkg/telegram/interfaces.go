package clients

import "encoding/json"

type BotClient interface {
	GetUpdates(timeout, offset int) ([]Update, error)
	SendMessage(int, string) error
	SendPhoto(int, string) error
	SendDocument(int, string) error
	SendPhotoRaw(int, []byte) error
	SendDocumentRaw(int, []byte) error
}

type Response struct {
	Ok     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
}

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageId int    `json:"message_id"`
	Date      int    `json:"date"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type Chat struct {
	Id        int     `json:"id"`
	Type      string  `json:"type"`
	Title     *string `json:"title"`
	Username  *string `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	IsForum   *string `json:"is_forum"`
}