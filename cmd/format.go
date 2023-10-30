package cmd

import (
	"encoding/json"
	"fmt"
)

func outputJson(input any) {
	j, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(j))
}
