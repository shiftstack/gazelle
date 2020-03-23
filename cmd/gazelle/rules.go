package main

var rules = []Rule{
	Rule{
		ID:    "bfe50482-5a77-46a5-a312-0a9479f09cb9",
		Title: "Provisioned VM in ERROR state",
		Query: `{
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
			}`,
	},
}
