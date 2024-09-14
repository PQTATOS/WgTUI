package tgclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/PQTATOS/WireTgBot/pkg/telegram"
)


type Client struct {
	HTTPClient   http.Client
	Token        string
}


func New(botToken string) *Client {
	return &Client{
		HTTPClient: http.Client{},
		Token: botToken,
	}
}


func (bot *Client) GetUpdates(timeout, offset int) ([]clients.Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=%d", bot.Token, offset, timeout)
	req, reqErr := http.NewRequest("GET", url, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("GetUpdates: error with request: %w", reqErr)
	}

	resp, respErr := bot.HTTPClient.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("GetUpdates: error with sending request %v: %w", req, respErr)
	}
	defer resp.Body.Close()

	body := clients.Response{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("GetUpdates: while decoding body: %w", err)
	}

	var updates []clients.Update
	if err := json.Unmarshal(body.Result, &updates); err != nil {
		return nil, fmt.Errorf("GetUpdates: while unmarshaling rawMessage: %w", err)
	}

	return updates, nil
}


func (bot *Client) SendMessage(chat_id int, text string) error {
	body := bytes.NewBufferString(
		fmt.Sprintf(`{"chat_id":%d,"parse_mode":"HTML","text":"%s"}`, chat_id, text),
	)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", bot.Token)
	req, reqErr := http.NewRequest("GET", url, body)
	if reqErr != nil {
		return fmt.Errorf("SendMessage: error with request: %w", reqErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))

	resp, respErr := bot.HTTPClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("SendMessage: error with sending request %v: %w", req, respErr)
	}
	defer resp.Body.Close()
	return nil
}


func (bot *Client) SendPhoto(chat_id int, filename string) error {
	err := bot.sendFile(chat_id, filename, "photo", "sendPhoto")
	if err != nil {
		return err
	}
	return nil
}


func (bot *Client) SendDocument(chat_id int, filename string) error {
	err := bot.sendFile(chat_id, filename, "document", "sendDocument")
	if err != nil {
		return err
	}
	return nil
}


func (bot *Client) sendFile(chat_id int, filename,  mediaType, apiReq string) error {
	img, fileErr := os.Open(filename)
	if fileErr != nil {
		return fmt.Errorf("sendFile: error while opening file: %w", fileErr)
	}
	defer img.Close()

	body := &bytes.Buffer{}
	formWrite := multipart.NewWriter(body)
	if err := formWrite.WriteField("chat_id", strconv.Itoa(chat_id)); err != nil {
		return fmt.Errorf("sendFile: error with creating chat_id field: %w", err)
	}
	fw, fwErr := formWrite.CreateFormFile(mediaType, filename)
	if fwErr != nil {
		return fmt.Errorf("sendFile: error with creating file field: %w", fwErr)
	}
	if _, err := io.Copy(fw, img); err != nil {
		return fmt.Errorf("sendFile: error while copying to fieldWriter: %w", err)
	}
	formWrite.Close()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", bot.Token, apiReq)
	req, reqErr := http.NewRequest("GET", url, body)
	if reqErr != nil {
		return fmt.Errorf("sendFile: error with request: %w", reqErr)
	}
	req.Header["Content-Type"] = []string{formWrite.FormDataContentType()}

	resp, respErr := bot.HTTPClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("sendFile: error with sending request %v: %w", req, respErr)
	}
	//reqB,_ := io.ReadAll(resp.Body)
	//slog.Info(string(reqB))
	defer resp.Body.Close()

	return nil
}
