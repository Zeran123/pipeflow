package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
)

const DBPATH = "data.db"
const DBBUCKET = "Bots"

type Alert struct {
	Status    string
	Summary   string
	Alertname string
	Instance  string
	StartAt   string
}

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

		source := getString(data, "source")
		target := getString(data, "target")
		url := getString(data, "url")
		id := RandStr(36)

		bot := Bot{source, target, url, id}
		db, err := bolt.Open(DBPATH, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Update(func(tx *bolt.Tx) error {
			b, _ := tx.CreateBucketIfNotExists([]byte(DBBUCKET))
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
		db, err := bolt.Open(DBPATH, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		bots := make([]Bot, 0, 5)
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DBBUCKET))
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
		db, err := bolt.Open(DBPATH, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DBBUCKET))
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
		db, err := bolt.Open(DBPATH, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DBBUCKET))
			v := b.Get([]byte(id))
			buf := bytes.NewBuffer(v)
			enc := gob.NewDecoder(buf)
			err := enc.Decode(&bot)
			fmt.Println(bot)
			return err
		})
		defer db.Close()
		if err != nil {
			c.String(http.StatusNotFound, "未找到id为"+id+"的数据")
			return
		}
		data, _ := c.GetRawData()
		alerts := make([]Alert, 0, 5)

		jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			startTime, err := time.Parse(time.RFC3339, getString(value, "startsAt"))
			msg := Alert{getString(value, "status"),
				getString(value, "annotations", "description"),
				getString(value, "labels", "alertname"),
				getString(value, "labels", "instance"),
				startTime.Format("01-02 15:04:05")}
			alerts = append(alerts, msg)
		}, "alerts")

		for i := 0; i < len(alerts); i++ {
			buf := new(bytes.Buffer)
			fmt.Println(alerts[i])
			tmpl, _ := template.ParseFiles("tmpl/wechat.tmpl")
			tmpl.Execute(buf, alerts[i])
			fmt.Println(buf)
			http.Post(bot.Url,
				"application/json",
				strings.NewReader(buf.String()))
		}
		c.String(http.StatusOK, "OK")
	})

	return r
}

// 从json中获取指定key的值
func getString(value []byte, keys ...string) string {
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

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run("127.0.0.1:8080")
}
