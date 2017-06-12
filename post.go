//CVP Post
package main

import (
	"net/http"
	"io/ioutil"
	"bytes"
)

func postCVP(URL string, cookie *http.Cookie, reqBody []byte) []byte {
	req, _ := http.NewRequest("POST", URL, bytes.NewBuffer(reqBody)) // need to figure this out to make generic
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	response, _ := client.Do(req)
	body, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()
	return body
}	
