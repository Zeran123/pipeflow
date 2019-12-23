package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

const dbPath = "data.db"
const dbBucket = "Bots"

type Bot struct {
	Source string
	Target string
	Url    string
	Id     string
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// 根据机器人的配置生成Id，用于生成WebHook的地址
	// 配置的内容为：
	// { source: "alertmanager", target: "wechatwork", url: "https://qyapi.weixin.qq.com/cgi-bin/xxx" }
	r.POST("/bots", func(c *gin.Context) {
		data, _ := c.GetRawData()

		source := GetJsonStrValue(data, "source")
		target := GetJsonStrValue(data, "target")
		url := GetJsonStrValue(data, "url")

		bot := Bot{source, target, url, ""}
		db := Blot{dbPath, dbBucket}
		savedBot, err := db.save(bot)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, "WebHook的地址为：POST /bots/"+savedBot.Id)
	})

	// 获取所有WebHook机器人的配置
	r.GET("/bots", func(c *gin.Context) {

		db := Blot{dbPath, dbBucket}
		bots, err := db.list()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		str, _ := json.Marshal(bots)
		c.String(http.StatusOK, string(str))
	})

	// 删除WebHook机器人配置
	r.DELETE("bots/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		db := Blot{dbPath, dbBucket}
		err := db.del(id)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, "id = "+id+" 已删除")
	})

	// 处理WebHook内容及推送消息到Target
	r.POST("/bots/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		db := Blot{dbPath, dbBucket}
		bot, err := db.get(id)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		data, _ := c.GetRawData()
		if bot.Source == "alertmanager" {
			ProcessFromAlertManager(bot, data)
		}
		c.String(http.StatusOK, "OK")
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run("127.0.0.1:8080")
}
