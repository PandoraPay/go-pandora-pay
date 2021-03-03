package helpers

import "encoding/json"

func GetJSON(obj interface{}, ignoreFields ...string) (out []byte, err error) {

	var toJson []byte
	if toJson, err = json.Marshal(obj); err != nil {
		return
	}

	if len(ignoreFields) == 0 {
		out = toJson
		return
	}

	toMap := map[string]interface{}{}
	json.Unmarshal(toJson, &toMap)

	for _, field := range ignoreFields {
		delete(toMap, field)
	}

	if toJson, err = json.Marshal(toMap); err != nil {
		return
	}

	out = toJson
	return
}
