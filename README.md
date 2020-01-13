## What

Returns tab-separated build information, ready to be pasted in the CI spreadsheet.

## Root cause analysis

The program looks for errored machines in `machines.json` and in `openstack_nodes.log`. If it finds any, it directly prints `Provisioned VM in ERROR state`.

The rules are coded in `pkg/rca/rule.go`. When new rules are coded, and add them to the batch in `pkg/rca/rca.go`.

## How

```shell
go build ./cmd/cireport
./cireport -job release-openshift-origin-installer-e2e-openstack-serial-4.2 -id 346-345
```

```HTML
<meta http-equiv="content-type" content="text/html; charset=utf-8"><meta name="generator" content="cireport"/><table xmlns="http://www.w3.org/1999/xhtml"><tbody><tr><td><a href="https://prow.svc.ci.openshift.org/view/gcs/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/345">345</a></td><td>2019-12-28 14:37:19 +0000 UTC</td><td>2h11m34s</td><td>SUCCESS</td><td></td><td><a href="https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/345/build-log.txt">https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/345/build-log.txt</a></td><td><a href="https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/345/artifacts/e2e-openstack-serial/machines.json">https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/345/artifacts/e2e-openstack-serial/machines.json</a></td><td><a href="https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/345/artifacts/e2e-openstack-serial/openstack_nodes.log">https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/345/artifacts/e2e-openstack-serial/openstack_nodes.log</a></td><td>cireport</td><td></td></tr></tbody></table>
<meta http-equiv="content-type" content="text/html; charset=utf-8"><meta name="generator" content="cireport"/><table xmlns="http://www.w3.org/1999/xhtml"><tbody><tr><td><a href="https://prow.svc.ci.openshift.org/view/gcs/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/346">346</a></td><td>2019-12-29 02:38:10 +0000 UTC</td><td>2h22m1s</td><td>FAILURE</td><td></td><td><a href="https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/346/build-log.txt">https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/346/build-log.txt</a></td><td><a href="https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/346/artifacts/e2e-openstack-serial/machines.json">https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/346/artifacts/e2e-openstack-serial/machines.json</a></td><td><a href="https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/346/artifacts/e2e-openstack-serial/openstack_nodes.log">https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-serial-4.2/346/artifacts/e2e-openstack-serial/openstack_nodes.log</a></td><td>cireport</td><td>Provisioned VM in ERROR state</td></tr></tbody></table>
```

This command gets entries ready for pasting into the spreadsheet, given `$CLIPBOARD-COPY` your favourite clipboard "copy" command:

```shell
./cireport -job release-openshift-ocp-installer-e2e-openstack-serial-4.2 -id 346-345 | tee /dev/stderr | "$CLIPBOARD-COPY"
```
