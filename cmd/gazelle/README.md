# Gazelle

Gazelle looks for failure Fingerprints in arbitrary CI runs.

A fingerprint is an ElasticSearch query that uniquely identifies a known
failure. We use Fingerprints to link failures to Bugzilla IDs, or other
interesting information.

Once known failures are filtered out from a failed run, we know when we have a
new failure to investigate.

## Endpoints

### Run metadata

```
GET /runs/job_URL
```

For example, to analyse the run n. 1311 of the job "release-openshift-ocp-installer-e2e-openstack-4.4", we would run this query:

```
/runs/prow.svc.ci.openshift.org/view/gcs/origin-ci-test/logs/release-openshift-ocp-installer-e2e-openstack-4.4/1311
```

The response payload contains information about the run:

```
{
  "jobURL": string (should match what passed in the request),
  "started": string (RFC3339 timestamp),
  "duration": string (duration in Go format),
  "result": string (most likely "SUCCESS" or "FAILURE"),
  "fingerprints": array of strings (the IDs of the matching fingerprints)
}
```

### Fingerprints

```
GET /fingerprints
```

Returns an array containing all known Fingerprints.

A Fingerprint is an ElasticSearch query with some metadata:
```
{
  "id": string (unique ID of the Fingerprint),
  "title": string,
  "description": string,
  "links": array of strings (urls of bugzillas, or github issues...),
  "query": string (ElasticSearch query)
}
```

### Query

A Query is an ElasticSearch query.

Every Run is a document. Every relevant artifact is a field.

Artifacts are parsed line-by-line; searches are performed against lines:

```json
{
  "settings": {
    "analysis": {
      "analyzer": {
        "line_by_line": {
          "type":      "pattern",
          "pattern":   "$",
          "lowercase": false
        }
      }
    }
  }
}
```

Example query:

```json
{
	"should": [{
		"regexp": {
			"build-log.txt": ".*to become ready: unexpected state 'ERROR', wanted target 'ACTIVE'.*"
		}
	},{
		"regexp": {
			"build-log.txt": ".*to become ready: timeout while waiting for state to become 'ACTIVE'.*"
		}
	},{
		"regexp": {
			"machines.json": ".*\"machine.openshift.io/instance-state\": \"ERROR\".*"
		}
	},{
		"regexp": {
			"openstack_nodes.log": ".*ERROR.*"
		}
	}]
}
```

## How it works

Upon receiving a request, Gazelle runs all known Fingerprint queries against a
Run. If the Run is not available in ElasticSearch, Gazelle uploads all the
required data to ElasticSearch.

All matching Fingerprint IDs are returned in the API response.

Fingerprints are the only data Gazelle owns. They are currently hardcoded. In
perspective, the aim of the project is to store a growing collection of failure
fingerprints. Fingerprints identify a failure and likely contain a link to Bugzilla.
