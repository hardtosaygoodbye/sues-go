package model

import "encoding/gob"

type Course struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Teacher string `json:"teacher"`
	Address string `json:"address"`
	Week    string `json:"week"`
	Time    string `json:"time"`
}

func init() {
	gob.Register(Course{})
}
