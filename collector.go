package main

import "github.com/gin-gonic/gin"
import "github.com/tucnak/telebot"
import "github.com/jinzhu/gorm"
import _ "github.com/lib/pq"
import "github.com/TomiHiltunen/geohash-golang"
import "net/http"
import "fmt"
import "os"
import "encoding/json"
import "./model"

var config model.Configuration
var db gorm.DB
var tgbot *telebot.Bot

func main() {
	// Config
	config_file, _ := os.Open("config.json")
	config_decoder := json.NewDecoder(config_file)
	if config_err := config_decoder.Decode(&config); config_err != nil {
		fmt.Println("Config decode error:", config_err)
	}

	// Database init
	if db_init, db_err := gorm.Open("postgres", config.Database); db_err != nil {
		fmt.Println("Database init error:", db_err)
	} else {
		db = db_init
	}
	db.Exec("CREATE EXTENSION IF NOT EXISTS hstore")
	db.AutoMigrate(&model.ChatEntity{}, &model.ActionEntity{}, &model.Portal{}, &model.Subscription{})
	db.LogMode(true)

	// Telegram init
	if tg_init, tg_err := telebot.NewBot(config.TelegramAPI); tg_err != nil {
		fmt.Println("Telegram API error:", tg_err)
	} else {
		tgbot = tg_init
	}

	srv_web()
}

func srv_web() {
	router := gin.Default()

	router.GET("/api/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Chat logger
/*	router.POST("/api/log/chat", func(c *gin.Context) {
		//tgchat := tgbot.Chat{ID: 69640640, Type: "private"}

		content := []byte(c.PostForm("data"))
		fmt.Println(c.PostForm("data"))
		var chat model.ChatEntity
		er := json.Unmarshal(content, &chat)
		if er != nil {
			c.String(http.StatusOK, "error")
			panic(er)
		} else {
			//tgbot.SendMessage(tgchat, chat.Date.Format(time.RFC3339)+": "+chat.Text, nil)
			db.NewRecord(chat)
			db.Create(&chat)
			c.String(http.StatusOK, "ok")
		}
	})*/

	// Actions logger
	router.POST("/api/log/act", func(c *gin.Context) {
		content := []byte(c.PostForm("data"))
		var action []model.ActionEntity
		er := json.Unmarshal(content, &action)
		if er != nil {
			c.String(http.StatusOK, "error")
			//panic(er) // TODO Log.txt
		} else {
			c.String(http.StatusOK, "ok")
			for _, e := range action {
				//db.LogMode(true)
				res := db.Where("guid = ?", e.Guid).First(&e)
				if res.RecordNotFound() == true {
					go QueueActions(e)
				} else {
					//fmt.Println("Founded and skipped")
				}
				c.String(http.StatusOK, "ok")
				//db.LogMode(false)
			}
		}
	})

	// Portals general logger
	router.POST("/api/portals/push", func(c *gin.Context) {
		content := []byte(c.PostForm("data"))
		var portal model.Portal
		er := json.Unmarshal(content, &portal)
		if er != nil {
			c.String(http.StatusOK, "error")
			//panic(er) // TODO Log.txt
		} else {
			c.String(http.StatusOK, "ok")
			res := db.Where("guid = ?", portal.Guid).First(&portal)
			if res.RecordNotFound() == true {
				go QueuePortal(portal)
			} else {
				//fmt.Println("Founded and skipped")
			}
			c.String(http.StatusOK, "ok")
			//db.LogMode(false)
		}
	})

	// Portals detalis logger
/*	router.POST("/api/portals/details", func(c *gin.Context) {
		content := []byte(c.PostForm("data"))
		var action []model.ActionEntity
		er := json.Unmarshal(content, &action)
		if er != nil {
			c.String(http.StatusOK, "error")
			//panic(er) // TODO Log.txt
		} else {
			c.String(http.StatusOK, "ok")
			for _, e := range action {
				//db.LogMode(true)
				res := db.Where("guid = ?", e.Guid).First(&e)
				if res.RecordNotFound() == true {
					go QueueActions(e)
				} else {
					//fmt.Println("Founded and skipped")
				}
				c.String(http.StatusOK, "ok")
				//db.LogMode(false)
			}
		}
	})*/
	router.Run(":8080")
}

func QueueActions(e model.ActionEntity)  {
	e.Geohash = geohash.Encode(e.Portal1.LatE6/1000000, e.Portal1.LngE6/1000000)
	e.Portal1.Geohash = geohash.Encode(e.Portal1.LatE6/1000000, e.Portal1.LngE6/1000000)
	e.P1Data, _ = json.Marshal(e.Portal1)
	fmt.Println("==========" + e.Portal2.Name)
	//if e.Portal2.Name != "" {
		//fmt.Println("e.Portal2 not nil, creating json:")
		//fmt.Println(e.Portal2)
		e.Portal2.Geohash = geohash.Encode(e.Portal2.LatE6/1000000, e.Portal2.LngE6/1000000)
		e.P2Data, _ = json.Marshal(e.Portal2)
	//}
	res := db.Create(&e)
	// Check for duplicates
	if res.Error != nil {
		fmt.Println("Double!") // DEBUG
	} else {
		// Send alert for followers
		var x[] model.Subscription
		db.Where("? ~ followers.fval", e.Geohash).Find(&x)
		// TODO: Check for duplicates alerts (eg. Jarvis viruses)
		for _, q := range x {
			tgchat := telebot.Chat{ID: q.Tg_id, Type: q.Tg_type}
			tgbot.SendMessage(tgchat, model.ActionToText(e), &telebot.SendOptions{ParseMode: telebot.ModeHTML, DisableWebPagePreview: true})
		}
	}
}

func QueuePortal(e model.Portal)  {
	e.Geohash = geohash.Encode(e.LatE6/1000000, e.LngE6/1000000)
	e.ModData, _ = json.Marshal(e.Mods)
	e.ResData, _ = json.Marshal(e.Resonators)
	res := db.Create(&e)
	// Check for duplicates
	if res.Error != nil {
		fmt.Println("Double!") // DEBUG
	} else {
		fmt.Println("Portal added!")
		// Send alert for followers
	}
}
