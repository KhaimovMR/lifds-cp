package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getCharacterSkills(characterId int) map[string]int {
	charSkills := make(map[string]int)

	if dbExists == false {
		return charSkills
	}

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
	if dbExists == false {
		return
	}

	var rows *sql.Rows
	var err error
	accountQuery := `select c.ID, c.Name, c.LastName, c.CreateTimestamp, c.DeleteTimestamp, a.IsActive, a.SteamID
	from ` + "`character`" + ` c
	inner join account a on a.ID = c.AccountID
	order by c.Name, c.LastName;`
	rows, err = dbConn.Query(accountQuery)

	if checkError(err, "Error in querying database") {
		return
	}

	defer rows.Close()

	for rows.Next() {
		row := new(Character)
		err = rows.Scan(&row.ID, &row.Name, &row.LastName, &row.CreateTimestamp, &row.DeleteTimestamp, &row.IsActive, &row.SteamID)

		if err != nil {
			fmt.Println("Error:", err)
		}

		charKeysSorted = append(charKeysSorted, row.ID)
		fillCharactersMutex.Lock()
		characters[row.ID] = row
		fillCharactersMutex.Unlock()
		err = rows.Err()

		if err != nil {
			break
		}
	}

	log.Printf("%v characters has been loaded", len(characters))
}

