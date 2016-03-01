package model

import "github.com/satori/go.uuid"
import "time"

type ChatEntity struct {
	Guid   uuid.UUID  `gorm:"primary_key"`   // "guid":"fbdfc6c416324ad59d31709d1876fd67.d"
	Team   int64      `json:"team"`   // "team":[1,0]
	Date   time.Time  `json:"date"`   // "date":"2016-02-25T09:09:49.166Z"
	Text   string     `json:"text"`
	Center Coordinate `json:"center"` // "center":{"lat":59.40390391923957,"lng":56.83176040649414}
}

type Coordinate struct {
	Lat float64
	Lng float64
}

type ActionEntity struct {
	Guid       uuid.UUID    `gorm:"primary_key"`
	Date       time.Time    `json:"date"`
	Player     string       `json:"player"`
	Team       int64        `json:"team"`
	Action     string       `json:"action"`
	ObjectType string       `json:"objectType"`
	Portal1    PortalEntity
	Portal2    PortalEntity
	Extra      string       `json:"extra"`
	Geohash    string
}

type PortalEntity struct {
	Guid    uuid.UUID `json:"guid"`
	Name    string    `json:"name"`
	Plain   string    `json:"plain"`
	Team    string    `json:"team"`
	Address string    `json:"address"`
	LatE6   float64   `json:"latE6"`
	LngE6   float64   `json:"lngE6"`
}

type Follower struct {
	Id      int64	  `gorm:"primary_key"`
	Fid     int64
	Ftype   string
	Fval    string
}
