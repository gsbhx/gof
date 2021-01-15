package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpClient() {
	response, err := http.NewRequest("GET", "http://localhost:8081", nil)
	if err != nil {
		fmt.Println("new request err:", err.Error())
		return
	}
	client := &http.Client{}
	client.Timeout = 5 * time.Second
	req, err := client.Do(response)
	if err != nil {
		fmt.Println("do request err:", err.Error())
		return
	}
	a, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("readAll err:", err.Error())
		return
	}
	fmt.Println(string(a))
	req.Body.Close()
}
