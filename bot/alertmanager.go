package bot

import (
	"bytes"
	"text/template"
	"time"

	"github.com/buger/jsonparser"
)

type Alert struct {
	Status    string
	Summary   string
	Alertname string
	Instance  string
	StartAt   string
}

func (a Alert) Format() string {
	buf := new(bytes.Buffer)
	tmpl, _ := template.ParseFiles("tmpl/alertmanager/alert2wechat.tmpl")
	tmpl.Execute(buf, a)
	return buf.String()
}

func ProcessFromAlertManager(s Store, rawData []byte) {
	if s.Target == wechatTarget {
		alerts := formatAlert2WechatWork(rawData)
		Send2Wechat(s, alerts)
	}
}

func formatAlert2WechatWork(rawData []byte) []interface{} {
	alerts := make([]interface{}, 0, 5)
	jsonparser.ArrayEach(rawData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		o := JsonObj{value}
		startTime, err := time.Parse(time.RFC3339, o.GetStr("startsAt"))
		msg := Alert{o.GetStr("status"),
			o.GetStr("annotations", "description"),
			o.GetStr("labels", "alertname"),
			o.GetStr("labels", "instance"),
			startTime.Format("01-02 15:04:05")}
		alerts = append(alerts, msg)
	}, "alerts")
	return alerts
}
