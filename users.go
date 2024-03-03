package main

import (
	"fmt"
	"os"
	"strings"
)

var DATA_FOLDER = "data/users"

func GetDataFile(username string) string {
	return fmt.Sprintf("%s/%s", DATA_FOLDER, strings.ToLower(username))
}

func SaveUserSheet(userKey string, data string) error {
	// Ensure the "users" directory exists
	os.Mkdir(DATA_FOLDER, os.ModePerm)

	fileName := GetDataFile(userKey)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) FindUserSheet(userName string) string {
	fileName := fmt.Sprintf("data/users/%s", strings.ToLower(userName))
	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return strings.TrimSpace(string(data))
}
