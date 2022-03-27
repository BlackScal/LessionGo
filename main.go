package main

import (
	"LessionGo/userService"
	"encoding/json"
	"fmt"
)

func exampleGetUserInfo(name string) ([]byte, error) {
	resp := make(map[string]interface{})
	//param check

	//get user info
	userInfo, err := userService.GetUserInfo(name)
	if err != nil {
		fmt.Printf("Get user info Failed:\n%+v", err)
		resp["code"] = 404
		resp["msg"] = "User Unexisted."
	} else {
		resp["code"] = 200
		resp["msg"] = "success"
		resp["data"] = userInfo
	}
	return json.Marshal(&resp)
}

func main() {
	// Guess a request comes
	userName := "UserB" 
	userInfo, err := exampleGetUserInfo(userName)
	if err != nil {
		panic(err)
	}
	
	//response
	fmt.Printf("%s\n", userInfo)

	return
}