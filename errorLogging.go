package main

import (
	"os"
)

func WriteErrorToFile(errorMessage string) {
	f, err := os.OpenFile("./logs/error_message.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if _, err := f.Write([]byte(errorMessage + "\n")); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}
