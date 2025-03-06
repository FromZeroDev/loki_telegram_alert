package sndmsstg

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type chat struct {
	ID  int64
	mut *sync.Mutex
}

type SendMessage struct {
	Text                     string `json:"text,omitempty"`
	ChatID                   int64  `json:"chat_id,omitempty"`
	RelayMessageID           int64  `json:"reply_to_message_id,omitempty"`
	AllowSendingWithoutReply bool   `json:"allow_sending_without_reply,omitempty"`
}

func (c chat) sendMessage(sm SendMessage) error {
	c.mut.Lock()
	timer := time.NewTimer(time.Second)
	defer func() {
		<-timer.C
		c.mut.Unlock()
	}()

	b, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	return send(b, "sendMessage")
}

type TelegramSender interface {
	SendMessage(SendMessage) error
}

type Tg struct {
	chats *sync.Map
	wg    *sync.WaitGroup
}

func New() TelegramSender {
	return Tg{
		chats: &sync.Map{},
		wg:    &sync.WaitGroup{},
	}
}

func (tg Tg) SendMessage(sm SendMessage) error {
	v, ok := tg.chats.Load(sm.ChatID)
	if !ok {
		v = chat{ID: sm.ChatID, mut: &sync.Mutex{}}
		tg.chats.Store(sm.ChatID, v)
	}
	for i := 0; i < len(sm.Text); {
		newSm := sm
		start := i
		end := i + 4096
		if len(sm.Text) < end {
			end = len(sm.Text)
		}
		newSm.Text = sm.Text[start:end]
		err := call(v.(chat).sendMessage, newSm)
		if err != nil {
			return err
		}
		i += 4096
	}
	return nil
}

type t interface {
	SendMessage
}

func call[T t](fn func(T) error, p T) error {
	for {
		err := fn(p)
		if err == nil {
			return nil
		}
		if err, ok := err.(TelegramError); ok {
			if err.ErrorCode != 429 {
				return err
			}
			if val, ok := err.Parameters["retry_after"]; ok {
				switch val := val.(type) {
				case json.Number:
					num, err := val.Int64()
					if err != nil {
						return fmt.Errorf("retry_after: %s %w", val, err)
					}
					time.Sleep(time.Duration(num) * time.Second)
				case float64:
					time.Sleep(time.Duration(val) * time.Second)
					continue
				default:
					log.Println("retry after fail conversion: ", err.Parameters["retry_after"])
				}
			}
		}
		return err
	}
}
