package main

import (
	"fmt"
)

// Used for quick script testing

// go run main.go config.go
func main() {
	config, _ := LoadConfig()
	fmt.Println(config)
}
