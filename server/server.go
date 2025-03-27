package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/FromZeroDev/loki_telegram_alert/common"
	"github.com/FromZeroDev/loki_telegram_alert/sndmsstg"
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /send", http.HandlerFunc(sendMessage))
	return mux
}

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

/*
The Alertmanager will send HTTP POST requests in the following JSON format to the configured endpoint:

	{
	  "version": "4",
	  "groupKey": <string>,              // key identifying the group of alerts (e.g. to deduplicate)
	  "truncatedAlerts": <int>,          // how many alerts have been truncated due to "max_alerts"
	  "status": "<resolved|firing>",
	  "receiver": <string>,
	  "groupLabels": <object>,
	  "commonLabels": <object>,field
	  "commonAnnotations": <object>,
	  "externalURL": <string>,           // backlink to the Alertmanager.
	  "alerts": [
	    {
	      "status": "<resolved|firing>",
	      "labels": <object>,
	      "annotations": <object>,
	      "startsAt": "<rfc3339>",
	      "endsAt": "<rfc3339>",
	      "generatorURL": <string>,      // identifies the entity that caused the alert
	      "fingerprint": <string>        // fingerprint to identify the alert
	    },
	    ...
	  ]
	}
*/
type GroupAlert struct {
	Alerts []Alert `json:"alerts"`
	Status string  `json:"status"`
}

type errors []error

func (errs errors) Error() string {
	b := strings.Builder{}
	for _, err := range errs {
		b.WriteString(err.Error())
	}
	return b.String()
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("reading body:", err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	buff := bytes.NewBuffer(b)

	var group GroupAlert
	dec := json.NewDecoder(buff)

	err = dec.Decode(&group)
	if err != nil {
		log.Println("error decoding alerts:", err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	if group.Status == "resolved" {
		w.WriteHeader(200)
		return
	}

	go send(group.Alerts)

	w.WriteHeader(200)
}

func send(alerts []Alert) {
	var errs errors = []error{}
	for _, a := range alerts {
		message := createMessage(a)
		err := sendTelegramMessage(a.Labels["job"], message)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len([]error(errs)) != 0 {
		log.Println("sending messages:", errs.Error())
	}
}

func createMessage(alert Alert) string {
	return alert.Annotations["message"]
}

const MPGBotErrorsGroup int64 = -1002080666885
const CTMBotErrorsGroup int64 = -1002331174364

func sendTelegramMessage(job string, message string) error {
	if !common.Config.SendFronted && job == "cutrans_frontend" {
		return nil
	}

	if strings.Contains(job, "ctm") {
		return sndmsstg.New().SendMessage(sndmsstg.SendMessage{
			Text:   message,
			ChatID: CTMBotErrorsGroup,
		})
	} else {
		return sndmsstg.New().SendMessage(sndmsstg.SendMessage{
			Text:   message,
			ChatID: MPGBotErrorsGroup,
		})
	}
}
