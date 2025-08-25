package go_raptor

import (
	"fmt"
	"strings"
)

/**
 * whenever this type is used we are referring to a globally unique identifier for the object
 * this will likely differ from the ingested GTFS IDs because they are only guaranteed to be unique within a feed
 * but not across feeds - however to make an efficient RAPTOR calculation we will be operating on lists of stops/stoptimes etc.. from multiple feeds
 */
type UniqueGtfsIdLike interface {
	uint32 | uint64 | int32 | int64 | string
}

/** we will usually want to operate on times in seconds since the start of the day - this makes for easy comparisons */
type TimestampInSeconds = int64

type GtfsStop[ID UniqueGtfsIdLike] interface {
	GetUniqueID() ID
}

type GtfsTransfer[ID UniqueGtfsIdLike] interface {
	GetFromUniqueStopID() ID
	GetToUniqueStopID() ID
	GetMinimumTransferTimeInSeconds() int
}

type GtfsStopTime[ID UniqueGtfsIdLike] interface {
	GetUniqueStopID() ID
	/* this is the unique trip ID - but could be repeated across days */
	GetUniqueTripID() ID
	/* this is the daily unique trip ID - should be actually unique */
	GetUniqueTripServiceID() ID
	GetStopSequence() int
	GetArrivalTimeInSeconds() TimestampInSeconds
	GetDepartureTimeInSeconds() TimestampInSeconds
}

type GtfsStopStruct[ID UniqueGtfsIdLike] struct {
	GtfsStop[ID]
	UniqueID ID
}

type GtfsTransferStruct[ID UniqueGtfsIdLike] struct {
	GtfsTransfer[ID]
	FromUniqueStopID             ID
	ToUniqueStopID               ID
	MinimumTransferTimeInSeconds int
}

type GtfsStopTimeStruct[ID UniqueGtfsIdLike] struct {
	GtfsStopTime[ID]
	UniqueStopID ID
	/* this is the trip ID which could be repeated across days */
	UniqueTripID ID
	/* this is the unique trip service ID which should not be repeated across days when allowing multi-day planning */
	UniqueTripServiceID    ID
	StopSequence           int
	ArrivalTimeInSeconds   TimestampInSeconds
	DepartureTimeInSeconds TimestampInSeconds
}

func (b GtfsStopStruct[T]) GetUniqueID() T {
	return b.UniqueID
}

func (b GtfsTransferStruct[T]) GetFromUniqueStopID() T {
	return b.FromUniqueStopID
}

func (b GtfsTransferStruct[T]) GetToUniqueStopID() T {
	return b.ToUniqueStopID
}

func (b GtfsTransferStruct[T]) GetMinimumTransferTimeInSeconds() int {
	return b.MinimumTransferTimeInSeconds
}

func (b GtfsStopTimeStruct[T]) GetUniqueStopID() T {
	return b.UniqueStopID
}

func (b GtfsStopTimeStruct[T]) GetUniqueTripID() T {
	return b.UniqueTripID
}

func (b GtfsStopTimeStruct[T]) GetUniqueTripServiceID() T {
	return b.UniqueTripServiceID
}

func (b GtfsStopTimeStruct[T]) GetStopSequence() int {
	return b.StopSequence
}

func (b GtfsStopTimeStruct[T]) GetArrivalTimeInSeconds() TimestampInSeconds {
	return b.ArrivalTimeInSeconds
}

func (b GtfsStopTimeStruct[T]) GetDepartureTimeInSeconds() TimestampInSeconds {
	return b.DepartureTimeInSeconds
}

type ViaTrip[ID UniqueGtfsIdLike] struct {
	UniqueTripID           ID
	UniqueTripServiceID    ID
	FromStopSequenceInTrip int
	ToStopSequenceInTrip   int
}

/**
 * this represents a single span in a round segment
 * this contains a piece of the journey it took to get to the current segment
 * - this essentially tells you the following:
 * -- we arrived at UniqueStopID at ArrivalTimeInSeconds
* -- we left UniqueStopID by taking UniqueTripID at DepartureTimeInSeconds
*/
type RoundSegmentSpan[ID UniqueGtfsIdLike] struct {
	FromUniqueStopID ID
	ToUniqueStopID   ID
	/* could be nil if walking transfer */
	ViaTrip                                *ViaTrip[ID]
	ArrivalTimeInSecondsToUniqueStopID     TimestampInSeconds
	DepartureTimeInSecondsFromUniqueStopID TimestampInSeconds
}

/**
 * a single segment of a round calculation - will be kept for each stop
 * this essentially will represent the earliest arrival time to this stop
 * and the chain it took to get to this stop
 * */
type RoundSegment[ID UniqueGtfsIdLike] struct {
	UniqueStopID         ID
	ArrivalTimeInSeconds TimestampInSeconds
	Spans                []RoundSegmentSpan[ID]
}

type Journey[ID UniqueGtfsIdLike] struct {
	ToUniqueStopID         ID
	FromUniqueStopID       ID
	DepartureTimeInSeconds TimestampInSeconds
	ArrivalTimeInSeconds   TimestampInSeconds
	Legs                   []RoundSegmentSpan[ID]
}

type StopTimePartitionsPartition struct {
	Timestamp TimestampInSeconds
	Index     int
}

type StopTimePartitions[ID UniqueGtfsIdLike] struct {
	Partitions                      []StopTimePartitionsPartition
	PartitionsByUniqueStopID        map[ID][]StopTimePartitionsPartition
	PartitionsByUniqueTripServiveID map[ID][]StopTimePartitionsPartition
}

type SimpleRaptorInput[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]] struct {
	FromStops []StopType
	ToStops   []StopType
	Transfers []TransferType
	StopTimes []StopTimeType
	Mode      RaptorMode
	/* will be used for either depart_at mode or arrive_by mode */
	TimeInSeconds    TimestampInSeconds
	MaximumTransfers int
	/* determines whether to allow walk-transferring more than once */
	AllowTransferHopping bool

	/* determines how to group times - defaults to 86400 seconds / per day */
	TimePartitionInterval TimestampInSeconds

	/** these can be passed if they are pre-calculated in memory before running raptor; useful for speeding up the actual raptor - uints refer to their list indexes from the input */
	TransfersByUniqueStopId        *map[ID][]int
	StopTimesByUniqueStopId        *map[ID][]int
	StopTimesByUniqueTripServiceId *map[ID][]int
	TimePartitions                 *StopTimePartitions[ID]
}

type PreparedRaptorInput[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]] struct {
	Input *SimpleRaptorInput[ID, StopType, TransferType, StopTimeType]

	FromStopsByUniqueStopId        map[ID]ID
	ToStopsByUniqueStopId          map[ID]ID
	TransfersByUniqueStopId        map[ID][]int
	StopTimesByUniqueStopId        map[ID][]int
	StopTimesByUniqueTripServiceId map[ID][]int
	TimePartitions                 StopTimePartitions[ID]
}

type RaptorMarkedStop[ID UniqueGtfsIdLike] struct {
	ID     ID
	Source RaptorMarkedStopSource
}

func (j RoundSegment[ID]) GetFingerPrint() string {
	parts := []string{}
	for _, leg := range j.Spans {
		tripID := ""
		if leg.ViaTrip != nil {
			tripID = fmt.Sprintf("%v", leg.ViaTrip.UniqueTripID)
		}
		parts = append(parts, fmt.Sprintf("%v|%v|%v", leg.FromUniqueStopID, tripID, leg.ToUniqueStopID))
	}
	return strings.Join(parts, "->")
}
