package telegram

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
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
	var telegramId int64
	var salt string
	n, err := fmt.Sscanf(state, "%d:%s", &telegramId, &salt)
	if err != nil {
		return -1, "", err
	}
	if n != 2 {
		return -1, "", fmt.Errorf("invalid state format")
	}
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
