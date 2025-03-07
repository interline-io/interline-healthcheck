package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/interline-io/interline-healthcheck/hc"
)

func getEnvKeys(v ...string) string {
	for _, key := range v {
		if os.Getenv(key) != "" {
			return key
		}
	}
	return ""
}

func main() {
	healthcheckId := ""
	workflowName := ""
	workflowStatus := ""
	workflowSetFail := false
	workflowSetSuccess := false
	flag.BoolVar(&workflowSetFail, "fail", false, "Set fail state")
	flag.BoolVar(&workflowSetSuccess, "success", false, "Set success state")
	flag.StringVar(
		&healthcheckId,
		"healthcheck-id",
		getEnvKeys("HEALTHCHECK_ID", "HEALTHCHECKSIO_CHECK_ID"),
		"Healthcheck ID, defaults to $HEALTHCHECK_ID or $HEALTHCHECKSIO_CHECK_ID",
	)
	flag.StringVar(
		&workflowName,
		"workflow-name",
		getEnvKeys("WORKFLOW_NAME", "HOSTNAME"),
		"Workflow name, defaults to $WORKFLOW_NAME or $HOSTNAME",
	)
	flag.StringVar(
		&workflowStatus,
		"workflow-status",
		getEnvKeys("WORKFLOW_STATUS"),
		"Workflow status, defaults to $WORKFLOW_STATUS",
	)
	flag.Parse()

	// Configure slack
	slackUrlSuccess := getEnvKeys("SLACK_URL_SUCCESS", "SLACK_URL_BOTS")
	slackUrlFailure := getEnvKeys("SLACK_URL_FAILURE", "SLACK_URL_GENERAL")

	// Configure workflowOk
	workflowOk := true
	if workflowStatus != "" && strings.ToLower(workflowStatus) != "succeeded" {
		workflowOk = false
	}
	if workflowSetFail {
		workflowOk = false
	}
	if workflowSetSuccess {
		workflowOk = true
	}

	// Debug
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
	}

	// Configure to run
	healthCheck := hc.NewHealthcheck(
		healthcheckId,
		workflowName,
		slackUrlSuccess,
		slackUrlFailure,
	)

	// Run subcommand
	var err error
	if cmd == "healthcheck_start" || cmd == "start" {
		err = healthCheck.Start()
	} else if cmd == "healthcheck_end" || cmd == "end" || cmd == "slack_notify" {
		err = healthCheck.End(workflowOk)
	} else {
		err = errors.New("invalid subcommand")
	}
	if err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
