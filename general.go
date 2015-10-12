package main

import (
	"fmt"
	auth "github.com/abbot/go-http-auth"
	"github.com/vaughan0/go-ini"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func exitWithMessage(message string) {
	fmt.Println(message)
	time.Sleep(time.Second * 3)
	os.Exit(0)
}

func loadConfiguration() map[string]map[string]string {
	config := make(map[string]map[string]string)
	iniFile, err := ini.LoadFile("lifds-cp.ini")

	if err != nil {
		exitWithMessage("Can't read configuration from lifds-cp.ini file.")
	}

	config["lifds"] = iniFile.Section("lifds")
	config["control-panel"] = iniFile.Section("control-panel")

	switch {
	case config["lifds"] == nil:
	case config["lifds"]["lifds-directory"] == "":
	case config["lifds"]["world-id"] == "":
	case config["control-panel"] == nil:
	case config["control-panel"]["port"] == "":
	case config["control-panel"]["address"] == "":
	case config["control-panel"]["server-up-at-start"] == "":
		exitWithMessage("Broken configuration in lifds-cp.ini file.")
	}

	return config
}

func runControlServer() {
	authenticator := auth.NewBasicAuthenticator("localhost", Secret)
	http.HandleFunc("/server", authenticator.Wrap(serverActionsHandler))
	http.HandleFunc("/server/status", authenticator.Wrap(serverStatusHandler))
	http.Handle("/index.html", authenticator.Wrap(indexHandler))
	http.Handle("/", http.FileServer(http.Dir("."+osPathSeparator+"html")))
	http.ListenAndServe(config["control-panel"]["address"]+":"+config["control-panel"]["port"], nil)
}

func processClientAction(clientAction string, w http.ResponseWriter) {
	switch clientAction {
	case "start":
		if exeCmd != nil && exeCmd.ProcessState == nil {
			log.Println("Preventing double start")
			return
		}

		gameSrvCanStartChan <- true
		message := "GETTING UP (MANUALLY)"
		gameSrvStatusChan <- message
		fmt.Fprint(w, "success")
		break
	case "stop":
		if exeCmd != nil && exeCmd.ProcessState != nil {
			log.Println("Preventing double kill")
			return
		}

		gameSrvCanStart = false
		message := "GETTING DOWN (MANUALLY)"
		gameSrvStatusChan <- message
		exeCmd.Process.Kill()
		message = "DOWN (MANUALLY)"
		gameSrvStatusChan <- message
		fmt.Fprint(w, "success")
		break
	default:
		log.Print("Wrong server action received")
		fmt.Fprint(w, "fail")
	}
}

func isTransitState() bool {
	var isTransitState bool

	if strings.Contains(gameSrvStatus, "GETTING") {
		isTransitState = true
	} else {
		isTransitState = true
	}

	return isTransitState
}

func getStatusResponse(debug bool) string {
	return fmt.Sprintf(responseBaseStr, debug, gameSrvStatus, currentSrvVersion, availableSrvVersion, topicVersion)
}

func getAdminPassword() string {
	cpPassword := config["control-panel"]["password"]

	if cpPassword != "" {
		return cpPassword
	}

	compiled, _ := regexp.Compile("<adminPassword>([^<]*)</adminPassword>")

	matches := compiled.FindStringSubmatch(worldCfgContents)

	if len(matches) > 1 {
		return matches[1]
	} else {
		return "password"
	}
}

func Secret(user, realm string) string {
	a := auth.MD5Crypt([]byte(adminPassword), []byte("gjnjVexnjNfr"), []byte("$1$"))
	return string(a)
}
