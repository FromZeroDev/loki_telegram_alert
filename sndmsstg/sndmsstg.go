package sndmsstg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var botApi string

func init() {
	var ok bool
	botApi, ok = os.LookupEnv("BOT_API")
	if !ok {
		panic("BOT_API enviroment variable is not set")
	}
}

func baseURL() string {
	return "https://api.telegram.org/bot" + botApi + "/"
}

type TelegramError struct {
	ErrorCode   uint                   `json:"error_code"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func (err TelegramError) Error() string {
	return err.Description
}

func send(b []byte, method string) error {
	client := http.Client{}
	req, err := http.NewRequest(
		http.MethodPost,
		baseURL()+method,
		bytes.NewBuffer(b),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		var tgErr TelegramError
		if err := json.Unmarshal(body, &tgErr); err != nil {
			return tgErr
		}
		return fmt.Errorf("telegram response status code %d. error %s", res.StatusCode, string(body))
	}
	return nil
}
