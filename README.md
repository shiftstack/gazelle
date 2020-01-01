## What

Returns tab-separated build information, ready to be pasted in the CI spreadsheet.

## Root cause analysis

The program looks for errored machines in `machines.json` and in `openstack_nodes.log`. If it finds any, it directly prints `Provisioned VM in ERROR state`.

The rules are coded in `pkg/rca/rule.go`. When new rules are coded, and add them to the batch in `pkg/rca/rca.go`.

## How

```
go build ./cmd/cireport
./cireport -job e2e-openstack-serial -target 4.2 -from 333 -to 334 | tee /dev/stderr | tac | wl-copy
```

```
334	2019-12-23 13:28:58 +0000 UTC	10m21s	FAILURE					cireport	Provisioned VM in ERROR state
333	2019-12-23 03:13:50 +0000 UTC	12m14s	FAILURE					cireport	Provisioned VM in ERROR state
```
