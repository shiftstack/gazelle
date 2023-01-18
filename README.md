## What

Polls OpenShift's Prow to get information about new jobs, and post notifications on Slack together with some additional information.

## Root cause analysis

At present, the program looks for errored machines. If it finds any, it prints `Provisioned VM in ERROR state`.

The rules are coded in `pkg/rca/rca.go`.

### Build gazelle

```shell
go build ./cmd/cireport
```

### Run gazelle

```shell
export SLACK_HOOK=<your Slack hook>
./cireport
```
