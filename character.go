package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"time"
)

func getCharacterSkills(characterId int) map[string]int {
	charSkills := make(map[string]int)

	if characterId == 0 {
		return charSkills
	}

	var rows *sql.Rows
	var err error
	query := "select SkillTypeID as skill, round(SkillAmount/10000000) as amount from skills where CharacterID = ?"
	rows, err = dbConn.Query(query, characterId)
	checkError(err, "Error in querying database")
	defer rows.Close()
	var skillTypeID int
	var skillAmount int

	for rows.Next() {
		err = rows.Scan(&skillTypeID, &skillAmount)

		if err != nil {
			fmt.Println("Error:", err)
		}

		charSkills[strconv.Itoa(skillTypeID)] = skillAmount
		err = rows.Err()

		if err != nil {
			break
		}
	}

	return charSkills
}

func fillCharacters() {
	var rows *sql.Rows
	var err error
	accountQuery := `select c.ID, c.Name, c.LastName, c.CreateTimestamp, c.DeleteTimestamp, a.IsActive, a.SteamID
	from ` + "`character`" + ` c
	inner join account a on a.ID = c.AccountID
	order by c.Name, c.LastName;`
	rows, err = dbConn.Query(accountQuery)
	checkError(err, "Error in querying database")
	defer rows.Close()

	for rows.Next() {
		row := new(Character)
		err = rows.Scan(&row.ID, &row.Name, &row.LastName, &row.CreateTimestamp, &row.DeleteTimestamp, &row.IsActive, &row.SteamID)

		if err != nil {
			fmt.Println("Error:", err)
		}

		charKeysSorted = append(charKeysSorted, row.ID)
		characters[row.ID] = row
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
	var accs []*Account

	for _, steamId := range accKeysSorted {
		acc := accounts[steamId]

		if acc.IsActive == isActive {
			accs = append(accs, acc)
		}
	}

	return accs
}

func getCharacterDeathLog(charId int) []string {
	var rows *sql.Rows
	var err error
	var logItems []string
	var logItem string
	query := "select Time, CharID, KillerID, IsKnockout from chars_deathlog where CharID = ? or KillerID = ? order by Time desc"
	rows, err = dbConn.Query(query, charId, charId)
	checkError(err, "Error in querying database")
	defer rows.Close()
	var itemTime uint64
	var victimId int
	var killerId int
	var isKnockout int
	var victimName string
	var killerName string
	var victim *Character
	var killer *Character

	for rows.Next() {
		err = rows.Scan(&itemTime, &victimId, &killerId, &isKnockout)

		if err != nil {
			fmt.Println("Error:", err)
		}

		victim = characters[victimId]
		killer = characters[killerId]

		if victim == nil {
			victimName = "&lt;user-id: " + strconv.Itoa(killerId) + "&gt;"
		} else {
			victimName = fmt.Sprintf("%s %s", characters[victimId].Name, characters[victimId].LastName)
		}

		if killer == nil {
			killerName = "&lt;user-id: " + strconv.Itoa(killerId) + "&gt;"
		} else {
			killerName = fmt.Sprintf("%s %s", characters[killerId].Name, characters[killerId].LastName)
		}

		timeVal := time.Unix(int64(itemTime), 0)
		timeString := fmt.Sprintf("%d-%02d-%02d %02d:%02d", timeVal.Year(), timeVal.Month(), timeVal.Day(), timeVal.Hour(), timeVal.Minute())

		if isKnockout == 1 {
			logItem = fmt.Sprintf("<span class=\"deathlog-date\">%v</span><strong>%v</strong> <span class=\"knocked-out-by\">knocked out by</span> <strong>%v</strong>", timeString, victimName, killerName)
		} else {
			logItem = fmt.Sprintf("<span class=\"deathlog-date\">%v</span><strong>%v</strong> <span class=\"killed-by\">killed by</span> <strong>%v</strong>", timeString, victimName, killerName)
		}

		logItems = append(logItems, logItem)
		err = rows.Err()

		if err != nil {
			break
		}
	}

	log.Println(logItems)
	return logItems
}
