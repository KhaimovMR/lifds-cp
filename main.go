package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

type Account struct {
	SteamID    uint64
	IsActive   bool
	Characters map[string]*Character
}

type Character struct {
	ID              int
	Name            string
	LastName        string
	CreateTimestamp time.Time
	DeleteTimestamp sql.NullInt64
	IsActive        bool
	SteamID         uint64
}

type CharacterLinkInfo struct {
	ID       string
	FullName string
}

type BannedAccountResponse struct {
	SteamID    uint64
	Characters []*CharacterLinkInfo
}

var (
	adminPassword           string
	lastStartTime           time.Time
	lastStopTime            time.Time
	serverToken             string
	exePath                 string
	worldCfgPath            string
	worldCfgContents        string
	localCfgPath            string
	localCfgContents        string
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
	dbConn                  *sql.DB
	accounts                map[uint64]*Account
	characters              map[string]*Character
	charKeysSorted          []string
	accKeysSorted           []uint64
	sqls                    map[string]string
)

func init() {
	osPathSeparator = string(os.PathSeparator)
	sqls = loadSqlQueries()
	config = loadConfiguration()

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
	accounts = make(map[uint64]*Account)
	characters = make(map[string]*Character)
	currentSrvVersion = "1"
	availableSrvVersion = "1"
	statusPolls = make(map[int]chan string)
	responseBaseStr = "{\"debug\": %t, \"status\": \"%s\", \"current_version\": \"%s\", \"available_version\": \"%s\", \"topic_version\": %d}"
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	initDbConnection()
	fillDbData()
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
