package models

import "time"

type Logs struct {
	Endpoint string      `bson:"endpoint"`
	Request  interface{} `bson:"request"`
	Response interface{} `bson:"response"`
	Meta     interface{} `bson:"meta"`
	LogAt    time.Time   `bson:"log_at"`
}
