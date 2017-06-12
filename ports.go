//get ports(interfaces) from named configlet
package main

import (
	"fmt"
	"strings"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strconv"
)

type topologySearch struct {
	Keywordlist		[]string	`json:"keywordList"`
	Total			int			`json:"total"`
	NetElementList	[]string	`json;"netElementList"`
	ContainerList	[]containerList	`json:"containerList"`
}

type containerList struct {
	Name		string	`json:"name"`
	Key			string	`json:"key"`
}

type netElementList struct {
	NetElementList	[]netElement `json:"netElementList"`
}

type netElement struct {
	Fqdn	string `json:"fqdn"`
}

type netPort struct {
	Port			jsonPort	`json:"port"`
}

type netPorts struct {
	Ports			[]jsonPort	`json:"ports"`
}

type jsonPort struct {
	Name			string		`json:"port_id"`
	State			bool		`json:"admin_state_up"`
	Descr			string		`json:"description"`
	VLAN			int			`json:"network_id"`
}

type requestPort struct {
	Port			jsonPortII	`json:"port"`
}

type jsonPortII struct {
	Name			string		`json:"port_id"`
	State			*bool		`json:"admin_state_up"`
	Descr			string		`json:"description"`
	VLAN			int			`json:"network_id"`
}

func getPort(PortName string) jsonPort {
	var result jsonPort
	switchName := strings.Split(PortName, ":")[0]
	//result.Name = strings.Split(PortName, ":")[1]
	result.Name = PortName
	_, _, config := getNamedConfiglet(switchName + "-ports")
	ports := strings.Split(config, "!")
	for _, port := range ports {
		if strings.Contains(port, strings.Split(PortName, ":")[1]) {
			result.State = true
			if strings.Contains(port, "shutdown") {
				result.State = false
			}
			lines := strings.Split(port, "\n")
			for _, line := range lines {
				if strings.Contains(line, "description") {
					c := strings.Split(line, " ")
					result.Descr = strings.Join(c[1:len(c)], " ")
				}
				if strings.Contains(line, "switchport access vlan") { 
					c := strings.Split(line, " ")
					result.VLAN, _ = strconv.Atoi(c[len(c)-1])
				}
			}	
		}
	}
	return result
}

func getPorts (switchNames []string) []jsonPort {
	var switchPorts []jsonPort 
	for _, s := range switchNames {
		_, _, config := getNamedConfiglet(s + "-ports")
		ports := strings.Split(config, "!")
		for _, port := range ports {
			var temp jsonPort
			lines := strings.Split(port, "\n")
			state := true
			if strings.Contains(port, "switchport mode access") {
				for _, line := range lines {
					if strings.Contains(line, "interface") {
						temp.Name = s + ":" + strings.Split(line, " ")[1]
					}
					if strings.Contains(line, "shutdown") {
						state = false
					}
					if strings.Contains(line, "description") {
						c := strings.Split(line, " ")
						temp.Descr = strings.Join(c[1:len(c)], " ")
					}
					if strings.Contains(line, "switchport access vlan") { 
						c := strings.Split(line, " ")
						temp.VLAN, _ = strconv.Atoi(c[len(c)-1])
					}
				}
			}
			if temp.Name != "" {
				temp.State = state
				switchPorts = append(switchPorts, temp)
			}
		}
	}
	return switchPorts
}

func getContainerKey(containerName string) string {
	cookie := login()
	topologyURL = strings.Replace(topologyURL, "changeme", containerName, 1)
	body := getCVP(baseURL + topologyURL, cookie)
	_ =logout(cookie)
	
	var topResp topologySearch
	_ = json.Unmarshal(body, &topResp)

	return topResp.ContainerList[0].Key
}

func getSwitchNames(containerKey string) []string {
	cookie := login()
	netElementListByContainerURL = strings.Replace(netElementListByContainerURL, "changeme", containerKey, 1)
	body := getCVP(baseURL + netElementListByContainerURL, cookie)
	_ =logout(cookie)
	
	var elementList netElementList
	_ = json.Unmarshal(body, &elementList)
	
	var switches []string
	
	for _, element := range elementList.NetElementList {
		switches = append(switches, element.Fqdn)
	}	
	
	return switches
}

func ports(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		key := getContainerKey(container)
		switches := getSwitchNames(key)
		var response netPorts
		response.Ports = getPorts(switches)
		temp, _ := json.Marshal(response)
		fmt.Fprintln(w, string(temp))
	}
	if r.Method == "POST" {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		trunkToPort(&b)
	}
}

