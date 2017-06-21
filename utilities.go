package main

import (
	//"fmt"
	"strings"
	"encoding/json"
	//"net/http"
	//"io/ioutil"
	//"strconv"
)

func getContainerKey(containerName string, c *authInfo) string {
	cookie := login(c)
	if cookie.Name == "" {
		return ""
	}
	topologyURL = strings.Replace(topologyURL, "changeme", containerName, 1)
	body := getCVP(baseURL + topologyURL, cookie)
	_ =logout(cookie)
	if cookie.Name == "" {
		return ""
	}
	var topResp topologySearch
	_ = json.Unmarshal(body, &topResp)

	return topResp.ContainerList[0].Key
}

func getSwitchNames(containerKey string, c *authInfo) []string {
	cookie := login(c)
	var switches []string
	if cookie.Name == "" {
		return switches
	}
	netElementListByContainerURL = strings.Replace(netElementListByContainerURL, "changeme", containerKey, 1)
	body := getCVP(baseURL + netElementListByContainerURL, cookie)
	_ =logout(cookie)
	
	var elementList netElementList
	_ = json.Unmarshal(body, &elementList)
	
	for _, element := range elementList.NetElementList {
		switches = append(switches, element.Fqdn)
	}	
	
	return switches
}