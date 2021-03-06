package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func Push(params []ParamToAgent, util string) {
	paramsBody, err := json.Marshal(params)
	if err != nil {
		log.Fatalln("[", util, "] , Error on formatting body:,", err)
	}

	postReq, err := http.NewRequest("POST", Config().Agent.PushURL, bytes.NewBuffer(paramsBody))
	postReq.Header.Set("Content-Type", "application/json; charset=UTF-8")
	postReq.Header.Set("Connection", "close")

	httpClient := &http.Client{}
	postResp, err := httpClient.Do(postReq)
	if err != nil {
		log.Println("[", util, "] Error on push:", err)
		return
	}
	defer postResp.Body.Close()
	log.Println("[", util, "] Pushing the HTTP Body...succeeded")
}
