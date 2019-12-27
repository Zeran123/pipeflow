package bot

import (
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
	return Format(a, "tmpl/alertmanager/alert2wechat.tmpl")
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
		msg := Alert{o.GetStr("status"),
			o.GetStr("annotations", "description"),
			o.GetStr("labels", "alertname"),
			o.GetStr("labels", "instance"),
			FormatTime(o.GetStr("startsAt"))}
		alerts = append(alerts, msg)
	}, "alerts")
	return alerts
}
