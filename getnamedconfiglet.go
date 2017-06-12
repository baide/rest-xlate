// get specific configlet
package main

import (
	"encoding/json"
	"fmt"
)

type NamedConfigletResponse struct {
	Key			string	`json:"key"`
	Name 		string	`json:"name"`
	Config		string	`json:"config"`
}

var getNamedConfigletURL = "/cvpservice/configlet/getConfigletByName.do?name="

func getNamedConfiglet(configletName string) (string, string, string) {
	cookie := login()
	body := getCVP(baseURL + getNamedConfigletURL + configletName, cookie)
    var result NamedConfigletResponse
    
    err := json.Unmarshal(body, &result)
    if err != nil { 
    	fmt.Println("json error")
    }
    
    _ =logout(cookie)
    
    return result.Key, result.Name, result.Config
}