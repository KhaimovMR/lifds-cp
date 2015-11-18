package main

import (
	"database/sql"
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

	deleteStubsQuery, err = ioutil.ReadFile("sql" + pathSeparator + "deleteStubs.sql")
	checkError(err, "Error in loading sql file")

	deleteLowQualityTreesQuery, err = ioutil.ReadFile("sql" + pathSeparator + "deleteLowQualityTrees.sql")
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
	case config["control-panel"]["online-statistics"] == "":
		config["control-panel"]["online-statistics"] = "off"
	}

	if config["lifds"]["lifds-exe-file-name"] == "" {
		exeFileName = "ddctd_cm_yo_server.exe"
	} else {
		exeFileName = config["lifds"]["lifds-exe-file-name"]
	}

	exePath = config["lifds"]["lifds-directory"] + pathSeparator + exeFileName
	worldCfgPath = config["lifds"]["lifds-directory"] + pathSeparator + "config" + pathSeparator + "world_" + config["lifds"]["world-id"] + ".xml"
	worldCfgContentsBuff, err := ioutil.ReadFile(worldCfgPath)
	worldCfgContents = string(worldCfgContentsBuff)
	localCfgPath = config["lifds"]["lifds-directory"] + pathSeparator + "config_local.cs"
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
	http.Handle("/", http.FileServer(http.Dir("."+pathSeparator+"html")))
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

func processClientAction(clientAction string, w http.ResponseWriter, params map[string]string) {
	switch clientAction {
	case "start":
		serverStartAction(w, params)
		break
	case "stop":
		serverStopAction(w, params)
		break
	case "delete-trees":
		//actionSqlExec(clientAction)
		fmt.Fprint(w, "success")
		break
	case "delete-stubs":
		//actionSqlExec(clientAction)
		fmt.Fprint(w, "success")
		break
	case "get-online-characters":
		getOnlineCharactersListAction(w)
		break
	case "get-character-death-log":
		getCharacterDeathLogAction(w, params)
		break
	case "get-character-online-history":
		getCharacterOnlineHistoryAction(w, params)
		break
	case "get-character-skills":
		getCharacterSkillsAction(w, params)
		break
	case "get-active-accounts":
		getActiveAccountsAction(w, params)
		break
	case "get-banned-accounts":
		getBannedAccountsAction(w, params)
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
	return fmt.Sprintf(
		responseBaseStr,
		debug,
		gameSrvStatus,
		currentSrvVersion,
		availableSrvVersion,
		topicVersion,
		config["control-panel"]["online-statistics"] == "on")
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
	go ensureDbExists()
}

func ensureDbExists() {
	for {
		_, err := dbConn.Exec(fmt.Sprintf("use lif_%v", config["lifds"]["world-id"]))

		if err == nil {
			dbExists = true
			break
		}

		time.Sleep(time.Second * 5)
	}
}

func fillDbData() {
	fillCharacters()
	fillAccounts()
}

func checkError(err error, message string) bool {
	if err != nil {
		log.Printf(message+": %v", err)
		return true
	}

	return false
}

func createCsFile(fileName string) {
	csContent, err := ioutil.ReadFile("cs" + pathSeparator + fileName)

	if checkError(err, fmt.Sprintf("Error on getting %s content", fileName)) {
		return
	}

	f, err := os.OpenFile(config["lifds"]["lifds-directory"]+pathSeparator+fileName, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println("Error in creating file: ", err)
		return
	}

	defer f.Close()
	_, err = f.Write(csContent)

	if err != nil {
		log.Println("Error in file writing:", err)
		return
	}

	fmt.Printf("File %s successfully written", fileName)
}

func includeCsFile(fileName string) {
	if checkIfIncludeNeeded(fileName) == false {
		return
	}

	writeCsInclude(fileName)
}

func checkIfIncludeNeeded(fileName string) bool {
	mainCsContent, err := ioutil.ReadFile(config["lifds"]["lifds-directory"] + pathSeparator + "main.cs")

	if err != nil {
		log.Println("Error on getting main.cs content:", err)
	}

	compiled, _ := regexp.Compile("exec[(]\"" + fileName + "\"[)];")
	matches := compiled.FindStringSubmatch(string(mainCsContent))

	if len(matches) > 0 {
		return false
	} else {
		return true
	}
}

func writeCsInclude(fileName string) {
	f, err := os.OpenFile(config["lifds"]["lifds-directory"]+pathSeparator+"main.cs", os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		log.Println("Error in creating file: ", err)
		return
	}

	defer f.Close()
	_, err = f.WriteString("\r\nexec(\"" + fileName + "\");")

	if err != nil {
		log.Println("Error in writing main.cs file:", err)
		return
	}

	fmt.Printf("File %s successfully included in main.cs", fileName)
}
