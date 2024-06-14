// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/gorilla/websocket"
)

func GetWebsocketConn(path string, profile *config.Profile, query *string) (*websocket.Conn, *http.Response, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	var serverUrl string
	var apiKey string

	if envApiUrl, ok := os.LookupEnv("DAYTONA_SERVER_API_URL"); ok {
		serverUrl = envApiUrl
		apiKey = os.Getenv("DAYTONA_SERVER_API_KEY")
	} else {
		var activeProfile config.Profile
		if profile == nil {
			var err error
			activeProfile, err = c.GetActiveProfile()
			if err != nil {
				return nil, nil, err
			}
		} else {
			activeProfile = *profile
		}

		serverUrl = activeProfile.Api.Url
		apiKey = activeProfile.Api.Key
	}

	url, err := url.JoinPath(serverUrl, path)
	if err != nil {
		return nil, nil, err
	}

	wsUrl, err := GetWebSocketUrl(url)
	if err != nil {
		return nil, nil, err
	}

	if query != nil {
		wsUrl = fmt.Sprintf("%s?%s", wsUrl, *query)
	}

	dialer := websocket.DefaultDialer

	// dialer.ReadBufferSize = 10000
	// dialer.WriteBufferSize = 10000

	return dialer.Dial(wsUrl, http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", apiKey)},
	})
}

func GetWebSocketUrl(apiUrl string) (string, error) {
	hostRegex := regexp.MustCompile(`(https*)://(.*)`)

	matches := hostRegex.FindStringSubmatch(apiUrl)

	if len(matches) != 3 {
		return "", errors.New("invalid API URL")
	}

	switch matches[1] {
	case "http":
		return fmt.Sprintf("ws://%s", matches[2]), nil
	case "https":
		return fmt.Sprintf("wss://%s", matches[2]), nil
	}

	return "", errors.New("invalid API URL protocol")
}
