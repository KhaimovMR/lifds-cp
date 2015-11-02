package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	auth "github.com/abbot/go-http-auth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/vaughan0/go-ini"
	"io/ioutil"
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

func loadSqlQueries() map[string]string {
	var err error
	var queries map[string]string
	var deleteStubsQuery, deleteLowQualityTreesQuery []byte
	queries = make(map[string]string)

	deleteStubsQuery, err = ioutil.ReadFile("sql" + osPathSeparator + "deleteStubs.sql")
	checkError(err, "Error in loading sql file")

	deleteLowQualityTreesQuery, err = ioutil.ReadFile("sql" + osPathSeparator + "deleteLowQualityTrees.sql")
	checkError(err, "Error in loading sql file")

	queries["delete-stubs"] = string(deleteStubsQuery)
	queries["delete-trees"] = string(deleteLowQualityTreesQuery)
	return queries
}

func loadConfiguration() map[string]map[string]string {
	config := make(map[string]map[string]string)
	iniFile, err := ini.LoadFile("lifds-cp.ini")

	if err != nil {
		exitWithMessage("Can't read configuration from lifds-cp.ini file.")
	}

	config["lifds"] = iniFile.Section("lifds")
	config["control-panel"] = iniFile.Section("control-panel")

	config["lifds"]["db-host"] = ""

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

	if config["lifds"]["lifds-exe-file-name"] == "" {
		exeFileName = "ddctd_cm_yo_server.exe"
	} else {
		exeFileName = config["lifds"]["lifds-exe-file-name"]
	}

	exePath = config["lifds"]["lifds-directory"] + osPathSeparator + exeFileName
	worldCfgPath = config["lifds"]["lifds-directory"] + osPathSeparator + "config" + osPathSeparator + "world_" + config["lifds"]["world-id"] + ".xml"
	worldCfgContentsBuff, err := ioutil.ReadFile(worldCfgPath)
	worldCfgContents = string(worldCfgContentsBuff)
	localCfgPath = config["lifds"]["lifds-directory"] + osPathSeparator + "config_local.cs"
	localCfgFile, err := ini.LoadFile(localCfgPath)
	dirtyDbHost, ok := localCfgFile.Get("", "$cm_config::DB::Connect::server")

	if ok == false {
		dirtyDbHost = ""
	}

	dirtyDbUser, ok := localCfgFile.Get("", "$cm_config::DB::Connect::user")

	if ok == false {
		dirtyDbUser = ""
	}

	dirtyDbPassword, ok := localCfgFile.Get("", "$cm_config::DB::Connect::password")

	if ok == false {
		dirtyDbPassword = ""
	}

	dbHost := getCleanLocalCfgValue(dirtyDbHost)
	dbUser := getCleanLocalCfgValue(dirtyDbUser)
	dbPassword := getCleanLocalCfgValue(dirtyDbPassword)
	config["lifds"]["db-host"] = dbHost
	config["lifds"]["db-user"] = dbUser
	config["lifds"]["db-password"] = dbPassword

	if err != nil {
		log.Printf("Can't open config file \"%s\"", worldCfgPath)
	}

	return config
}

func runControlServer() {
	authenticator := auth.NewBasicAuthenticator("localhost", Secret)
	http.HandleFunc("/server", authenticator.Wrap(ServerActionsHandler))
	http.HandleFunc("/server/status", authenticator.Wrap(ServerStatusHandler))
	http.Handle("/index.html", authenticator.Wrap(indexHandler))
	http.Handle("/", http.FileServer(http.Dir("."+osPathSeparator+"html")))
	http.ListenAndServe(config["control-panel"]["address"]+":"+config["control-panel"]["port"], nil)
}

