package hc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
)

func SlackNotify(workflowName string, success bool) error {
	slackUrl := os.Getenv("SLACK_URL_BOTS")
	slackEmoji := ":globe_with_meridians:"
	if !success {
		slackUrl = os.Getenv("SLACK_URL_GENERAL")
		slackEmoji = ":X:"
	}
	if slackUrl == "" {
		return errors.New("set $SLACK_URL_BOTS and $SLACK_URL_GENERAL")
	}
	text := fmt.Sprintf("workflow %s success: %t %s", workflowName, success, slackEmoji)
	resp, err := http.Post(slackUrl, "application/json", toJson(map[string]string{"text": text}))
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

func HealthcheckStart(workflowName string, healthcheckId string) error {
	if healthcheckId == "" {
		return errors.New("set --healthcheck-id or $HEALTHCHECKSIO_CHECK_ID")
	}
	hcUrl := fmt.Sprintf("https://hc-ping.com/%s/start", healthcheckId)
	url, err := url.Parse(hcUrl)
	if err != nil {
		return err
	}
	if workflowName != "" {
		q := url.Query()
		q.Add("rid", workflowUuid(workflowName))
		url.RawQuery = q.Encode()
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

func HealthcheckEnd(workflowName string, healthcheckId string, success bool) error {
	if healthcheckId == "" {
		return errors.New("set --healthcheck-id or $HEALTHCHECKSIO_CHECK_ID")
	}
	exitCode := 0
	if !success {
		exitCode = 1
	}
	hcUrl := fmt.Sprintf("https://hc-ping.com/%s/%d", healthcheckId, exitCode)
	url, err := url.Parse(hcUrl)
	if err != nil {
		return err
	}
	if workflowName != "" {
		q := url.Query()
		q.Add("rid", workflowUuid(workflowName))
		url.RawQuery = q.Encode()
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

func toJson(v any) io.Reader {
	jj, _ := json.Marshal(v)
	return bytes.NewReader(jj)
}

func checkResponse(resp *http.Response) error {
	respData, _ := io.ReadAll(resp.Body)
	respString := string(respData)
	if resp.StatusCode > 299 {
		return errors.New("request failed: " + respString)
	}
	fmt.Println("response ok:", respString)
	return nil
}

func workflowUuid(workflowName string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(workflowName)).String()
}
