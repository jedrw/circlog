package cmd

import (
	"encoding/json"
	"fmt"
)

func outputJson(input any) error {
	j, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}
