package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

var (
	adminPassword           string
	lastStartTime           time.Time
	lastStopTime            time.Time
	serverToken             string
	exePath                 string
	worldCfgPath            string
	worldCfgContents        string
	exeFileName             string
	exeCmd                  *exec.Cmd
	doneChan                chan error
	gameSrvStatus           string
	currentSrvVersion       string
	availableSrvVersion     string
	gameSrvStatusChan       chan string
	currentSrvVersionChan   chan string
	availableSrvVersionChan chan string
	gameSrvCanStart         bool
	gameSrvCanStartChan     chan bool
	statusPolls             map[int]chan string
	responseBaseStr         string
	config                  map[string]map[string]string
	lifdsDirectory          string
	worldId                 string
	topicVersion            int
	indexHtml               string
	osPathSeparator         string
)

func init() {
	config = loadConfiguration()

	if config["lifds"]["lifds-exe-file-name"] == "" {
		exeFileName = "ddctd_cm_yo_server.exe"
	} else {
		exeFileName = config["lifds"]["lifds-exe-file-name"]
	}

	osPathSeparator = string(os.PathSeparator)
	exePath = config["lifds"]["lifds-directory"] + osPathSeparator + exeFileName
	worldCfgPath = config["lifds"]["lifds-directory"] + osPathSeparator + "config" + osPathSeparator + "world_" + config["lifds"]["world-id"] + ".xml"
	worldCfgContentsBuff, err := ioutil.ReadFile(worldCfgPath)
	worldCfgContents = string(worldCfgContentsBuff)

	if err != nil {
		log.Printf("Can't open config file \"%s\"", worldCfgPath)
	}

	indexHtmlFile, err := ioutil.ReadFile("html/index.html")
	indexHtml = string(indexHtmlFile)

	if err != nil {
		log.Printf("Can't open index.html file")
	}

	adminPassword = getAdminPassword()
	topicVersion = 1

	if config["control-panel"]["server-up-at-start"] == "on" {
		gameSrvCanStart = true
	} else {
		gameSrvCanStart = false
	}

	gameSrvCanStartChan = make(chan bool)
	serverToken = "asdf"
	doneChan = make(chan error)
	gameSrvStatusChan = make(chan string)
	currentSrvVersionChan = make(chan string)
	availableSrvVersionChan = make(chan string)
	availableSrvVersionChan = make(chan string)
	currentSrvVersion = "1"
	availableSrvVersion = "1"
	statusPolls = make(map[int]chan string)
	responseBaseStr = "{\"debug\": %t, \"status\": \"%s\", \"current_version\": \"%s\", \"available_version\": \"%s\", \"topic_version\": %d}"
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
}

func main() {
	go runGameServerLoop()
	go runControlServer()
	go statusPollsReleaseWorker()

	for {
		select {
		case <-doneChan:
			if gameSrvStatus == "UP" {
				gameSrvStatusChan <- "DOWN (FAULT)"
			}
		}
	}
}
