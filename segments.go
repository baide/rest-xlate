//get segments(vlans) from named configlet
package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"strconv"
)

type netSegment struct {
	Segment			jsonSegment	`json:"segment"`
}

type netSegments struct {
	Segments			[]jsonSegment	`json:"segments"`
}

type jsonSegment struct {
	Name			string		`json:"id"`
	VLAN			int			`json:"segmentation_id"`
}

func segments(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var response netSegments
		response.Segments = getSegments(container)
		temp, _ := json.Marshal(response)
		fmt.Fprintln(w, string(temp))
	}
	if r.Method == "POST" {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		newVlan(&b)
	}
}

func segment(w http.ResponseWriter, r *http.Request) {
	components := strings.Split(r.URL.Path, "/")
	p := components[len(components)-1]
	if r.Method == "GET" {
		var response netSegment
		response.Segment = getSegment(container, p)
		temp, _ := json.Marshal(response)
		fmt.Fprintln(w, string(temp))
	}
	if r.Method == "PUT" {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		trunkChange(&b, r.URL.Path)
	}
}

func getSegment(container string, vlanTarget string) jsonSegment {
	var result jsonSegment
	_, _, config := getNamedConfiglet(container + "-vlans")
	vlans := strings.Split(config, "!")
	for _, vlan := range vlans {
		if strings.Contains(vlan, vlanTarget) {
			lines := strings.Split(vlan, "\n")
			for _, line := range lines {
				if strings.Contains(line, "vlan ") {
					result.VLAN, _ = strconv.Atoi(strings.Split(line, " ")[1])
				}
				if strings.Contains(line, "name") {
					//c := strings.Split(line, " ")
					//result.Name = strings.Join(c[2:len(c)], " ")
					result.Name = strings.TrimSpace(strings.Replace(line, "name ", "", 1))
				}
			}
		}
	}
	return result
}

func getSegments (container string) []jsonSegment {
	var segments []jsonSegment 
	_, _, config := getNamedConfiglet(container + "-vlans")
	vlans := strings.Split(config, "!")
	for _, vlan := range vlans {
		var temp jsonSegment
		if strings.Contains(vlan, "vlan") {
			lines := strings.Split(vlan, "\n")
			for _, line := range lines {
				if strings.Contains(line, "vlan ") {
					temp.VLAN, _ = strconv.Atoi(strings.Split(line, " ")[1])
				}
				if strings.Contains(line, "name") {
					temp.Name = strings.TrimSpace(strings.Replace(line, "name ", "", 1))
				}
			}
			segments = append(segments, temp)
		}
	}
	return segments
}

func newVlan(request *[]byte)  {
	key, name, config := getNamedConfiglet(container + "-vlans")
	//_, _, config := getNamedConfiglet(container + "-vlans")
	var v jsonSegment
	_ = json.Unmarshal(*request, &v)
	vlans := strings.Split(config, "!")
	newVlan := "vlan " + strconv.Itoa(v.VLAN) + "\n\t" + "name " + v.Name
	var newConfig []string
	vlanRange := strings.Split(strings.TrimSpace(strings.Replace(vlans[1], " allowed range ", "", 1)), "-")
	lowerRange, _ := strconv.Atoi(vlanRange[0])
	upperRange, _ := strconv.Atoi(vlanRange[1])
	newConfig = append(newConfig, "! " + strings.TrimSpace(vlans[1]))
	previous := 0
	current := 0
	if (lowerRange <= v.VLAN && upperRange >= v.VLAN) {
		for i, vlan := range vlans {
			vlan = strings.TrimSpace(vlan)
			lines := strings.Split(vlan, "\n")
			for _, line := range lines {
				if strings.Contains(line, "vlan ") {
					current, _ = strconv.Atoi(strings.Split(line, " ")[1])
					if (previous < v.VLAN && v.VLAN < current) {
						newConfig = append(newConfig, newVlan)
						newConfig = append(newConfig, vlan)
					} else if (current < v.VLAN || current > v.VLAN) {
						newConfig = append(newConfig, vlan)
						if i == len(vlans)-1 {
							newConfig = append(newConfig, newVlan)
						}
					}
				}
			}
		}
		config = strings.Join(newConfig, "\n!\n")
		//fmt.Println("\n" + config)
		_ = updateConfiglet(key, name, config)
	}
}
				







