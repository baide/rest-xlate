//handle login and logoff to CVP
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"encoding/base64"
	//"fmt"
	"io/ioutil"
)

type authInfo struct { //json struct for authentication
	UserId 		string	`json:"userId"`
	Password 	string	`json:"password"`
}

type authResp struct {
	Username	string `json:"username"`
}

var authURL = "/web/login/authenticate.do"
var logoutURL = "/cvpservice/login/logout.do"

func login(c *authInfo) *http.Cookie {
	auth, _ := json.Marshal(c)
	
	req, _ := http.NewRequest("POST", baseURL + authURL, bytes.NewBuffer(auth)) 
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var a authResp
	err := json.Unmarshal(body, &a)
    if err != nil || a.Username == "" { 
    	t := &http.Cookie{}
    	return t
    }
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
	
func extractCred(w http.ResponseWriter, r *http.Request, c *authInfo)  {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		http.Error(w, "Not Authorized", 401)
		return
	}
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		http.Error(w, "Not Authorized", 401)
		return
	}
	t := strings.Split(string(b), ":")
	if len(t) != 2 {
		return
	}
	c.UserId = t[0]
	c.Password = t[1]	
}
	