package main

import "github.com/gin-gonic/gin"
import "github.com/tucnak/telebot"
import "github.com/jinzhu/gorm"
import _ "github.com/lib/pq"
import "github.com/TomiHiltunen/geohash-golang"
import "net/http"
import "fmt"
import "os"
import "strconv"
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

	/*db.CreateTable(&ChatEntity{})
	db.CreateTable(&ActionEntity{})
	db.CreateTable(&PortalEntity{})
	db.CreateTable(&Follower{})*/

	// Chat entity
	/*router.POST("/api/log/chat", func(c *gin.Context) {
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

	// Actions array {}
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
	router.Run(":8080")
}

func QueueActions(e model.ActionEntity)  {
	fmt.Println("RecordNotFound")
	e.Geohash = geohash.Encode(e.Portal1.LatE6/1000000, e.Portal1.LngE6/1000000)
	res := db.Create(&e)
	// Check for duplicates
	if res.Error != nil {
		fmt.Println("Double!") // DEBUG
	} else {
		// Send alert for followers
		var x[] model.Follower
		db.Where("? ~ followers.fval", e.Geohash).Find(&x)
		// TODO: Check for duplicates alerts (eg. Jarvis viruses)
		for _, q := range x {
			tgchat := telebot.Chat{ID: q.Fid, Type: q.Ftype}
			tgbot.SendMessage(tgchat, ActionToText(e), &telebot.SendOptions{ParseMode: telebot.ModeHTML, DisableWebPagePreview: true})
		}
	}
}

func ActionToText(e model.ActionEntity) string {
	var text string
	if (e.Team == 0) {
		text = "Энлайт"
	} else {
		text = "Резист"
	}
	text = text + " @"+e.Player
	if (e.Action == "captured") {
		text = text + " захватил портал " + PortalLink(e.Portal1)
	}
	if (e.Action == "create") {
		if (e.ObjectType == "resonator") {
			text = text + " вставил рез в " + PortalLink(e.Portal1)
		}
		if (e.ObjectType == "link") {
			text = text + " создал линк " + PortalLink(e.Portal1) + " <-> " + PortalLink(e.Portal2)
		}
		if (e.ObjectType == "field") {
			text = text + " создал поле @" + PortalLink(e.Portal1) + " +" + e.Extra + "MUs"
		}
		if (e.ObjectType == "fracker") {
			text = text + " вставил фракер в " + PortalLink(e.Portal1)
		}
	}
	if (e.Action == "destroy") {
		if (e.ObjectType == "resonator") {
			text = text + " сломал рез в " + PortalLink(e.Portal1)
		}
		if (e.ObjectType == "link") {
			text = text + " сломал линк " + PortalLink(e.Portal1) + " <-> " + PortalLink(e.Portal2)
		}
		if (e.ObjectType == "field") {
			text = text + " сломал поле @" + PortalLink(e.Portal1) + " -" + e.Extra + "MUs"
		}
	}
	//text = text + " в " + e.Date.Format("18:04")
	return text
}

func PortalLink(e model.PortalEntity) string {
	// 59.409593,56.792797&z=17&pll=59.409593,56.792797
	var text string = "<a href=\"https://www.ingress.com/intel?ll=" + FloatToString(e.LatE6/1000000) + "," + FloatToString(e.LngE6/1000000) + "&z=16&pll=" + FloatToString(e.LatE6/1000000) + "," + FloatToString(e.LngE6/1000000) +"\">" + e.Name + "</a>"
	return text
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}