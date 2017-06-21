// CVP GET
package main

import (
	"net/http"
	"io/ioutil"
)

func getCVP(URL string, cookie *http.Cookie) []byte {
	req, _ := http.NewRequest("GET", URL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
    req.AddCookie(cookie)
    response, _ := client.Do(req)
    body, _ := ioutil.ReadAll(response.Body)
    response.Body.Close()
    
    return body
}