package main

import "os"

func main() {

	SettingsPath, err := os.ReadFile("../../configs/settings.yaml")
	if err != nil {
		panic(err)
	}
	print(string(SettingsPath))
}
