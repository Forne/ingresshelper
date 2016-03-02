package model

import (
	"github.com/satori/go.uuid"
	"time"
	"./jsonb"
	"strconv"
)

type ChatEntity struct {
	Guid      uuid.UUID	`gorm:"primary_key" sql:"type:uuid"`   // "guid":"fbdfc6c416324ad59d31709d1876fd67.d"
	Team      uint		`json:"team"`   // "team":[1,0]
	DateTime  time.Time	`json:"date"`   // "date":"2016-02-25T09:09:49.166Z"
	Text      string	`json:"text"`
	Center    string	`sql:"-",json:"center"` // "center":{"lat":59.40390391923957,"lng":56.83176040649414}
}

type ActionEntity struct {
	Guid	   uuid.UUID	`gorm:"primary_key" sql:"type:uuid"`
	Date       time.Time    `json:"date"`
	Player     string       `json:"player"`
	Team       int64        `json:"team"`
	Action     string       `json:"action"`
	ObjectType string       `json:"objectType"`
	Portal1	   PortalEntity
	P1Data     jsonb.JSONRaw `sql:"type:jsonb"`
	Portal2	   PortalEntity
	P2Data     jsonb.JSONRaw `sql:"type:jsonb"`
	Extra      string       `json:"extra"`
	Geohash    string
}

type PortalEntity struct {
	Name   string    `json:"name"`
	Team    string
	LatE6   float64   `json:"latE6"`
	LngE6   float64   `json:"lngE6"`
	Plain   string    `json:"plain"`
	Address string    `json:"address"`
	Geohash string
}

type Portal struct {
	Guid	uuid.UUID `gorm:"primary_key" sql:"type:uuid"`
	Title   string    `json:"name"`
	Team    string
	LatE6   float64   `json:"latE6"`
	LngE6   float64   `json:"lngE6"`
	Level   int64
	Health  int64
	Owner   string
	Resonators Resonator
	ResData jsonb.JSONRaw `sql:"type:jsonb"`
	Mods    Mod
	ModData jsonb.JSONRaw `sql:"type:jsonb"`
	Image   string
	Mission bool
	Mission50plus bool
	Timestamp int64
	Plain   string    `json:"plain"`
	Address string    `json:"address"`
	Geohash string
}

type Resonator struct {
	Owner   string
	Level   uint
	Energy  uint
}

type Mod struct {
	Owner   string
	Name    string
	Rarity  string
}

type Subscription struct {
	Id        int64	  `gorm:"primary_key"`
	Tg_id     int64
	Tg_type   string
	Type      string
	Value     string
	Params    string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// Actions
func ActionToText(e ActionEntity) string {
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
			text = text + " создал линк " + PortalLink(e.Portal1) + " - " + PortalLink(e.Portal2)
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
			text = text + " сломал линк " + PortalLink(e.Portal1) + " - " + PortalLink(e.Portal2)
		}
		if (e.ObjectType == "field") {
			text = text + " сломал поле @" + PortalLink(e.Portal1) + " -" + e.Extra + "MUs"
		}
	}
	//text = text + " в " + e.Date.Format("18:04")
	return text
}

// Portals
func PortalLink(e Portal) string {
	// 59.409593,56.792797&z=17&pll=59.409593,56.792797
	var text string = "["+e.Name + "](https://www.ingress.com/intel?ll=" + FloatToString(e.LatE6/1000000) + "," + FloatToString(e.LngE6/1000000) + "&z=16&pll=" + FloatToString(e.LatE6/1000000) + "," + FloatToString(e.LngE6/1000000) +")"
	return text
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}