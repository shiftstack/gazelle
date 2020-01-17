package rca

import "fmt"

type Cause interface {
	IsInfra() bool

	fmt.Stringer
}

const (
	CauseErroredVM      CauseInfra = "Provisioned VM in ERROR state"
	CauseErroredVolume  CauseInfra = "Provisioned Volume in ERROR state"
	CauseMachineTimeout CauseInfra = "Provisioned VM in BUILD state after 30m0s"
	CauseReleaseImage   CauseInfra = "Unable to import release image"

	CauseClusterTimeout CauseGeneric = "Timeout waiting for cluster to initialize"
)

type CauseGeneric string

func (c CauseGeneric) IsInfra() bool {
	return false
}

// String implements fmt.Stringer
func (c CauseGeneric) String() string {
	return string(c)
}

type CauseInfra string

func (c CauseInfra) IsInfra() bool {
	return true
}

// String implements fmt.Stringer
func (c CauseInfra) String() string {
	return string(c)
}

type CauseQuota string

func (c CauseQuota) IsInfra() bool {
	return true
}

// String implements fmt.Stringer
func (c CauseQuota) String() string {
	return "Quota exceeded for resource: " + string(c)
}
