package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
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
		id := RandStr(36)

		bot := Bot{source, target, url, id}
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Update(func(tx *bolt.Tx) error {
			b, _ := tx.CreateBucketIfNotExists([]byte(dbBucket))
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(bot)
			err = b.Put([]byte(bot.Id), buf.Bytes())
			return err
		})
		defer db.Close()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, "WebHook的地址为：POST /bots/"+bot.Id)
	})

	// 获取所有WebHook机器人的配置
	r.GET("/bots", func(c *gin.Context) {
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		bots := make([]Bot, 0, 5)
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(dbBucket))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				var bot Bot
				buf := bytes.NewBuffer(v)
				enc := gob.NewDecoder(buf)
				err := enc.Decode(&bot)
				if err != nil {
					log.Fatal(err)
				}
				bots = append(bots, bot)
			}
			return err
		})
		defer db.Close()
		str, _ := json.Marshal(bots)
		c.String(http.StatusOK, string(str))
	})

	// 删除WebHook机器人配置
	r.DELETE("bots/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(dbBucket))
			err := b.Delete([]byte(id))
			return err
		})
		defer db.Close()
		c.String(http.StatusOK, "id = "+id+" 已删除")
	})

	// 处理WebHook内容及推送消息到Target
	r.POST("/bots/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		var bot Bot
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(dbBucket))
			v := b.Get([]byte(id))
			buf := bytes.NewBuffer(v)
			enc := gob.NewDecoder(buf)
			err := enc.Decode(&bot)
			return err
		})
		defer db.Close()
		if err != nil {
			c.String(http.StatusNotFound, "未找到id为"+id+"的数据")
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
