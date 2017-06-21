//get trunks(interfaces) from named configlet
package main

import (
	"fmt"
	"strings"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"unicode"
	"strconv"
	"sort"
)

type netTrunk struct {
	Port			jsonTrunk	`json:"trunk"`
}

type netTrunks struct {
	Ports			[]jsonTrunk	`json:"trunks"`
}

type jsonTrunk struct {
	Name			string		`json:"port_id"`
	State			bool		`json:"admin_state_up"`
	Descr			string		`json:"description"`
	VLAN			[]int		`json:"sub_ports"`
}

type requestTrunk struct {
	Port			jsonTrunkII	`json:"trunk"`
}

type jsonTrunkII struct {
	Name			string		`json:"port_id"`
	State			*bool		`json:"admin_state_up"`
	Descr			string		`json:"description"`
	VLAN			[]int		`json:"sub_ports"`
}


func getTrunk(PortName string, c *authInfo) jsonTrunk {
	var result jsonTrunk
	switchName := strings.Split(PortName, ":")[0]
	result.Name = PortName
	_, _, config := getNamedConfiglet(switchName + "-ports", c)
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
				if strings.Contains(line, "switchport trunk vlan allowed") { 
					newline := strings.Replace(line, "switchport trunk vlan allowed", "", 1)
					words := strings.Split(newline, ",")
					for _, word := range words {
						word = strings.TrimSpace(word)
						if isInt(word) {
							i, _ := strconv.Atoi(word)
							result.VLAN = append(result.VLAN, i)
						} else if strings.Contains(word, "-") {
							start, _ := strconv.Atoi(strings.Split(word, "-")[0])
							end, _ := strconv.Atoi(strings.Split(word, "-")[1])
							for i := start; i <= end; i++ {
								result.VLAN = append(result.VLAN, i)
							}
						}
					}
				}
			}	
		}
		sort.Ints(result.VLAN)
	}
	return result
}

func getTrunks (switchNames []string, c *authInfo) []jsonTrunk {
	var trunkPorts []jsonTrunk 
	for _, s := range switchNames {
		_, _, config := getNamedConfiglet(s + "-ports", c)
		ports := strings.Split(config, "!")
		for _, port := range ports {
			var temp jsonTrunk
			lines := strings.Split(port, "\n")
			state := true
			if strings.Contains(port, "switchport mode trunk") {
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
					if strings.Contains(line, "switchport trunk vlan allowed") { 
						newline := strings.Replace(line, "switchport trunk vlan allowed", "", 1)
						words := strings.Split(newline, ",")
						for _, word := range words {
							word = strings.TrimSpace(word)
							if isInt(word) {
								i, _ := strconv.Atoi(word)
								temp.VLAN = append(temp.VLAN, i)
							} else if strings.Contains(word, "-") {
								start, _ := strconv.Atoi(strings.Split(word, "-")[0])
								end, _ := strconv.Atoi(strings.Split(word, "-")[1])
								for i := start; i <= end; i++ {
									temp.VLAN = append(temp.VLAN, i)
								}
							}
						}
					}
				}
			}
			if temp.Name != "" {
				temp.State = state
				sort.Ints(temp.VLAN)
				trunkPorts = append(trunkPorts, temp)
			}
		}
	}
	return trunkPorts
}

func trunks(w http.ResponseWriter, r *http.Request) {
	c := &authInfo{}
	extractCred(w, r, c)
	if r.Method == "GET" {
		key := getContainerKey(container, c)
		switches := getSwitchNames(key, c)
		var response netTrunks
		response.Ports = getTrunks(switches, c)
		temp, _ := json.Marshal(response)
		fmt.Fprintln(w, string(temp))
	}
	if r.Method == "POST" {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		portToTrunk(&b, c)
	}
}

func trunk(w http.ResponseWriter, r *http.Request) {
	c := &authInfo{}
	extractCred(w, r, c)
	components := strings.Split(r.URL.Path, "/")
	p := components[len(components)-1]
	if r.Method == "GET" {
		var response netTrunk
		response.Port = getTrunk(p, c)
		temp, _ := json.Marshal(response)
		fmt.Fprintln(w, string(temp))
	}
	if r.Method == "PUT" {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		trunkChange(&b, r.URL.Path, c)
	}
}

