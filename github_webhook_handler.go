package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		out, err := exec.Command("git", "pull", "origin", "master").Output()
		if err != nil {
			log.Println("git pull failed: ", err)
			fmt.Fprintf(w, "FAILED")
			return
		}
		log.Println(string(out))
		fmt.Fprintf(w, "ALL OK:"+string(out))
	})

	http.ListenAndServe(":7629", nil)
}
