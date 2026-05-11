package agentprompt

type TrustPolicy int

const (
	TrustRelaxed TrustPolicy = iota
	TrustStandard
	TrustProduction
)

func (t TrustPolicy) String() string {
	switch t {
	case TrustRelaxed:
		return "relaxed"
	case TrustStandard:
		return "standard"
	case TrustProduction:
		return "production"
	default:
		return "unknown"
	}
}
