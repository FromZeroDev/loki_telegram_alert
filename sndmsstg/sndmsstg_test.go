package sndmsstg_test

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"testing"

	"github.com/FromZeroDev/loki_telegram_alert/sndmsstg"

	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	err := sndmsstg.New().SendMessage(sndmsstg.SendMessage{
		ChatID: -4125996068,
		Text:   "Hola Ross",
	})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
}

func TestSendMessageLarge(t *testing.T) {
	var generateRandomCharacter = func() byte {
		return byte(32 + rand.UintN(95))
	}
	text := make([]byte, 0, 12096)
	for i := 0; i < 12096; i++ {
		text = append(text, generateRandomCharacter())
	}
	err := sndmsstg.New().SendMessage(sndmsstg.SendMessage{
		ChatID: -4125996068,
		Text:   string(text),
	})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
}

func TestTgParseError(t *testing.T) {
	body := `{"ok":false,"error_code":429,"description":"Too Many Requests: retry after 31",
		"parameters":{"retry_after":31}}`

	{
		var tgErr sndmsstg.TelegramError
		reader := json.NewDecoder(bytes.NewBuffer([]byte(body)))
		reader.UseNumber()
		err := reader.Decode(&tgErr)
		require.NoError(t, err)

		require.True(t, 429 == tgErr.ErrorCode)
		val, ok := tgErr.Parameters["retry_after"].(json.Number)
		require.True(t, ok)
		num, err := val.Int64()
		require.NoError(t, err)
		require.Equal(t, int64(31), num)
		require.True(t, 31 == num)
	}

	{
		var tgErr sndmsstg.TelegramError
		reader := json.NewDecoder(bytes.NewBuffer([]byte(body)))
		err := reader.Decode(&tgErr)
		require.NoError(t, err)

		require.True(t, 429 == tgErr.ErrorCode)
		val, ok := tgErr.Parameters["retry_after"].(float64)
		require.True(t, ok)
		num := int(val)
		require.Equal(t, 31, num)
		require.True(t, 31 == num)
	}
}