func fillAccounts() {
	fillAccountsMutex.Lock()

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
	fillAccountsMutex.Unlock()
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

	if dbExists == false {
		return logItems
	}

	var logItem string
	query := "select Time, CharID, KillerID, IsKnockout from chars_deathlog where CharID = ? or KillerID = ? order by Time desc"
	rows, err = dbConn.Query(query, charId, charId)

	if checkError(err, "Error in querying database") {
		return logItems
	}

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

func getCharacterOnlineHistory(charId int) []string {
	var rows *sql.Rows
	var err error
	var logItems []string

	if dbExists == false {
		return logItems
	}

	var logItem string
	query := "select unix_timestamp(LoggedInAt), unix_timestamp(LoggedOutAt), IsLoggedOut from lifdscp_online_history where CharacterID = ? order by LoggedInAt desc"
	rows, err = dbConn.Query(query, charId)

	if checkError(err, "Error in querying database") {
		return logItems
	}

	defer rows.Close()
	var loggedInAt uint64
	var loggedOutAt uint64
	var isLoggedOut int

	for rows.Next() {
		err = rows.Scan(&loggedInAt, &loggedOutAt, &isLoggedOut)

		if err != nil {
			fmt.Println("Error:", err)
		}

		loggedInTime := time.Unix(int64(loggedInAt), 0)
		loggedInString := fmt.Sprintf(
			"%d-%02d-%02d %02d:%02d",
			loggedInTime.Year(),
			loggedInTime.Month(),
			loggedInTime.Day(),
			loggedInTime.Hour(),
			loggedInTime.Minute())

		if isLoggedOut == 1 {
			loggedOutTime := time.Unix(int64(loggedOutAt), 0)
			loggedOutString := fmt.Sprintf(
				"%d-%02d-%02d %02d:%02d",
				loggedOutTime.Year(),
				loggedOutTime.Month(),
				loggedOutTime.Day(),
				loggedOutTime.Hour(),
				loggedOutTime.Minute())
			logItem = fmt.Sprintf("<span class=\"online-history-date\">%v - %v</span>", loggedInString, loggedOutString)
		} else {
			logItem = fmt.Sprintf("<span class=\"online-history-date\">%v - till this moment</span>", loggedInString)
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

func getTotalCharacterOnlineTime(charId int) float32 {
	var totalHours float32
	var err error
	query := "select sum(unix_timestamp(LoggedOutAt) - unix_timestamp(LoggedInAt))/3600 from lifdscp_online_history where CharacterID = ? and IsLoggedOut = 1"
	rows, err := dbConn.Query(query, charId)

	if checkError(err, "Error in querying database") {
		return 0
	}

	defer rows.Close()
	rows.Next()
	rows.Scan(&totalHours)

	return totalHours
}

func createStatisticsTables() {
	if dbExists == false {
		log.Println("DB doesn't exists. Delaying statistics tables creation.")
		time.Sleep(time.Second * 10)
		go createStatisticsTables()
		return
	}

	query := "create table if not exists lifdscp_online_history (ID bigint unsigned not null auto_increment, CharacterID int unsigned not null, LoggedInAt timestamp not null default CURRENT_TIMESTAMP, LoggedOutAt timestamp not null default '0000-00-00 00:00:00', IsLoggedOut tinyint(1) not null default 0 comment '0 - still logged in, 1 - logged out', primary key (ID), key idx_CharacterID (CharacterID), key idx_LoggedInAt (LoggedInAt), key idx_LoggedOutAt (LoggedOutAt), key idx_IsLoggedOut (IsLoggedOut)) ENGINE=InnoDB auto_increment=1 default charset=utf8 collate=utf8_general_ci"
	_, err := dbConn.Exec(query)

	if err != nil {
		log.Println("Error on lifdscp_online_character table create:", err)
	}

	query = "create table if not exists lifdscp_online_character (CharacterID int unsigned not null, LoggedInAt timestamp not null default CURRENT_TIMESTAMP, primary key (CharacterID)) ENGINE=InnoDB auto_increment=1 default charset=utf8 collate=utf8_general_ci"
	_, err = dbConn.Exec(query)

	if err != nil {
		log.Println("Error on lifdscp_online_history table create:", err)
	}
}

func getOnlineCharactersList() []*OnlineCharacterInfo {
	var onlineCharsList []*OnlineCharacterInfo
	var char *Character
	var ok bool

	if dbExists == false {
		return onlineCharsList
	}

	onlineChars := getOnlineCharacters()
	log.Println("Online chars:", onlineChars)

	for charId, onlineTime := range onlineChars {
		char, ok = characters[charId]

		if ok == false {
			log.Printf("No character with id %v found", charId)
			continue
		}

		onlineChar := new(OnlineCharacterInfo)
		onlineChar.ID = strconv.Itoa(charId)
		onlineChar.FullName = fmt.Sprintf("%v %v", char.Name, char.LastName)
		onlineChar.OnlineTime = strconv.Itoa(onlineTime)
		onlineCharsList = append(onlineCharsList, onlineChar)
	}

	return onlineCharsList
}

func getOnlineCharacters() map[int]int {
	var rows *sql.Rows
	var err error
	var onlineChars map[int]int
	onlineChars = make(map[int]int)

	if dbExists {
		fillDbData()
	} else {
		log.Println("Passing empty onlineChars map, because of dbExists == false")
		return onlineChars
	}

	var onlineChar int
	var onlineTime int
	query := "select loc.CharacterID, (unix_timestamp(now()) - unix_timestamp(loc.LoggedInAt)) OnlineTime from lifdscp_online_character loc inner join `character` c on c.ID = loc.CharacterID order by c.Name, c.LastName"
	rows, err = dbConn.Query(query)

	if checkError(err, "Error in querying database") {
		return onlineChars
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&onlineChar, &onlineTime)
		log.Println(onlineChar)

		if err != nil {
			fmt.Println("Error:", err)
		}

		onlineChars[onlineChar] = onlineTime / 60
		err = rows.Err()

		if err != nil {
			break
		}
	}

	return onlineChars
}

func clearOnlineCharacters() {
	if dbExists == false {
		return
	}

	res, err := dbConn.Exec("delete from lifdscp_online_character")

	if err != nil {
		log.Println("Error on clearing online characters:", err)
		return
	}

	rowsAffected, err := res.RowsAffected()

	log.Printf("%v online characters cleared", rowsAffected)
}

func closeOfflineCharacterSessions() {
	if dbExists == false {
		return
	}

	query := "update lifdscp_online_history llh set llh.LoggedOutAt = NOW(), llh.IsLoggedOut = 1 where llh.IsLoggedOut = 0 and llh.CharacterID not in (select loc.CharacterID from lifdscp_online_character loc)"
	res, err := dbConn.Exec(query)

	if err != nil {
		log.Println("Error on closing offline character session:", err)
		return
	}

	rowsAffected, err := res.RowsAffected()
	log.Printf("%v offline character sessions closed", rowsAffected)
}

func openOnlineCharacterSessions() {
	if dbExists == false {
		return
	}

	query := "insert into lifdscp_online_history (CharacterID, LoggedInAt) select loc.CharacterID, loc.LoggedInAt from lifdscp_online_character loc where not exists (select llh.CharacterID from lifdscp_online_history llh where llh.IsLoggedOut = 0 and llh.CharacterID = loc.CharacterID)"
	res, err := dbConn.Exec(query)

	if err != nil {
		log.Println("Error on closing offline character session:", err)
		return
	}

	rowsAffected, err := res.RowsAffected()
	log.Printf("%v offline character sessions opened", rowsAffected)
}
