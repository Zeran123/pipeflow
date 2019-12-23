package bot

import (
	"math/rand"
	"time"

	"github.com/buger/jsonparser"
)

// 从json中获取指定key的值
func GetJsonStrValue(value []byte, keys ...string) string {
	str, _ := jsonparser.GetString(value, keys...)
	return str
}

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
