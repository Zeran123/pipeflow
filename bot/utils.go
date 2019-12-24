package bot

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// 生成随机字符串
func RandStr(strlen int) string {
	rand.Seed(time.Now().Unix())
	data := make([]byte, strlen)
	var num int
	for i := 0; i < strlen; i++ {
		num = rand.Intn(57) + 65
		for {
			if num > 90 && num < 97 {
				num = rand.Intn(57) + 65
			} else {
				break
			}
		}
		data[i] = byte(num)
	}
	return string(data)
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
