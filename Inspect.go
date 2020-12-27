package Inpsect

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

func vars(objs map[string]interface{}) {
	inspectVarsPath, exists := os.LookupEnv("INSPECT_VARS")
	if !exists {
		return
	}

	data, jsonErr := json.Marshal(objs)
	if jsonErr != nil {
		return
	}

	if _, err := os.Stat(inspectVarsPath); os.IsNotExist(err) {
		os.Mkdir(inspectVarsPath, 0755)
	}

	file, fileErr := os.OpenFile(inspectVarsPath, os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.Encode(data)
}

func dump(obj interface{}) string {
	return spew.Sdump(obj)
}

func printDump(obj interface{}) {
	fmt.Println(dump(obj))
}
