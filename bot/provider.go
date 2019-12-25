package bot

import (
	"fmt"
	"net/http"
	"strings"
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
