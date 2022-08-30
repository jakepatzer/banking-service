package functions

import (
	"encoding/json"
	"fmt"
)

type RequestError struct {
	AccountID   string
	RequestBody string
	Err         string
}

func (e RequestError) Error() string {
	body, _ := json.Marshal(e)
	return fmt.Sprintf("Error processing request: %s", body)
}

func MarshalOutput(output interface{}) string {
	json, err := json.Marshal(output)
	if err != nil {
		return fmt.Sprintf("%v", output)
	} else {
		return string(json)
	}
}
