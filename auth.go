//handle login and logoff to CVP
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	//"io/ioutil"
)

type AuthInfo struct { //json struct for authentication
	UserId 		string	`json:"userId"`
	Password 	string	`json:"password"`
}

var authURL = "/web/login/authenticate.do"
var logoutURL = "/cvpservice/login/logout.do"

func login() *http.Cookie {
	tempauth := AuthInfo{username, password}
	auth, _ := json.Marshal(tempauth)
	resp, _ := client.Post(baseURL + authURL, "application/json", bytes.NewBuffer(auth))
    resp.Body.Close()
    cookies := resp.Cookies()
    authcookie := extract_cookie(cookies)
    return authcookie
}

func extract_cookie(cookies []*http.Cookie) *http.Cookie {
	authcookie := &http.Cookie{}
	for _, cookie := range cookies {
		if cookie.Name == "JSESSIONID" {
			authcookie = cookie
		}
	}
	return authcookie
}

func logout(cookie *http.Cookie,) string {
	response := postCVP(baseURL + logoutURL, cookie, nil)
	return string(response)
}	
	
	