package main

import "github.com/tucnak/telebot"
import "github.com/jinzhu/gorm"
import _ "github.com/lib/pq"
import "github.com/xeonx/timeago"
import "time"
import "fmt"
import "os"
import "strconv"
import "strings"
import "regexp"
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

	// Start services
	srv_bot()
}

func srv_bot() {
	actions := make(map[string]string)
	fmt.Println("Telegram bot started!")
	messages := make(chan telebot.Message)
	tgbot.Listen(messages, 1*time.Second)
	for message := range messages {
		var thash string = message.Chat.Type + strconv.FormatInt(message.Chat.ID, 10)
		var cmd string
		if (actions[thash] != "") {
			cmd = actions[thash] + " " + strings.ToLower(message.Text)
		} else {
			cmd = strings.ToLower(message.Text)
		}
		if cmd == "/whoi" {
			tgbot.SendMessage(message.Chat,
				"Hello, "+message.Sender.FirstName+"! ChatID: " + strconv.FormatInt(message.Chat.ID, 10) + "(" + message.Chat.Type + ")",
				nil)
		} else if (strings.HasPrefix(cmd, "/where") || strings.HasPrefix(cmd, "бот где")) {
			re1, _ := regexp.Compile("[@][A-Za-z0-9]+")
			re2, _ := regexp.Compile("@")
			var w []string = re1.FindStringSubmatch(message.Text)
			if len(w) == 0 {
				tgbot.SendMessage(message.Chat, "Введите имя игрока начиная с @ ", nil)
				actions[thash] = cmd
			} else {
				var player string = w[0]
				player = re2.ReplaceAllString(player, "")
				var entity model.ActionEntity
				res := db.Where("player ilike ?", player).Order("date desc").Limit(1).First(&entity)
				if res.RecordNotFound() == true {
					tgbot.SendMessage(message.Chat, "Я не знаю где "+player,nil)
				} else {
					tgbot.SendMessage(message.Chat,
						entity.Player + " " + entity.Action + " " + entity.ObjectType + " " + timeago.Russian.Format(entity.Date) + ".",
						nil)
				}
				actions[thash] = ""
			}
		} else if (strings.HasPrefix(cmd, "/log") || strings.HasPrefix(strings.ToLower(cmd), "бот логи")) {
			re1, _ := regexp.Compile("[@][A-Za-z0-9]+")
			re2, _ := regexp.Compile("@")
			var w []string = re1.FindStringSubmatch(message.Text)
			if len(w) == 0 {
				tgbot.SendMessage(message.Chat, "Введите имя игрока начиная с @ ", nil)
				actions[thash] = cmd
			} else {
				var player string = w[0]
				player = re2.ReplaceAllString(player, "")
				var entity []model.ActionEntity
				db.Where("player ilike ?", player).Order("date desc").Limit(20).Find(&entity)
				if len(entity) == 0 {
					tgbot.SendMessage(message.Chat, "Я не знаю игрока "+player,nil)
				} else {
					var tgmsg string
					for _, e := range entity {
						tgmsg = tgmsg + e.Player + " " + e.Action + " " + e.ObjectType + " " + timeago.Russian.Format(e.Date) + ".\n"
					}
					tgbot.SendMessage(message.Chat,	tgmsg,nil)
				}
				actions[thash] = ""
			}
		} else if (strings.HasPrefix(cmd, "/subs") || strings.HasPrefix(strings.ToLower(cmd), "бот подписки")) {
			var subs []model.Follower
			db.Where("fid = ? and ftype = ?", message.Chat.ID, message.Chat.Type).Order("fval desc").Find(&subs)
			if len(subs) == 0 {
				tgbot.SendMessage(message.Chat, "У вас нет подписок :(",nil)
			} else {
				var tgmsg string
				for _, e := range subs {
					tgmsg = tgmsg + strconv.FormatInt(e.Id, 10) + ": %" + e.Fval+ "\n"
				}
				tgbot.SendMessage(message.Chat,	tgmsg,nil)
			}
		} else if (strings.HasPrefix(cmd, "/sub") || strings.HasPrefix(strings.ToLower(cmd), "бот подписаться на")) {
			re1, _ := regexp.Compile("[%][A-Za-z0-9]+")
			re2, _ := regexp.Compile("%")
			var w []string = re1.FindStringSubmatch(message.Text)
			if len(w) == 0 {
				tgbot.SendMessage(message.Chat, "Введите %GeoHash участка для отслеживания (например команда /follow %v68 будет отслеживать квадрат включающий в себя Красновишерск, Соликамск и Березники). Для создания геотегов используйте: http://geohash.gofreerange.com/", nil)
				actions[thash] = cmd
			} else {
				var region string = w[0]
				region = re2.ReplaceAllString(region, "")
				follow := model.Follower{Fid: message.Chat.ID, Ftype: message.Chat.Type, Fval: region}
				err := db.Create(&follow)
				if err.Error != nil {
					tgbot.SendMessage(message.Chat, "Произошла какая-то фигня, и я не подписал вас на регион о_О",nil)
				} else {
					var tgmsg string = "Теперь я слежу за происходящим в " + follow.Fval
					tgbot.SendMessage(message.Chat,	tgmsg, nil)
				}
				actions[thash] = ""
			}
		} else if (strings.HasPrefix(cmd, "/unsub") || strings.HasPrefix(strings.ToLower(cmd), "бот отписаться от")) {
			re1, _ := regexp.Compile("[0-9]+")
			//re2, _ := regexp.Compile("!")
			var w []string = re1.FindStringSubmatch(message.Text)
			if len(w) == 0 {
				tgbot.SendMessage(message.Chat, "Для отмены подписки введите номер отслеживания, узнать который можно командой /subs (пример /unsubs 1)", nil)
				actions[thash] = cmd
			} else {
				var sub int64
				sub, _ = strconv.ParseInt(w[0], 10, 64)
				var follow model.Follower
				res := db.Where("id = ? and fid = ? and ftype = ?", sub, message.Chat.ID, message.Chat.Type).Limit(1).First(&follow)
				if res.RecordNotFound() == true {
					tgbot.SendMessage(message.Chat, "Я не вижу подписку с таким номером \"" + w[0] + "\". Узнайте ваши подписки командой /subs.",nil)
				} else {
					res2 := db.Delete(&follow)
					if res2.Error != nil {
						tgbot.SendMessage(message.Chat, "Произошла какая-то фигня, и я не отписал вас от региона о_О",nil)
					} else {
						var tgmsg string = "Подписка "+w[0]+" отменена!"
						tgbot.SendMessage(message.Chat,	tgmsg, nil)
					}
				}
				actions[thash] = ""
			}
		}
	}
}