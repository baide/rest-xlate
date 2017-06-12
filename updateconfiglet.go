// update configlet
package main

import (
	"encoding/json"
)

type configlet struct {
	Config		string	`json:"config"`
	Key			string	`json:"key"`
	Name 		string	`json:"name"`
}

var updateConfigletURL = "/cvpservice/configlet/updateConfiglet.do"

func updateConfiglet(key string, name string, config string) []byte {
	cookie := login()
	confupdate := config
	tempUpdate := configlet{confupdate, key, name}
	updateBody, _ := json.Marshal(tempUpdate)
	response := postCVP(baseURL + updateConfigletURL, cookie, updateBody)
	_ =logout(cookie)
	return response
}
