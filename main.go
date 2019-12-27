package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zeran2048/pipeflow/bot"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// 根据机器人的配置生成Id，用于生成WebHook的地址
	// 配置的内容为：
	// { source: "alertmanager", target: "wechatwork", url: "https://qyapi.weixin.qq.com/cgi-bin/xxx" }
	r.POST("/bot", func(c *gin.Context) {
		raw, _ := c.GetRawData()
		o := bot.JsonObj{Data: raw}
		b := bot.Store{}
		b.Source = o.GetStr("source")
		b.Target = o.GetStr("target")
		b.Url = o.GetStr("url")
		savedBot, err := b.Save()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, "WebHook的地址为：POST /bots/"+savedBot.Id)
	})

	// 获取所有WebHook机器人的配置
	r.GET("/bots", func(c *gin.Context) {
		bots, err := bot.List()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		str, _ := json.Marshal(bots)
		c.String(http.StatusOK, string(str))
	})

	// 删除WebHook机器人配置
	r.DELETE("bot/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		b := bot.Store{}
		b.Id = id
		err := b.Del()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, "id = "+id+" 已删除")
	})

	// 处理WebHook内容及推送消息到Target
	r.POST("/bot/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		b, err := bot.Get(id)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		data, _ := c.GetRawData()
		if b.Source == "alertmanager" {
			bot.ProcessFromAlertManager(b, data)
		} else if b.Source == "gitlab" {
			bot.ProcessFromGitlab(b, data)
		} else if b.Source == "devops" {
			bot.ProcessFromDevOps(b, data)
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
