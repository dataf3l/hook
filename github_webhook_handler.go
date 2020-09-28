package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Configuration struct
type Configuration struct {
	Commands     string
	Dev          string
	Master       string
	Emails       []string
	SlackWebhook string
}

// SlackRequestBody is Slack request structure
type SlackRequestBody struct {
	Text string `json:"text"`
}

func main() {
	config := GetConfig("./ci.json")
	// fmt.Println(config)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Commands := ReadCommand(config.Commands)
		for _, Command := range Commands {
			cmd := exec.Command("sh", "-c", Command)
			CommandOutput, err := cmd.CombinedOutput()
			if err != nil {
				log.Println(Command, err)
				fmt.Fprintf(w, "FAILED for Command : "+string(Command))
				// time.Now().String() to get server time
				SendSlackNotification(config.SlackWebhook, "FAILED for Command : "+string(Command)+" At: "+time.Now().String())
				return
			}
			log.Println(string(CommandOutput))
			fmt.Fprintf(w, "ALL OK:\n"+string(CommandOutput))
		}
	})

	http.ListenAndServe(":7629", nil)
}

// ReadCommand read commands from text file, which will get executed in whole program
func ReadCommand(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var instructions []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		instructions = append(instructions, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return instructions
}

// GetConfig get configuration to be used in this program
func GetConfig(filePath string) Configuration {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}
	// fmt.Println(configuration.Users)
	return configuration
}

// SendSlackNotification will post to an 'Incoming Webook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func SendSlackNotification(webhookURL string, msg string) error {

	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}
