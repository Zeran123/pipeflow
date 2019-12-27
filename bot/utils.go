package bot

import (
	"fmt"
	"math/rand"
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

func FormatTime(str string) string {
	time, err := time.Parse(time.RFC3339, str)
	if err == nil {
		return time.Format("01-02 15:04:05")
	} else {
		fmt.Println(err)
		return ""
	}
}
