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

type Healthcheck struct {
	HealthcheckId   string
	WorkflowName    string
	SlackUrlSuccess string
	SlackUrlFailure string
}

func NewHealthcheck(healthcheckId, workflowName, slackUrlSuccess, slackUrlFailure string) *Healthcheck {
	return &Healthcheck{
		HealthcheckId:   healthcheckId,
		WorkflowName:    workflowName,
		SlackUrlSuccess: slackUrlSuccess,
		SlackUrlFailure: slackUrlFailure,
	}
}

func NewHealthcheckFromEnv() *Healthcheck {
	return &Healthcheck{
		HealthcheckId:   os.Getenv("HEALTHCHECK_ID"),
		WorkflowName:    os.Getenv("WORKFLOW_NAME"),
		SlackUrlSuccess: os.Getenv("SLACK_URL_SUCCESS"),
		SlackUrlFailure: os.Getenv("SLACK_URL_FAILURE"),
	}
}

func (hc *Healthcheck) Start() error {
	if hc.HealthcheckId == "" {
		return nil
	}
	hcUrl := fmt.Sprintf("https://hc-ping.com/%s/start", hc.HealthcheckId)
	url, err := url.Parse(hcUrl)
	if err != nil {
		return err
	}
	if hc.WorkflowName != "" {
		q := url.Query()
		q.Add("rid", workflowUuid(hc.WorkflowName))
		url.RawQuery = q.Encode()
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

func (hc *Healthcheck) End(success bool) error {
	if hc.HealthcheckId == "" {
		return nil
	}
	exitCode := 0
	if !success {
		exitCode = 1
	}
	hcUrl := fmt.Sprintf("https://hc-ping.com/%s/%d", hc.HealthcheckId, exitCode)
	url, err := url.Parse(hcUrl)
	if err != nil {
		return err
	}
	if hc.WorkflowName != "" {
		q := url.Query()
		q.Add("rid", workflowUuid(hc.WorkflowName))
		url.RawQuery = q.Encode()
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

func (hc *Healthcheck) SlackNotify(success bool) error {
	slackUrl := hc.SlackUrlSuccess
	slackEmoji := ":globe_with_meridians:"
	if !success {
		slackUrl = hc.SlackUrlFailure
		slackEmoji = ":X:"
	}
	if slackUrl == "" {
		return nil
	}
	text := fmt.Sprintf("workflow %s success: %t %s", hc.WorkflowName, success, slackEmoji)
	resp, err := http.Post(slackUrl, "application/json", toJson(map[string]string{"text": text}))
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
