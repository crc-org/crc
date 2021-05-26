package preflight

type LabelName uint32

const (
	Os LabelName = iota
	NetworkMode

	// Keep it last
	lastLabelName // will be used in OS-specific go files to extend LabelName
)

type LabelValue uint32

const (
	// os
	Darwin LabelValue = iota
	Linux
	Windows

	// network mode
	User
	System

	// Keep it last
	lastLabelValue // will be used in OS-specific go files to extend LabelValue
)

var (
	None = labels{}
)

type labels map[LabelName]LabelValue
