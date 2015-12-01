package main

import (
	"log"
	"time"
)

func runGameServerLoop() {
	for {
		if gameSrvCanStart {
			config = loadConfiguration()

			if config["control-panel"]["online-statistics"] == "on" {
				createStatisticsTables()
				createCsFile("lifdscp_stats.cs")
				includeCsFile("lifdscp_stats.cs")
				clearOnlineCharacters()
			} else {
				excludeCs("lifdscp_stats.cs")
			}

			runGameServer()
		}

		select {
		case gameSrvCanStart = <-gameSrvCanStartChan:
		case <-time.After(time.Second * 5):
		}
	}
}

func statusPollsReleaseWorker() {
	for {
		select {
		case gameSrvStatus = <-gameSrvStatusChan:
		case currentSrvVersion = <-currentSrvVersionChan:
		case availableSrvVersion = <-availableSrvVersionChan:
		}

		topicVersion += 1

		if topicVersion > 1000 {
			topicVersion = 1
		}

		log.Println(gameSrvStatus)

		for i, responseChan := range statusPolls {
			delete(statusPolls, i)
			responseChan <- getStatusResponse(false)
		}
	}
}

func onlineStatisticsWorker() {
	for {
		if config["control-panel"]["online-statistics"] == "off" {
			break
		}

		closeOfflineCharacterSessions()
		openOnlineCharacterSessions()
		time.Sleep(time.Second * 10)
	}
}
