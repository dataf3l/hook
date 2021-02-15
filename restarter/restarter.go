package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"os/exec"
	"os"
	"os/signal"
	"syscall"
	"strconv"
)
var gDoRestartMe bool
var gPID int
var gFolder string
var gProcessExecutable string

func killProcess(){
	log.Println("KILL")
	syscall.Kill(-gPID, syscall.SIGKILL)
}
func startChildProcess(){
	log.Println("START " + gProcessExecutable + " on " + gFolder)

	os.Chdir(gFolder) // some folder name

	cmd := exec.Command(gProcessExecutable) // whatever sh file
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Println("subprocess failed:",err)
	}
	gPID = cmd.Process.Pid
        err = cmd.Wait()
	if err != nil {
		log.Println("wait:",err)
	}
        fmt.Printf("Command finished with error: %v\n", err)
	log.Printf("Just ran subprocess %d, exiting\n", cmd.Process.Pid)
}
func doRestart() {
	killProcess()
	startChildProcess()
}
func doMonitor(t time.Time){

	log.Printf("Monitor... %t %d \n", gDoRestartMe,gPID)
	if gDoRestartMe {
		log.Println("Restart Requested...")
		// do the restart thing.
		go doRestart()
		gDoRestartMe = false
	}
}

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Status: PID:%d", gPID)
}

func restartHandler(w http.ResponseWriter, r *http.Request) {
	gDoRestartMe = true
	fmt.Fprintf(w, "Restart requested, wait 30 seconds or less.")
}

func startWebServer(port int){
	http.HandleFunc("/v1/status", statusHandler)
	http.HandleFunc("/v1/restart_process", restartHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d",port), nil))
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal, cleaning up...")
		killProcess()
		os.Exit(0)
	}()
}


func main() {
	if len(os.Args) < 4 {
		log.Println("Usage: ./restarter folder_name process_executable port ")
		return
	}
	gFolder = os.Args[1]
	gProcessExecutable = os.Args[2]
	port, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Println("Invalid port, using 7415")
		port = 7415
	}

	gDoRestartMe = false

	SetupCloseHandler()

	go doEvery(5*time.Second, doMonitor)
	go startChildProcess()
	startWebServer(port)
}
// https://gist.github.com/ryanfitz/4191392
// https://superuser.com/questions/140461/using-watch-with-pipes
// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
// https://golangcode.com/handle-ctrl-c-exit-in-terminal/
// https://stackoverflow.com/questions/39508086/golang-exec-background-process-and-get-its-pid
// https://coderwall.com/p/ik5xxa/run-a-subprocess-and-connect-to-it-with-golang

