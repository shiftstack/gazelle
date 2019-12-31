## What

Returns tab-separated build information, ready to be pasted in the CI spreadsheet.

## How

```
go build ./cmd/cireport
./cireport -job release-openshift-ocp-installer-e2e-openstack-serial-4.3 -from 1 -to 2
```

```
2       2019-10-01 23:13:56 +0000 UTC   28s     FAILURE
1       2019-10-01 19:13:26 +0000 UTC   23s     FAILURE
```
