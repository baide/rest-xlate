// program params
package main

import (
	"crypto/tls"
	"net/http"
)

var username = "baide"
var password = "arista"
var baseURL = "https://cvp.home.lab"

var tr = &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
	
var client = &http.Client{Transport: tr}

var container = "demo"

var topologyURL = "/cvpservice/provisioning/searchTopology.do?queryParam=changeme&startIndex=0&endIndex=0"	
var netElementListByContainerURL = "/cvpservice//provisioning/getAllNetElementListByContainer.do?nodeId=changeme&ignoreAdd=true&startIndex=0&endIndex=0"
