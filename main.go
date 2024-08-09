package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
)

func main() {
	healthcheckId := ""
	workflowName := ""
	workflowStatus := ""
	workflowOk := true
	workflowSetFail := false
	workflowSetSuccess := false
	flag.StringVar(&healthcheckId, "healthcheck-id", os.Getenv("HEALTHCHECKSIO_CHECK_ID"), "Healthcheck ID, defaults to $HEALTHCHECKSIO_CHECK_ID")
	flag.StringVar(&workflowName, "workflow-name", os.Getenv("WORKFLOW_NAME"), "Workflow name, defaults to $WORKFLOW_NAME")
	flag.StringVar(&workflowStatus, "workflow-status", os.Getenv("WORKFLOW_STATUS"), "Workflow status, defaults to $WORKFLOW_STATUS")
	flag.BoolVar(&workflowSetFail, "fail", false, "Set fail state")
	flag.BoolVar(&workflowSetSuccess, "success", false, "Set success state")
	flag.Parse()
	if workflowStatus != "" && strings.ToLower(workflowStatus) != "succeeded" {
		workflowOk = false
	}
	if workflowSetFail {
		workflowOk = false
	}
	if workflowSetSuccess {
		workflowOk = true
	}
	cmd := flag.Arg(0)
	fmt.Println(
		"cmd:", cmd,
		"workflowName:", workflowName,
		"workflowStatus:", workflowStatus,
		"workflowOk:", workflowOk,
		"healthcheckId:", healthcheckId,
	)
	if workflowName == "" {
		fail("set --workflow-name or $WORKFLOW_NAME")
		return
	}
	// Run subcommand
	var err error
	if cmd == "slack_notify" {
		err = slackNotify(workflowName, workflowOk)
	} else if cmd == "healthcheck_start" {
		if healthcheckId == "" {
			fail("set --healthcheck-id or $HEALTHCHECKSIO_CHECK_ID")
		}
		err = healthcheckStart(workflowName, healthcheckId)
	} else if cmd == "healthcheck_end" {
		if healthcheckId == "" {
			fail("set --healthcheck-id or $HEALTHCHECKSIO_CHECK_ID")
		}
		err = healthcheckEnd(workflowName, healthcheckId, workflowOk)
	} else {
		err = errors.New("invalid subcommand")
	}
	if err != nil {
		fail(err.Error())
	}
}

func slackNotify(workflowName string, success bool) error {
	slackUrl := os.Getenv("SLACK_URL_BOTS")
	slackEmoji := ":globe_with_meridians:"
	if !success {
		slackUrl = os.Getenv("SLACK_URL_GENERAL")
		slackEmoji = ":X:"
	}
	text := fmt.Sprintf("workflow %s success: %t %s", workflowName, success, slackEmoji)
	resp, err := http.Post(slackUrl, "application/json", toJson(map[string]string{"text": text}))
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

func healthcheckStart(workflowName string, healthcheckId string) error {
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

func healthcheckEnd(workflowName string, healthcheckId string, success bool) error {
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

func fail(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
