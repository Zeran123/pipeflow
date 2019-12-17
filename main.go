package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

type Alert struct {
	Status    string
	Summary   string
	Alertname string
	Instance  string
	StartAt   string
}

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	r.POST("/msg", func(c *gin.Context) {
		data, _ := c.GetRawData()

		var f interface{}
		err := json.Unmarshal(data, &f)
		if err == nil {
			m := f.(map[string]interface{})
			arr := m["alerts"].([]interface{})
			alert := arr[0].(map[string]interface{})
			labels := alert["labels"].(map[string]interface{})
			annotations := alert["annotations"].(map[string]interface{})
			alertMsg := Alert{getJsonValue(alert, "status"), getJsonValue(annotations, "description"),
				getJsonValue(labels, "alertname"),
				getJsonValue(labels, "instance"),
				getJsonValue(alert, "startsAt")}
			tmpl, err := template.New("alert").Parse("{ \"msgtype\": \"markdown\", \"markdown\": { \"content\": \"[{{.Status}}] {{.Summary}}\n>AlertName:{{.Alertname}}\n >Instance:{{.Instance}}\n >StartAt:{{.StartAt}}\" } }")
			buf := new(bytes.Buffer)
			if err == nil {
				tmpl.Execute(buf, alertMsg)
			}
			fmt.Println(buf.String())
			http.Post("",
				"application/json",
				strings.NewReader(buf.String()))
		}

		// msg := string(data)
		// fmt.Println(msg)
		c.String(http.StatusOK, "OK")
		// var f interface{}
		// err := json.Unmarshal(data, &f)
		// if err == nil {
		// 	m := f.(map[string]interface{})
		// 	msg := "{\"version\": " + getJsonValue(m, "version") + "}"
		// 	c.String(http.StatusOK, msg)
		// } else {
		// 	fmt.Println(err)
		// }
	})

	return r
}

func getJsonValue(m map[string]interface{}, k string) string {
	switch vv := m[k].(type) {
	case string:
		return vv
	case int:
		return strconv.Itoa(vv)
	default:
		return ""
	}
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run("127.0.0.1:8080")
}
