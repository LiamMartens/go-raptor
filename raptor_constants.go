package go_raptor

type RaptorMode string
type RaptorMarkedStopSource = string

const (
	RaptorModeDepartAt RaptorMode = "depart_at"
	RaptorModeArriveBy RaptorMode = "arrive_by"
)

const (
	RaptorMarkedStopSourceArrival  RaptorMarkedStopSource = "arrival"
	RaptorMarkedStopSourceTransfer RaptorMarkedStopSource = "transfer"
)
