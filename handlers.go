package main

import (
	"fmt"
	auth "github.com/abbot/go-http-auth"
	"net/http"
	"strconv"
	"time"
)

func indexHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	fmt.Fprint(w, indexHtml)
}

func ServerActionsHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	r.ParseForm()
	clientToken, ok := r.PostForm["token"]

	if ok == false {
		fmt.Fprint(w, "{\"error\": \"token wasn't specified\"}")
		return
	}

	if clientToken[0] != serverToken {
		fmt.Fprint(w, "{\"error\": \"wrong token\"}")
		return
	}

	clientAction, ok := r.PostForm["action"]

	if ok == false {
		fmt.Fprint(w, "{\"error\": \"action wasn't specified\"}")
		return
	}

	processClientAction(clientAction[0], w)
}

func ServerStatusHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	r.ParseForm()

	clientToken, ok := r.PostForm["token"]

	if ok == false {
		fmt.Fprint(w, "{\"error\": \"token wasn't specified\"}")
		return
	}

	if clientToken[0] != serverToken {
		fmt.Fprint(w, "{\"error\": \"wrong token\"}")
		return
	}

	clientTopicVersion, ok := r.PostForm["topic_version"]
	clientTopicVersionInt, err := strconv.Atoi(clientTopicVersion[0])

	if err != nil {
		clientTopicVersionInt = 0
	}

	if ok && clientTopicVersionInt != topicVersion {
		response := getStatusResponse(true)
		fmt.Fprint(w, response)
		return
	}

	timeout := make(chan bool)
	responseChan := make(chan string)
	pollIndex := len(statusPolls)
	statusPolls[pollIndex] = responseChan

	go func() {
		time.Sleep(time.Second * 35)
		timeout <- true
	}()

	select {
	case responseBody := <-responseChan:
		fmt.Fprint(w, responseBody)
		return
	case <-timeout:
		delete(statusPolls, pollIndex)
		w.WriteHeader(503)
		return
	}
}
