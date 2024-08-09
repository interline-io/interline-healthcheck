#!/bin/python3
import sys
import os
import uuid
import requests

def healthcheck_end(healthcheck_id, workflow_name, success=True):
    check_params = {}
    if workflow_name:
        check_params['rid'] = uuid.uuid3(uuid.NAMESPACE_DNS, workflow_name)
    if success:
        res = requests.get(f"https://hc-ping.com/{healthcheck_id}/0", check_params)
        print(res.status_code, res.content)
        res.raise_for_status()         
    else:
        res = requests.get(f"https://hc-ping.com/{healthcheck_id}/fail", check_params)
        print(res.status_code, res.content)
        res.raise_for_status() 

def healthcheck_start(healthcheck_id, workflow_name):
    check_params = {}
    if workflow_name:
        check_params['rid'] = uuid.uuid3(uuid.NAMESPACE_DNS, workflow_name)
    res = requests.get(f"https://hc-ping.com/{healthcheck_id}/start", check_params)
    print(res.status_code, res.content)
    res.raise_for_status() 

def slack_notify(workflow_name, success=True):
    slack_url = os.environ.get('SLACK_URL_BOTS') or ''
    slack_emoji=":globe_with_meridians:"
    if not success:
        slack_url = os.environ.get('SLACK_URL_GENERAL') or ''
        slack_emoji=":X:"
    slack_params = {
        "text": f"workflow {workflow_name} success: {success} {slack_emoji}"
    }
    res = requests.post(
        slack_url,
        json=slack_params,
        headers={"Content-Type": "application/json"}
    )
    print(res.status_code, res.content)
    res.raise_for_status() 

def fail(msg):
    print(msg)
    sys.exit(1)

def main_cli():
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--healthcheck-id', help='Healthcheck ID, defaults to $HEALTHCHECKSIO_CHECK_ID')
    parser.add_argument('--workflow-name', help='Workflow name, defaults to $WORKFLOW_NAME')
    parser.add_argument('--workflow-status', help='Workflow status, defaults to $WORKFLOW_STATUS')
    parser.add_argument('--fail', action='store_true', help='Set fail state')
    parser.add_argument('--success', action='store_true', help="Set success state")
    parser.add_argument('command')
    args = parser.parse_args()
    
    healthcheck_id = args.healthcheck_id or os.environ.get('HEALTHCHECKSIO_CHECK_ID')
    workflow_name = args.workflow_name or os.environ.get('WORKFLOW_NAME') or ''
    workflow_status = args.workflow_status or (os.environ.get('WORKFLOW_STATUS') or '').lower()
    workflow_ok = True
    if workflow_status and workflow_status != "succeeded":
        workflow_ok = False
    if args.fail:
        workflow_ok = False
    if args.success:
        workflow_ok = True
    
    print("cmd:", args.command, "workflow_name:", workflow_name, "workflow_status:", workflow_status, "workflow_ok:", workflow_ok, "healthcheck_id:", healthcheck_id)
    if not workflow_name:
        fail("set WORKFLOW_NAME")

    if args.command == "slack_notify":
        slack_notify(workflow_name, success=workflow_ok)
        return
    
    if args.command == "healthcheck_start":
        if not healthcheck_id:
            fail("set HEALTHCHECKSIO_CHECK_ID")
        healthcheck_start(healthcheck_id, workflow_name)
        return

    if args.command == "healthcheck_end":
        if not healthcheck_id:
            fail("set HEALTHCHECKSIO_CHECK_ID")
        healthcheck_end(healthcheck_id, workflow_name, success=workflow_ok)
        return

    fail("invalid subcommand:", args.command)


if __name__ == "__main__":
    main_cli()