func port(w http.ResponseWriter, r *http.Request) {
	components := strings.Split(r.URL.Path, "/")
	p := components[len(components)-1]
	if r.Method == "GET" {
		var response netPort
		response.Port = getPort(p)
		temp, _ := json.Marshal(response)
		fmt.Fprintln(w, string(temp))
	}
	if r.Method == "PUT" {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		portChange(&b, r.URL.Path)
	}
}

func portChangeVLAN(port string, vlan int, i int) string {
	newport := ""
	if i != 0 {
		newport += "\n" //need CR to separate from previous port in final configlet
	}
	lines := strings.Split(port, "\n")
	for _, line := range lines {
		if !(strings.Contains(strings.ToLower(line), "switchport access vlan") || line == "") {
			newport += line + "\n"
		} else if strings.Contains(strings.ToLower(line), "switchport access vlan") {
			newport += "\tswitchport access vlan " + strconv.Itoa(vlan) + "\n"
		}
	}
	return newport
}

func portChangeState(port string, portState bool, i int) string {
	newport := ""
	if i != 0 {
		newport += "\n" //need CR to separate from previous port in final configlet
	}
	lines := strings.Split(port, "\n")
	for _, line := range lines {
		if !(strings.Contains(strings.ToLower(line), "shutdown") || line == "") {
			newport += line + "\n"
		}
	}
	if portState == false {
		newport += "\tshutdown\n"
	}
	return newport
}

func portChangeDescr(port string, descr string, intName string, i int) string {
	newport := ""
	if i != 0 {
		newport += "\n" //need CR to separate from previous port in final configlet
	}
	lines := strings.Split(port, "\n")
	for _, line := range lines {
		if strings.Contains(line, "interface " + intName) {
			newport += line + "\n"
			newport += "\tdescription " + descr + "\n"
		} else if !(strings.Contains(strings.ToLower(line), "description") || line == "") {
			newport += line + "\n"
		}
	}
	return newport
}

func portChangetoAccess(port string, vlan int) string {
	newport := ""
	lines := strings.Split(port, "\n")
	for _, line := range lines {
		if strings.Contains(line, "switchport mode trunk") {
			newport += "\tswitchport mode access\n"
		} else if strings.Contains(line, "switchport trunk vlan") {
			newport += "\tswitchport access vlan " + strconv.Itoa(vlan) + "\n"
		} else  {
			newport += line + "\n"
		}
	}
	return newport
}

func portChange(request *[]byte, path string)  {
	var p requestPort
	_ = json.Unmarshal(*request, &p)
	c := strings.Split(path, "/")
	port := c[len(c)-1]
	nameComponents := strings.Split(port, ":")
	switchName := nameComponents[0]
	intName := nameComponents[1]
	key, name, config := getNamedConfiglet(switchName + "-ports")
	ports := strings.Split(config, "!")
	for i, port := range ports {
		if strings.Contains(port, "interface " + intName) {
			if p.Port.State != nil {
				port = portChangeState(port, *p.Port.State, i)
			}
			if p.Port.VLAN != 0 {
				port = portChangeVLAN(port, p.Port.VLAN, i)  
			}
			if p.Port.Descr != "" {
				port = portChangeDescr(port, p.Port.Descr, intName, i)  
			}
		}
		ports[i] = port
	}
	config = strings.Join(ports, "!")
	_ = updateConfiglet(key, name, config)
}

func trunkToPort(request *[]byte)  {
	var p requestPort
	_ = json.Unmarshal(*request, &p)
	c := strings.Split(p.Port.Name, ":")
	switchName := c[0]
	intName := c[1]
	key, name, config := getNamedConfiglet(switchName + "-ports")
	ports := strings.Split(config, "!")
	for i, port := range ports {
		if strings.Contains(port, "interface " + intName) {
			if p.Port.State != nil && p.Port.VLAN != 0 {
				port = portChangeState(port, *p.Port.State, i)
				if p.Port.Descr != "" {
					port = portChangeDescr(port, p.Port.Descr, intName, i)
				}
				port = portChangetoAccess(port, p.Port.VLAN)
			}
		}
		ports[i] = port
	}
	config = strings.Join(ports, "!")
	_ = updateConfiglet(key, name, config)
}
 