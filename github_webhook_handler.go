package main

import (
        "fmt"
        "net/http"
	"os/exec"
	"log"
)

func main() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                out, err := exec.Command("git", "pull","origin","master").Output()
		if err != nil {
                	log.Println("git pull failed: ", err)
                	fmt.Fprintf(w, "FAILED")
			return
		}
		log.Println(out)
                fmt.Fprintf(w, "ALL OK")
        })

        http.ListenAndServe(":7629", nil)
}
