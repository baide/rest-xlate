// authenticate to cvp with GO
package main

import (
	//"fmt"
	"net/http"
	"log"
	"github.com/gorilla/mux"
)



func main() {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/ports", ports).Methods("GET", "POST")
	r.HandleFunc("/ports/{int}", port).Methods("GET", "PUT")
	r.HandleFunc("/trunks", trunks).Methods("GET", "POST")
	r.HandleFunc("/trunks/{int}", trunk).Methods("GET", "PUT")
	r.HandleFunc("/segments", segments).Methods("GET", "POST")
	r.HandleFunc("/segments/{vlan}", segment).Methods("GET")
	//r.HandleFunc("/test-auth", test).Methods("POST")
	
	log.Fatal(http.ListenAndServe("localhost:8000", r))
}


	 
	
	