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
	"gopkg.in/gomail.v2"
	"strings"
)

// Configuration struct
type Configuration struct {
	ProjectName  string   `json:"project_name"`
	Commands     string   `json:"commands"`
	Dev          string   `json:"dev"`
	Master       string   `json:"master"`
	Emails       []string `json:"emails"`
	SlackWebhook string   `json:"slack_webhook"`
	Port         string   `json:"port"`
	EventName    string   `json:"event_name"`
}

type Payload struct {
	ObjectKind string `json:"object_kind"`
	EventName  string `json:"event_name"`
	Ref        string `json:"ref"`
}

// SlackRequestBody is Slack request structure
type SlackRequestBody struct {
	Text string `json:"text"`
}
func Now() string{
	return time.Now().Format("2006-01-02 15:04:05-0700")
}
type Logger interface {
	AddEvent(title string, body string,extra string, istatus int, cmdIndex int, total int)
	GetLog() string
}
type SLogger struct {
	MessageSeparator string
	TitleStart       string
	TitleEnd         string
	GoodColorStart   string
	BadColorStart    string
        ColorEnd         string
	Start            string
	End              string
	SuccessIcon      string
	FailIcon         string
	MessageStart     string
	MessageEnd       string
	Events           []string
}
func (log *SLogger) AddEvent(title string, body string, extra string, istatus int, cmdIndex int, total int) {
	statusMsg := ""
	if istatus == 0 {
		statusMsg = log.SuccessIcon
		title = log.GoodColorStart + title + log.ColorEnd
	}else{
		statusMsg = log.FailIcon
		title = log.BadColorStart + title + log.ColorEnd
	}
	//mtitle := fmt.Sprintf("\n------------ [%d/%d %s] -------------\n",cmdIndex,total,statusMsg)
	//mtitle := fmt.Sprintf("\n------------ [%d/%d %s] -------------\n",cmdIndex,total,statusMsg)
	if extra != "" {
		body += log.BadColorStart + extra + log.ColorEnd
	}
	log.Events = append(log.Events, fmt.Sprintf("%s %s %s [%d/%d] \n %s%s%s",log.MessageSeparator ,  statusMsg, title, cmdIndex, total, log.MessageStart, body ,  log.MessageEnd))
}
func (log *SLogger) GetLog(iproc int) string {
	out := ""
	if iproc == 0 {
		out += log.GoodColorStart + log.TitleStart + "SUCCESS " + log.SuccessIcon + log.TitleEnd + log.ColorEnd
	}else{
		out += log.BadColorStart  + log.TitleStart + "FAIL "    + log.FailIcon    + log.TitleEnd + log.ColorEnd
	}
	out += "\n\n"
	out += log.Start
	for _, msg := range log.Events {
		out += msg
	}
	out += log.End
	return out
}
// GetLoggers will returns loggers for the Email, Slack and the Console
// Console is assumed ansi:
// https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux
//
// Slack doesn't support colors.
//
// Email supports some html
func GetLoggers() []SLogger {
	ml := SLogger {MessageSeparator:"<hr/>",
		      TitleStart: "<h2>",
		      TitleEnd: "</h2>",
		      GoodColorStart:"<font color=green>",
		      BadColorStart:"<font color=red>",
		      ColorEnd:"</font>",
	              Start:"<pre>",
		      End:"</pre>",
		      SuccessIcon:"&#9989;",
		      FailIcon:"&#10060;",
		      MessageStart:"<pre>",
		      MessageEnd:"</pre>",
	              Events:[]string{}}

	sl := SLogger {MessageSeparator:"",
		      TitleStart: "",
		      TitleEnd: "",
		      GoodColorStart:"",
		      BadColorStart:"",
		      ColorEnd:"",
		      Start:"",
		      End:"",
		      SuccessIcon:"✅", // "\x27\x05",
		      FailIcon:"❌",
		      MessageStart:"\n```\n",
		      MessageEnd:"\n```\n",
	              Events:[]string{}}

	cl := SLogger {MessageSeparator:"-------------------------------------",
		      TitleStart: "",
		      TitleEnd: "",
		      GoodColorStart:"\\033[0;32m",
		      BadColorStart:"\\033[0;31m",
		      ColorEnd:"\033[0m",
		      End:"",
		      SuccessIcon:"✅",
		      FailIcon:"❌", // "\x27\x4C",
		      MessageStart:"\n",
		      MessageEnd:"\n",
	              Events:[]string{}}
	loggers := []SLogger {ml, sl, cl}
	return loggers
}
func main() {
	config := GetConfig("./ci.json")
	// fmt.Println(config)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loggers := GetLoggers()
		commands := ReadCommand(config.Commands)
		count:= 0
		iproc := 0
		for cmdIndex, Command := range commands {
			//read json
			cmd := exec.Command("sh", "-c", Command)
			commandOutputBytes, err := cmd.CombinedOutput()
			commandOutput := string(commandOutputBytes)

			//send stdout on email
			istatus := 0
			extra := ""
			if err != nil {
				istatus = 1
				iproc = 1
				extra = "\n" + err.Error()
			}
			for lid := range loggers {
				mtitle := Command
				loggers[lid].AddEvent(mtitle, commandOutput , extra, istatus, cmdIndex + 1, len(commands))
			}
			count++
			if err != nil {
				break
			}

		}
		slackSubject := fmt.Sprintf("%s: %d/%d ", config.ProjectName, count, len(commands))

		processStatusMessages := []string{"ALL OK ✅ ","FAIL ❌ "}
		subject := fmt.Sprintf("%s: %s %d/%d ", config.ProjectName, processStatusMessages[iproc], count, len(commands))
		emailSubject := subject + Now()

		// Send message to participants of the project:
		SendSlackNotification(config.SlackWebhook, slackSubject + "\n\n" + loggers[1].GetLog(iproc))
		SendEmailNotification2(emailSubject,loggers[0].GetLog(iproc),config.Emails)

		// For the logs
		log.Println(loggers[2].GetLog(iproc))

		// For the web user
		fmt.Fprintf(w, loggers[0].GetLog(iproc))

	})

	http.ListenAndServe(":"+config.Port, nil)
}

func PayloadParser(r *http.Request) Payload {

	decoder := json.NewDecoder(r.Body)
	payload := Payload{}
	err := decoder.Decode(&payload)

	if err != nil {
		panic(err)
	}
	return payload

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

	slackBody, err := json.Marshal(SlackRequestBody{Text: msg})
	if err != nil {
		return err
	}
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

func SendEmailNotification2(subject string, body string, to []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "leadrepository@gmail.com")
	m.SetHeader("To", strings.Join(to, ","))
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("FROM_EMAIL"), os.Getenv("PASS_EMAIL"))

	if err := d.DialAndSend(m); err != nil {
		log.Println("Sending the email failed:" + err.Error())
		return err
	}
	return nil
}

