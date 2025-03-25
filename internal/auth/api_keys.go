package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetApiKey(headers http.Header) (string, error) {
	for _, token := range headers.Values("Authorization") {
		if strings.Contains(token, "ApiKey ") {
			return strings.TrimPrefix(token, "ApiKey "), nil
		}
	}
	return "", errors.New("authorization header doesn't exist")
}
