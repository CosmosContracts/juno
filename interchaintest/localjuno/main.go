package main

import (
	"fmt"
	"strings"
)

// go run main.go config.go
func main() {
	config, _ := LoadConfig()
	fmt.Println(config)

	for _, values := range config.Chains[0].Genesis.Modify {
		loc := strings.Split(values.Key, ".")
		// fmt.Println(loc, values.Value)

		var result []interface{}
		for _, component := range loc {
			result = append(result, component)
		}

		fmt.Println(result)
	}
}
