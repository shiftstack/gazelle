package rca

type Cause string

const (
	CauseErroredVM     Cause = "Provisioned VM in ERROR state"
	CauseErroredVolume Cause = "Provisioned Volume in ERROR state"
)
