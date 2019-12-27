package bot

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type Provider interface {
	Format() string
}

func Send2Wechat(s Store, data []interface{}) {
	for _, any := range data {
		if provider, ok := (any).(Provider); ok {
			strVal := provider.Format()
			fmt.Println(strVal)
			fmt.Println("send to wechat bot : " + s.Url)
			http.Post(s.Url,
				"application/json",
				strings.NewReader(strVal))
		}
	}
}

func Format(e interface{}, tmplPath string) string {
	fmt.Println(e)
	buf := new(bytes.Buffer)
	tmpl, _ := template.ParseFiles(tmplPath)
	tmpl.Execute(buf, e)
	return strings.ReplaceAll(buf.String(), "    ", "")
}
