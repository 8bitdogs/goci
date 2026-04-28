package payload

type Request interface {
	TargetBranch() string

	FullName() string

	// Commit SHA of the push event
	Sha() string
}
