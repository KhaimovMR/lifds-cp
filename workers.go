package main

import (
	"log"
	"time"
)

func runGameServerLoop() {
	for {
		if gameSrvCanStart {
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