func actionSqlExec(action string) {
	queryParts := strings.Split(sqls[action], "//")

	for _, queryPart := range queryParts {
		result, err := dbConn.Exec(queryPart)
		checkError(err, "Error on action query: "+action)
		log.Printf("Action %v query result: %v", action, result)
	}
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
	case "delete-trees":
		//actionSqlExec(clientAction)
		fmt.Fprint(w, "success")
		break
	case "delete-stubs":
		//actionSqlExec(clientAction)
		fmt.Fprint(w, "success")
		break
	case "get-active-accounts":
		log.Println("Getting active accounts...")
		fillDbData()
		activeAccsSlice := getAccounts(true)
		activeAccs := make([]*BannedAccountResponse, len(activeAccsSlice))
		i := 0

		for _, activeAcc := range activeAccsSlice {
			j := 0
			bA := new(BannedAccountResponse)
			bA.SteamID = activeAcc.SteamID
			bA.Characters = make([]string, len(activeAcc.Characters))

			for _, charItem := range activeAcc.Characters {
				bA.Characters[j] = charItem.Name + " " + charItem.LastName
				j++
			}

			activeAccs[i] = bA
			i++
		}

		stringified, err := json.Marshal(activeAccs)

		if err != nil {
			log.Panic("Stringification failed")
		}

		fmt.Fprint(w, string(stringified))
		fmt.Println(string(stringified))
		break
	case "get-banned-accounts":
		log.Println("Getting banned accounts...")
		fillDbData()
		bannedAccsSlice := getAccounts(false)
		bannedAccs := make([]*BannedAccountResponse, len(bannedAccsSlice))
		i := 0

		for _, bannedAcc := range bannedAccsSlice {
			j := 0
			bA := new(BannedAccountResponse)
			bA.SteamID = bannedAcc.SteamID
			bA.Characters = make([]string, len(bannedAcc.Characters))

			for _, charItem := range bannedAcc.Characters {
				bA.Characters[j] = charItem.Name + " " + charItem.LastName
				j++
			}

			bannedAccs[i] = bA
			i++
		}

		stringified, err := json.Marshal(bannedAccs)

		if err != nil {
			log.Panic("Stringification failed")
		}

		fmt.Fprint(w, string(stringified))
		fmt.Println(string(stringified))
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

func getCleanLocalCfgValue(dirtyValue string) string {
	compiled, _ := regexp.Compile("\"([^\"]*)\"")
	matches := compiled.FindStringSubmatch(dirtyValue)

	if len(matches) > 1 {
		return matches[1]
	} else {
		return ""
	}
}

func Secret(user, realm string) string {
	a := auth.MD5Crypt([]byte(adminPassword), []byte("gjnjVexnjNfr"), []byte("$1$"))
	return string(a)
}

func initDbConnection() {
	var err error
	connectStr := fmt.Sprintf(
		"%v:%v@tcp(%v:3306)/lif_%v?charset=utf8&parseTime=true",
		config["lifds"]["db-user"],
		config["lifds"]["db-password"],
		config["lifds"]["db-host"],
		config["lifds"]["world-id"])
	fmt.Println(connectStr)
	dbConn, err = sql.Open("mysql", connectStr)
	checkError(err, "MySQL connection failed")
	dbConn.SetMaxIdleConns(3)
	dbConn.SetMaxOpenConns(9)
}

func fillDbData() {
	fillCharacters()
	fillAccounts()
}

func fillCharacters() {
	var rows *sql.Rows
	var err error
	accountQuery := `select c.Name, c.LastName, c.CreateTimestamp, c.DeleteTimestamp, a.IsActive, a.SteamID
	from ` + "`character`" + ` c
	inner join account a on a.ID = c.AccountID
	order by c.Name, c.LastName;`
	rows, err = dbConn.Query(accountQuery)
	checkError(err, "Error in querying database")
	defer rows.Close()

	for rows.Next() {
		row := new(Character)
		err = rows.Scan(&row.Name, &row.LastName, &row.CreateTimestamp, &row.DeleteTimestamp, &row.IsActive, &row.SteamID)

		if err != nil {
			fmt.Println("Error:", err)
		}

		charKeysSorted = append(charKeysSorted, row.Name)
		characters[row.Name] = row
		err = rows.Err()

		if err != nil {
			break
		}
	}

	log.Printf("%v characters has been loaded", len(characters))
}

func fillAccounts() {
	for _, key := range charKeysSorted {
		charItem := characters[key]

		if accounts[charItem.SteamID] == nil {
			accounts[charItem.SteamID] = new(Account)
			accounts[charItem.SteamID].SteamID = charItem.SteamID
			accounts[charItem.SteamID].Characters = make(map[string]*Character)
			accKeysSorted = append(accKeysSorted, charItem.SteamID)
		}

		accounts[charItem.SteamID].IsActive = charItem.IsActive
		accounts[charItem.SteamID].Characters[charItem.Name] = charItem
	}

	log.Printf("%v accounts has been loaded", len(accounts))
}

func getAccounts(isActive bool) []*Account {
	var inactiveAccs []*Account

	for _, steamId := range accKeysSorted {
		acc := accounts[steamId]

		if acc.IsActive == isActive {
			inactiveAccs = append(inactiveAccs, acc)
		}
	}

	return inactiveAccs
}

func checkError(err error, message string) {
	if err != nil {
		log.Println(message)
		log.Printf("Error: %v", err)
	}
}
