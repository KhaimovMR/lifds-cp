package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func serverStartAction(w http.ResponseWriter, params map[string]string) {
	if exeCmd != nil && exeCmd.ProcessState == nil {
		log.Println("Preventing double start")
		return
	}

	gameSrvCanStartChan <- true
	message := "GETTING UP (MANUALLY)"
	gameSrvStatusChan <- message
	fmt.Fprint(w, "success")
}

func serverStopAction(w http.ResponseWriter, params map[string]string) {
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
}

func getActiveAccountsAction(w http.ResponseWriter, params map[string]string) {
	log.Println("Getting active accounts...")
	fillDbData()
	activeAccsSlice := getAccounts(true)
	activeAccs := make([]*BannedAccountResponse, len(activeAccsSlice))
	i := 0

	for _, activeAcc := range activeAccsSlice {
		j := 0
		bA := new(BannedAccountResponse)
		bA.SteamID = activeAcc.SteamID
		bA.Characters = make([]*CharacterLinkInfo, len(activeAcc.Characters))

		for _, charItem := range activeAcc.Characters {
			charLinkInfo := new(CharacterLinkInfo)
			charLinkInfo.ID = strconv.Itoa(charItem.ID)
			charLinkInfo.FullName = charItem.Name + " " + charItem.LastName
			bA.Characters[j] = charLinkInfo
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
}

func getBannedAccountsAction(w http.ResponseWriter, params map[string]string) {
	log.Println("Getting banned accounts...")
	fillDbData()
	bannedAccsSlice := getAccounts(false)
	bannedAccs := make([]*BannedAccountResponse, len(bannedAccsSlice))
	i := 0

	for _, bannedAcc := range bannedAccsSlice {
		j := 0
		bA := new(BannedAccountResponse)
		bA.SteamID = bannedAcc.SteamID
		bA.Characters = make([]*CharacterLinkInfo, len(bannedAcc.Characters))

		for _, charItem := range bannedAcc.Characters {
			charLinkInfo := new(CharacterLinkInfo)
			charLinkInfo.ID = strconv.Itoa(charItem.ID)
			charLinkInfo.FullName = charItem.Name + " " + charItem.LastName
			bA.Characters[j] = charLinkInfo
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
}

func getCharacterSkillsAction(w http.ResponseWriter, params map[string]string) {
	charId, err := strconv.Atoi(params["char_id"])

	if err != nil {
		fmt.Fprint(w, "{\"error\": \"char_id has wrong symbols\"}")
	}

	charSkills := getCharacterSkills(charId)
	stringified, err := json.Marshal(charSkills)

	if err != nil {
		fmt.Fprint(w, "{\"error\": \"Stringification failed\"}")
		log.Println(err)
	}

	fmt.Fprint(w, string(stringified))
}

func getCharacterDeathLogAction(w http.ResponseWriter, params map[string]string) {
	fillDbData()
	charId, err := strconv.Atoi(params["char_id"])

	if err != nil {
		fmt.Fprint(w, "{\"error\": \"char_id has wrong symbols\"}")
	}

	charLog := getCharacterDeathLog(charId)
	stringified, err := json.Marshal(charLog)

	if err != nil {
		fmt.Fprint(w, "{\"error\": \"Stringification failed\"}")
		log.Println(err)
	}

	fmt.Fprint(w, string(stringified))
}