func trunkChangeVLAN(port string, vlan []int, i int) string {
	newport := ""
	if i != 0 {
		newport += "\n" //need CR to separate from previous port in final configlet
	}
	lines := strings.Split(port, "\n")
	for _, line := range lines {
		if !(strings.Contains(strings.ToLower(line), "switchport trunk vlan allowed") || line == "") {
			newport += line + "\n"
		} else if strings.Contains(strings.ToLower(line), "switchport trunk vlan allowed") {
			newport += "\tswitchport trunk vlan allowed" + vlanInttoString(vlan) + "\n"
		}
	}
	return newport
}

func trunkChangeState(port string, portState bool, i int) string {
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

func trunkChangeDescr(port string, descr string, intName string, i int) string {
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

func isInt(s string) bool {
    for _, c := range s {
        if !unicode.IsDigit(c) {
            return false
        }
    }
    return true
}

func vlanInttoString (vlans []int) string {
	var result []string
	count := 0
	for i:= 0; i <= (len(vlans)-1); i++ {
		if i != len(vlans) - 1 {
			if vlans[i] + 1 == vlans[i+1] {
				count++
			} else if count != 0 {
				result = append(result, strconv.Itoa(vlans[i-count]) + "-" + strconv.Itoa(vlans[i]))
				count = 0
			} else {
				result = append(result, strconv.Itoa(vlans[i]))
				count = 0
			}
		} else if count != 0 {
			result = append(result, strconv.Itoa(vlans[i-count]) + "-" + strconv.Itoa(vlans[i]))
		} else {
			result = append(result, strconv.Itoa(vlans[i]))
		}
	}
	return strings.Join(result, ",")
}

func portChangetoTrunk(port string, vlan []int) string {
	newport := ""
	lines := strings.Split(port, "\n")
	for _, line := range lines {
		if strings.Contains(line, "switchport mode access") {
			newport += "\tswitchport mode trunk\n"
		} else if strings.Contains(line, "switchport access vlan") {
			newport += "\tswitchport trunk vlan allowed " + vlanInttoString(vlan) + "\n"
		} else {
			newport += line + "\n"
		}
	}
	return newport
}
 
func trunkChange(request *[]byte, path string, c *authInfo)  {
	var p requestTrunk
	_ = json.Unmarshal(*request, &p)
	b := strings.Split(path, "/")
	port := b[len(b)-1]
	nameComponents := strings.Split(port, ":")
	switchName := nameComponents[0]
	intName := nameComponents[1]
	key, name, config := getNamedConfiglet(switchName + "-ports", c)
	ports := strings.Split(config, "!")
	for i, port := range ports {
		if strings.Contains(port, "interface " + intName) {
			if p.Port.State != nil {
				port = trunkChangeState(port, *p.Port.State, i)
			}
			if len(p.Port.VLAN) > 0 {
				port = trunkChangeVLAN(port, p.Port.VLAN, i)  
			}
			if p.Port.Descr != "" {
				port = trunkChangeDescr(port, p.Port.Descr, intName, i)  
			}
		}
		ports[i] = port
	}
	config = strings.Join(ports, "!")
	_ = updateConfiglet(key, name, config, c)
}

func portToTrunk(request *[]byte, c *authInfo)  {
	var p requestTrunk
	_ = json.Unmarshal(*request, &p)
	b := strings.Split(p.Port.Name, ":")
	switchName := b[0]
	intName := b[1]
	key, name, config := getNamedConfiglet(switchName + "-ports", c)
	ports := strings.Split(config, "!")
	for i, port := range ports {
		if strings.Contains(port, "interface " + intName) {
			if (p.Port.State != nil) && (len(p.Port.VLAN) > 0) {
				port = portChangeState(port, *p.Port.State, i)
				if p.Port.Descr != "" {
					port = portChangeDescr(port, p.Port.Descr, intName, i)
				}
				port = portChangetoTrunk(port, p.Port.VLAN)
			}
		}
		ports[i] = port
	}
	config = strings.Join(ports, "!")
	_ = updateConfiglet(key, name, config, c)
}
