package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/OblivionOcean/Goh/example/complex"
	"github.com/OblivionOcean/Goh/example/template"
)

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
		var userList = []string{
			"Alice",
			"Bob",
			"Tom",
		}

		buffer := new(bytes.Buffer)
		template.UserList("User List", userList, buffer)

		w.Write(buffer.Bytes())
	})

	http.HandleFunc("/page", func(w http.ResponseWriter, req *http.Request) {
		buffer := new(bytes.Buffer)
		complex.Page("Complex Page", "Hello from the complex layout", "/page", true, buffer)
		w.Write(buffer.Bytes())
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
