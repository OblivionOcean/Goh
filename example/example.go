package main

import (
	"bytes"
	"net/http"

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

	http.ListenAndServe(":8080", nil)
}
