package telegram

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

func generateSalt() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func makeState(telegramId string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", telegramId, salt), nil
}

func parseState(state string) (telegramId string, salt string, err error) {
	parts := strings.Split(state, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid state")
	}
	return parts[0], parts[1], nil
}

func parseOAuth2RedirectURL(redirectURL string) (code string, state string, err error) {
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return "", "", err
	}

	queryValues := parsedURL.Query()
	code = queryValues.Get("code")
	state = queryValues.Get("state")
	return code, state, nil
}
