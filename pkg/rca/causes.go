package rca

type Cause string

const (
	CauseErroredVM      Cause = "Provisioned VM in ERROR state"
	CauseErroredVolume  Cause = "Provisioned Volume in ERROR state"
	CauseMachineTimeout Cause = "Provisioned VM in BUILD state after 30m0s"
	CauseClusterTimeout Cause = "Timeout waiting for cluster to initialize"
)

func (c Cause) IsInfra() bool {
	switch c {
	case CauseErroredVM, CauseErroredVolume, CauseMachineTimeout:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer
func (c Cause) String() string {
	return string(c)
}
