package telegram

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func generateSalt() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func makeState(telegramId int64) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d:%s", telegramId, salt), nil
}

func parseState(state string) (int64, string, error) {
	parts := strings.Split(state, ":")
	if len(parts) != 2 {
		return -1, "", fmt.Errorf("invalid state")
	}
	telegramId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return -1, "", err
	}
	salt := parts[1]
	return telegramId, salt, nil
}

func parseOAuth2RedirectURL(redirectURL string) (string, string, error) {
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return "", "", err
	}

	queryValues := parsedURL.Query()
	code := queryValues.Get("code")
	state := queryValues.Get("state")
	return code, state, nil
}
