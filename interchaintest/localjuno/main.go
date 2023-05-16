package main

import (
	"fmt"
	"strings"
)

// Used for quick script testing

// go run main.go config.go
func main() {
	config, _ := LoadConfig()
	fmt.Println(config)

	for _, values := range config.Chains[0].Genesis.Modify {
		loc := strings.Split(values.Key, ".")

		var result []interface{}
		for _, component := range loc {
			// Need to also account for number interfaces here.
			result = append(result, component)
		}

		fmt.Println(result)
	}
}
