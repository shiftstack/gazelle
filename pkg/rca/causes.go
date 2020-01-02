package rca

type Cause string

const (
	CauseErroredVM      Cause = "Provisioned VM in ERROR state"
	CauseErroredVolume  Cause = "Provisioned Volume in ERROR state"
	CauseMachineTimeout Cause = "Provisioned VM in BUILD state after 30m0s"
)
