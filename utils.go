package go_raptor

func GetTimePartition(timestamp TimestampInSeconds, interval TimestampInSeconds, upper bool) TimestampInSeconds {
	lower := timestamp - (timestamp % interval)
	if !upper || lower == timestamp {
		return lower
	}
	return ((lower / interval) + 1) * interval
}
