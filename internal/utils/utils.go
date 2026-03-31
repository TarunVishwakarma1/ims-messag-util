package utils

import (
	"encoding/json"
	"log/slog"
	"strconv"
)

func ConvertBody[T any](body []byte) (T, error) {
	var data T

	if err := json.Unmarshal(body, &data); err != nil {
		slog.Error("Error in parsing passed body", "error", err)
		return data, err
	}

	return data, nil
}

func ConvertStringToInteger(str string) (int, error) {
	i, e := strconv.Atoi(str)
	if e != nil {
		slog.Error("Error in converting to Integer", "str value", str, "Error", e)
		return 0, e
	}
	return i, nil

}
