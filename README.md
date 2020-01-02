## What

Returns tab-separated build information, ready to be pasted in the CI spreadsheet.

## Root cause analysis

The program looks for errored machines in `machines.json` and in `openstack_nodes.log`. If it finds any, it directly prints `Provisioned VM in ERROR state`.

The rules are coded in `pkg/rca/rule.go`. When new rules are coded, and add them to the batch in `pkg/rca/rca.go`.

## How

```shell
go build ./cmd/cireport
./cireport -job e2e-openstack-serial -target 4.2 -from 345 -to 346
```

```
345     2019-12-28 14:37:19 +0000 UTC   2h11m34s        SUCCESS                                 cireport
346     2019-12-29 02:38:10 +0000 UTC   2h22m1s FAILURE                                 cireport        Provisioned VM in ERROR state
```

This command gets entries ready for pasting into the spreadsheet, given `$CLIPBOARD-COPY` your favourite clipboard "copy" command:

```shell
./cireport -job e2e-openstack-serial -target 4.2 -from 345 -to 346 | tee /dev/stderr | tac | "$CLIPBOARD-COPY"
```
