package main

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/buger/jsonparser"
)

const toWechatTmplPath = "tmpl/alertmanager2wechat"

type Alert struct {
	Status    string
	Summary   string
	Alertname string
	Instance  string
	StartAt   string
}

func ProcessFromAlertManager(bot Bot, rawData []byte) {
	if bot.Target == "wechat" {
		alerts := formatAlert2WechatWork(rawData)
		sendAlert2WechatWork(alerts, bot.Url)
	}
}

func formatAlert2WechatWork(rawData []byte) []Alert {
	alerts := make([]Alert, 0, 5)
	jsonparser.ArrayEach(rawData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		startTime, err := time.Parse(time.RFC3339, GetJsonStrValue(value, "startsAt"))
		msg := Alert{GetJsonStrValue(value, "status"),
			GetJsonStrValue(value, "annotations", "description"),
			GetJsonStrValue(value, "labels", "alertname"),
			GetJsonStrValue(value, "labels", "instance"),
			startTime.Format("01-02 15:04:05")}
		alerts = append(alerts, msg)
	}, "alerts")
	return alerts
}

func sendAlert2WechatWork(alerts []Alert, url string) {
	for i := 0; i < len(alerts); i++ {
		buf := new(bytes.Buffer)
		tmpl, _ := template.ParseFiles(toWechatTmplPath)
		tmpl.Execute(buf, alerts[i])
		http.Post(url,
			"application/json",
			strings.NewReader(buf.String()))
	}
}
