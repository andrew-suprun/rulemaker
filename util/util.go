package util

import (
	"encoding/json"
	"log"
)

func ToJson(in interface{}) (res string) {
	if in == nil {
		return ""
	}
	bts, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		log.Panicf("Json marshal error: %v. In: %v", err, in)
	}
	return string(bts)
}
