package rules

import (
	"encoding/json"
	"os"

	"github.com/syncfuture/go/serr"
)

func ReadRulesFromFile(filename string) (map[string]interface{}, error) {
	jsonBytes, err := os.ReadFile("rules_sample.json")
	if err != nil {
		return nil, serr.WithStack(err)
	}

	rules, err := ReadRules(jsonBytes)
	if err != nil {
		return nil, err
	}

	return rules, nil
}

func ReadRules(jsonBytes []byte) (map[string]interface{}, error) {
	var r map[string]interface{}
	var err error

	err = json.Unmarshal(jsonBytes, &r)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return r, nil
}
