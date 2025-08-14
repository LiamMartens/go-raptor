package goraptor

/**
 * whenever this type is used we are referring to a globally unique identifier for the object
 * this will likely differ from the ingested GTFS IDs because they are only guaranteed to be unique within a feed
 * but not across feeds - however to make an efficient RAPTOR calculation we will be operating on lists of stops/stoptimes etc.. from multiple feeds
 */
type UniqueGtfsIdLike interface {
	int | string
}

/** we will usually want to operate on times in seconds since the start of the day - this makes for easy comparisons */
type TimeInSecondsSinceStartOfDay = int

type GtfsStop[ID UniqueGtfsIdLike] interface {
	GetUniqueID() ID
}

type GtfsTransfer[ID UniqueGtfsIdLike] interface {
	GetFromUniqueStopID() ID
	GetToUniqueStopID() ID
	GetMinimumTransferTimeInSeconds() int
}

type GtfsStopTime[ID UniqueGtfsIdLike] interface {
	/** this is the globally unique stop id of this stop time entry */
	GetUniqueStopID() ID
	/** this is the globally unique trip id for this stop time entry */
	GetUniqueTripID() ID
	GetStopSequence() int
	GetArrivalTimeInSeconds() TimeInSecondsSinceStartOfDay
	GetDepartureTimeInSeconds() TimeInSecondsSinceStartOfDay
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
	UniqueStopID           ID
	UniqueTripID           ID
	StopSequence           int
	ArrivalTimeInSeconds   int
	DepartureTimeInSeconds int
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

func (b GtfsStopTimeStruct[T]) GetStopSequence() int {
	return b.StopSequence
}

func (b GtfsStopTimeStruct[T]) GetArrivalTimeInSeconds() int {
	return b.ArrivalTimeInSeconds
}

func (b GtfsStopTimeStruct[T]) GetDepartureTimeInSeconds() int {
	return b.DepartureTimeInSeconds
}

/**
 * this represents a single span in a round segment
 * this contains a piece of the journey it took to get to the current segment
 * - this essentially tells you the following:
 * -- we arrived at UniqueStopID at ArrivalTimeInSeconds
* -- we left UniqueStopID by taking UniqueTripID at DepartureTimeInSeconds
*/
type RoundSegmentSpan[ID UniqueGtfsIdLike] struct {
	UniqueStopID           ID
	UniqueTripID           ID
	StopSequenceInTrip     int
	ArrivalTimeInSeconds   TimeInSecondsSinceStartOfDay
	DepartureTimeInSeconds TimeInSecondsSinceStartOfDay
}

/**
 * a single segment of a round calculation - will be kept for each stop
 * this essentially will represent the earliest arrival time to this stop
 * and the chain it took to get to this stop
 * */
type RoundSegment[ID UniqueGtfsIdLike] struct {
	UniqueStopID         ID
	ArrivalTimeInSeconds TimeInSecondsSinceStartOfDay
	Spans                []RoundSegmentSpan[ID]
}

type RaptorMode string

const (
	RaptorModeDepartAt RaptorMode = "depart_at"
	RaptorModeArriveBy RaptorMode = "arrive_by"
)

type SimpleRaptorInput[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]] struct {
	FromStops []StopType
	ToStops   []StopType
	Transfers []TransferType
	StopTimes []StopTimeType
	Mode      RaptorMode
	/* expected YYYYMMDD */
	DateOfService string
	/* will be used for either depart_at mode or arrive_by mode */
	TimeInSeconds    TimeInSecondsSinceStartOfDay
	MaximumTransfers int
}

type PreparedRaptorInput[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]] struct {
	Input SimpleRaptorInput[ID, StopType, TransferType, StopTimeType]

	FromStopsByUniqueStopId map[ID]ID
	ToStopsByUniqueStopId   map[ID]ID
	TransfersByUniqueStopId map[ID][]TransferType
	StopTimesByUniqueStopId map[ID][]StopTimeType
	StopTimesByUniqueTripId map[ID][]StopTimeType
}

func PrepareRaptorInput[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]](
	input SimpleRaptorInput[ID, StopType, TransferType, StopTimeType],
) PreparedRaptorInput[ID, StopType, TransferType, StopTimeType] {
	/** prepares the raptor input with additional lookup maps */

	/** create a map of to_stops by unique ID for easy lookup */
	to_stops_by_unique_stop_id := map[ID]ID{}
	for _, stop := range input.ToStops {
		to_stops_by_unique_stop_id[stop.GetUniqueID()] = stop.GetUniqueID()
	}

	/** create a map of from_stops by unique ID for easy lookup */
	from_stops_by_unique_stop_id := map[ID]ID{}
	for _, stop := range input.FromStops {
		from_stops_by_unique_stop_id[stop.GetUniqueID()] = stop.GetUniqueID()
	}

	/** create a map of transfers from stop IDs for easy lookup */
	transfers_by_unique_stop_id := map[ID][]TransferType{}
	for _, transfer := range input.Transfers {
		if _, has_key := transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()]; !has_key {
			transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()] = []TransferType{}
		}
		transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()] = append(transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()], transfer)
	}

	/** create a map of stop times by stop id and by trip id for easy lookup */
	stop_times_by_unique_stop_id := map[ID][]StopTimeType{}
	stop_times_by_unique_trip_id := map[ID][]StopTimeType{}
	for _, stop_time := range input.StopTimes {
		if _, has_key := stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()]; !has_key {
			stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()] = []StopTimeType{}
		}
		if _, has_key := stop_times_by_unique_trip_id[stop_time.GetUniqueTripID()]; !has_key {
			stop_times_by_unique_trip_id[stop_time.GetUniqueTripID()] = []StopTimeType{}
		}
		stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()] = append(stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()], stop_time)
		stop_times_by_unique_trip_id[stop_time.GetUniqueTripID()] = append(stop_times_by_unique_trip_id[stop_time.GetUniqueTripID()], stop_time)
	}

	return PreparedRaptorInput[ID, StopType, TransferType, StopTimeType]{
		Input:                   input,
		FromStopsByUniqueStopId: from_stops_by_unique_stop_id,
		ToStopsByUniqueStopId:   to_stops_by_unique_stop_id,
		TransfersByUniqueStopId: transfers_by_unique_stop_id,
		StopTimesByUniqueStopId: stop_times_by_unique_stop_id,
		StopTimesByUniqueTripId: stop_times_by_unique_trip_id,
	}
}

/**
 * below are the basic raptor implementations using either depart_at and arrive_by
 * the logic is generally the same but they are reversed in their iteration due to the arrive by or depart at conditions
 * this assumes all the stop times are valid for the service on the requested date -- thus before calling stop times should be filtered
 * according to the gtfs calendar / services - this implementation only deals with Raptor
 * additionally stop times are expected to be ordered in ascending order by their stop sequence
 */

func SimpleRaptorDepartAt[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]](
	input SimpleRaptorInput[ID, StopType, TransferType, StopTimeType],
) []RoundSegment[ID] {
	prepared_input := PrepareRaptorInput(input)

	/* below is the start of the raptor based algorithm */
	/* this map contains the earliest arrival time at each stop across rounds - keeping track of all the segments */
	earliest_arrival_time_segments_by_unique_stop_id := map[ID]RoundSegment[ID]{}
	/* this is the result slice which contains all the potential journeys (meaning segments which reach the end destination) */
	potential_journeys_found := []RoundSegment[ID]{}

	/* to start we need to mark which stops we are going to check during the current round - at the start this will only be the from stops */
	/* this will be replaced between rounds because we will be checking the next set of transferred to stops */
	stops_marked_for_round := make(map[ID]ID, len(input.FromStops))
	for _, stop := range input.FromStops {
		stops_marked_for_round[stop.GetUniqueID()] = stop.GetUniqueID()
	}

	/* we will also initialize the initial segments for the from_stops -> essentially saying we have arrived at said stops at the depart_at time */
	for _, from_stop := range input.FromStops {
		earliest_arrival_time_segments_by_unique_stop_id[from_stop.GetUniqueID()] = RoundSegment[ID]{
			UniqueStopID:         from_stop.GetUniqueID(),
			ArrivalTimeInSeconds: input.TimeInSeconds,
			/* we arrived here "as-is" so no spans yet */
			Spans: []RoundSegmentSpan[ID]{},
		}
	}

	/* now we can start the rounds up until N transfers */
	for range input.MaximumTransfers {
		/* this will be the set of next stops to check for the next round */
		stops_marked_for_next_round := map[ID]ID{}
		/* in each round we will check all marked stops for the trips we could take - we will do this by going through the stop times */
		for _, marked_stop_unique_id := range stops_marked_for_round {
			/* this should always exist because any marked stop should have been added to the segment list */
			current_segment_for_stop := earliest_arrival_time_segments_by_unique_stop_id[marked_stop_unique_id]
			stop_times_for_marked_stop := prepared_input.StopTimesByUniqueStopId[marked_stop_unique_id]
			/* we will go through the stop times and find the departures we can make based on our current earliest arrival time at the  "marked_stop" */
			for _, stop_time_for_marked_stop := range stop_times_for_marked_stop {
				if stop_time_for_marked_stop.GetDepartureTimeInSeconds() < current_segment_for_stop.ArrivalTimeInSeconds {
					/* if the departure time of this stop time happens before my earliest arrival time - I won't be able to make it -> skipping */
					continue
				}
				/*
				 * if we CAN make it we will want to look up the stop times after the current one in the trip.
				 * we're essentially just going down the line and storing each stop time if the arrival time is earlier than the currently stored one
				 * (meaning I could get to this stop earlier than initially expected)
				 */
				stop_times_for_unique_trip_id_after_current_stop := prepared_input.StopTimesByUniqueTripId[stop_time_for_marked_stop.GetUniqueTripID()][stop_time_for_marked_stop.GetStopSequence():]
				/* the stop times are expected to be in order of sequence ascending */
			following_stop_times_loop:
				for _, following_stop_time := range stop_times_for_unique_trip_id_after_current_stop {
					existing_segment, has_existing_segment := earliest_arrival_time_segments_by_unique_stop_id[following_stop_time.GetUniqueStopID()]
					/* if this stop was not arrived at yet OR if this arrival is before the recorded arrival */
					if !has_existing_segment || existing_segment.ArrivalTimeInSeconds > following_stop_time.GetArrivalTimeInSeconds() {
						updated_spans := make([]RoundSegmentSpan[ID], len(current_segment_for_stop.Spans)+1)
						/* copy current segment spans + add a new span for how to get to this stop */
						copy(updated_spans, current_segment_for_stop.Spans)
						updated_spans[len(updated_spans)-1] = RoundSegmentSpan[ID]{
							UniqueStopID:           stop_time_for_marked_stop.GetUniqueStopID(),
							UniqueTripID:           stop_time_for_marked_stop.GetUniqueTripID(),
							StopSequenceInTrip:     stop_time_for_marked_stop.GetStopSequence(),
							ArrivalTimeInSeconds:   stop_time_for_marked_stop.GetArrivalTimeInSeconds(),
							DepartureTimeInSeconds: stop_time_for_marked_stop.GetDepartureTimeInSeconds(),
						}
						earliest_arrival_time_segments_by_unique_stop_id[following_stop_time.GetUniqueStopID()] = RoundSegment[ID]{
							UniqueStopID:         following_stop_time.GetUniqueStopID(),
							ArrivalTimeInSeconds: following_stop_time.GetArrivalTimeInSeconds(),
							Spans:                updated_spans,
						}
					}
					/* next we can mark this stop to check in the next round AND add any potential transfers from this stop to mark */
					stops_marked_for_next_round[following_stop_time.GetUniqueStopID()] = following_stop_time.GetUniqueStopID()
					potential_transfers_for_stop := prepared_input.TransfersByUniqueStopId[following_stop_time.GetUniqueStopID()]
					for _, transfer_stop := range potential_transfers_for_stop {
						stops_marked_for_next_round[transfer_stop.GetToUniqueStopID()] = transfer_stop.GetToUniqueStopID()
						/* for each transferrable station we'll also add an earliest arrival segment which is the current arrival time + the minimum transfer time (if the arrival is earlier than the previously recorded one) */
						arrival_time_at_transfer_stop := following_stop_time.GetArrivalTimeInSeconds() + transfer_stop.GetMinimumTransferTimeInSeconds()
						existing_transfer_segment, has_existing_transfer_segment := earliest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()]
						if !has_existing_transfer_segment || existing_transfer_segment.ArrivalTimeInSeconds > arrival_time_at_transfer_stop {
							/* copy current segment spans from the original arrival station + add a new one for the transfer itself */
							updated_spans := make([]RoundSegmentSpan[ID], len(existing_segment.Spans)+1)
							copy(updated_spans, existing_segment.Spans)
							updated_spans[len(updated_spans)-1] = RoundSegmentSpan[ID]{
								UniqueStopID:           following_stop_time.GetUniqueStopID(),
								UniqueTripID:           following_stop_time.GetUniqueTripID(),
								StopSequenceInTrip:     following_stop_time.GetStopSequence(),
								ArrivalTimeInSeconds:   following_stop_time.GetArrivalTimeInSeconds(),
								DepartureTimeInSeconds: following_stop_time.GetDepartureTimeInSeconds(),
							}
							earliest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()] = RoundSegment[ID]{
								UniqueStopID:         transfer_stop.GetToUniqueStopID(),
								ArrivalTimeInSeconds: arrival_time_at_transfer_stop,
								Spans:                updated_spans,
							}
						}
					}
					/* lastly we can check if this stop is actually one of our destination stops - in which case the segment is corresponding to a complete journe7 */
					stops_which_could_be_destination := []ID{following_stop_time.GetUniqueStopID()}
					for _, transfer_stop := range potential_transfers_for_stop {
						stops_which_could_be_destination = append(stops_which_could_be_destination, transfer_stop.GetToUniqueStopID())
					}

					for _, potential_destination_stop := range stops_which_could_be_destination {
						if _, is_destination_stop := prepared_input.ToStopsByUniqueStopId[potential_destination_stop]; is_destination_stop {
							segment := earliest_arrival_time_segments_by_unique_stop_id[potential_destination_stop]
							segment_spans := make([]RoundSegmentSpan[ID], len(segment.Spans))
							copy(segment_spans, segment.Spans)
							potential_journeys_found = append(potential_journeys_found, RoundSegment[ID]{
								UniqueStopID:         segment.UniqueStopID,
								ArrivalTimeInSeconds: segment.ArrivalTimeInSeconds,
								Spans:                segment_spans,
							})

							/* this also means we can stop this loop */
							break following_stop_times_loop
						}
					}
				}
			}
		}
		/* replace stops marked map */
		stops_marked_for_round = stops_marked_for_next_round
	}

	return potential_journeys_found
}

func SimpleRaptorArriveBy[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]](
	input SimpleRaptorInput[ID, StopType, TransferType, StopTimeType],
) []RoundSegment[ID] {
	/* for arrive by we will always want to iterate stop times in reverse - so let's reverse it once before preparing the input */
	num_stop_times := len(input.StopTimes)
	stop_times_in_reverse := make([]StopTimeType, num_stop_times)
	for i, v := range input.StopTimes {
		stop_times_in_reverse[num_stop_times-1-i] = v
	}
	prepared_input := PrepareRaptorInput(SimpleRaptorInput[ID, StopType, TransferType, StopTimeType]{
		FromStops:        input.FromStops,
		ToStops:          input.ToStops,
		Transfers:        input.Transfers,
		Mode:             input.Mode,
		DateOfService:    input.DateOfService,
		TimeInSeconds:    input.TimeInSeconds,
		MaximumTransfers: input.MaximumTransfers,
		StopTimes:        stop_times_in_reverse,
	})

	/* below is the start of the raptor based algorithm */
	/* this map contains the latest possible arrival time at each stop across rounds (nearest to the arrive by time) - keeping track of all the segments */
	latest_arrival_time_segments_by_unique_stop_id := map[ID]RoundSegment[ID]{}
	/* this is the result slice which contains all the potential journeys (meaning segments which reach the end destination) */
	potential_journeys_found := []RoundSegment[ID]{}

	/* to start we need to mark which stops we are going to check during the current round - at the start this will only be the destinations stops */
	/* this will be replaced between rounds because we will be checking the next set of transferred to stops */
	stops_marked_for_round := make(map[ID]ID, len(input.ToStops))
	for _, stop := range input.ToStops {
		stops_marked_for_round[stop.GetUniqueID()] = stop.GetUniqueID()
	}

	/* we will also initialize the initial segments for the to_stops -> essentially saying we have not been able to arrive yet */
	for _, to_stop := range input.ToStops {
		latest_arrival_time_segments_by_unique_stop_id[to_stop.GetUniqueID()] = RoundSegment[ID]{
			UniqueStopID:         to_stop.GetUniqueID(),
			ArrivalTimeInSeconds: input.TimeInSeconds,
			/* we arrived here "as-is" so no spans yet */
			Spans: []RoundSegmentSpan[ID]{},
		}
	}

	/* now we can start the rounds up until N transfers */
	for range input.MaximumTransfers {
		/* this will be the set of next stops to check for the next round */
		stops_marked_for_next_round := map[ID]ID{}
		/* in each round we will check all marked stops for the trips we could take - we will do this by going through the stop times in  reverse */
		for _, marked_stop_unique_id := range stops_marked_for_round {
			/* this should always exist because any marked stop should have been added to the segment list */
			current_segment_for_stop := latest_arrival_time_segments_by_unique_stop_id[marked_stop_unique_id]
			stop_times_for_marked_stop := prepared_input.StopTimesByUniqueStopId[marked_stop_unique_id]
			/* we will go through the stop times and find the latest arrivals which are still before my expected can make based on our current earliest arrival time at the  "marked_stop" */
			for _, stop_time_for_marked_stop := range stop_times_for_marked_stop {
				if stop_time_for_marked_stop.GetArrivalTimeInSeconds() > current_segment_for_stop.ArrivalTimeInSeconds {
					/* if the arrival time of this stop time happens after the current segment arrival time then we are too late */
					continue
				}
				/*
				 * if we CAN make it we will want to look up the stop times before the current one in the trip.
				 * we're essentially just going down the line in reverse and storing each stop time if the arrival time is later than the currently stored one
				 * (meaning I could get to this stop later than initially expected)
				 */
				stop_times_for_unique_trip_id := prepared_input.StopTimesByUniqueTripId[stop_time_for_marked_stop.GetUniqueTripID()]
				/* to get these we want to reverse the stop sequence and skip one to exclude my current stop which I already checked */
				stop_times_for_unique_trip_id_after_current_stop := stop_times_for_unique_trip_id[len(stop_times_for_unique_trip_id)-stop_time_for_marked_stop.GetStopSequence()+1:]
				/* the stop times are expected to be in order of sequence descending */
			preceeding_stop_times_loop:
				for _, preceeding_stop_time := range stop_times_for_unique_trip_id_after_current_stop {
					existing_segment, has_existing_segment := latest_arrival_time_segments_by_unique_stop_id[preceeding_stop_time.GetUniqueStopID()]
					/* if this stop was not arrived at yet OR if this arrival is after the recorded arrival */
					if !has_existing_segment || preceeding_stop_time.GetArrivalTimeInSeconds() > existing_segment.ArrivalTimeInSeconds {
						/* copy current segment spans + add a new span for how to get to this stop - we'll be adding them at the start to make sure it's chronological */
						updated_spans := append([]RoundSegmentSpan[ID]{
							{
								UniqueStopID:           stop_time_for_marked_stop.GetUniqueStopID(),
								UniqueTripID:           stop_time_for_marked_stop.GetUniqueTripID(),
								StopSequenceInTrip:     stop_time_for_marked_stop.GetStopSequence(),
								ArrivalTimeInSeconds:   stop_time_for_marked_stop.GetArrivalTimeInSeconds(),
								DepartureTimeInSeconds: stop_time_for_marked_stop.GetDepartureTimeInSeconds(),
							},
						}, current_segment_for_stop.Spans...)
						latest_arrival_time_segments_by_unique_stop_id[preceeding_stop_time.GetUniqueStopID()] = RoundSegment[ID]{
							UniqueStopID:         preceeding_stop_time.GetUniqueStopID(),
							ArrivalTimeInSeconds: preceeding_stop_time.GetArrivalTimeInSeconds(),
							Spans:                updated_spans,
						}
					}
					/* next we can mark this stop to check in the next round AND add any potential transfers from this stop to mark */
					stops_marked_for_next_round[preceeding_stop_time.GetUniqueStopID()] = preceeding_stop_time.GetUniqueStopID()
					potential_transfers_for_stop := prepared_input.TransfersByUniqueStopId[preceeding_stop_time.GetUniqueStopID()]
					for _, transfer_stop := range potential_transfers_for_stop {
						stops_marked_for_next_round[transfer_stop.GetToUniqueStopID()] = transfer_stop.GetToUniqueStopID()
						/* for each transferrable station we'll also add a latest arrival segment which is the current arrival time - the minimum transfer time (if the arrival is later than the previously recorded one) */
						arrival_time_at_transfer_stop := preceeding_stop_time.GetArrivalTimeInSeconds() - transfer_stop.GetMinimumTransferTimeInSeconds()
						existing_transfer_segment, has_existing_transfer_segment := latest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()]
						if !has_existing_transfer_segment || arrival_time_at_transfer_stop > existing_transfer_segment.ArrivalTimeInSeconds {
							/* copy current segment spans from the original arrival station + add a new one for the transfer itself */
							updated_spans := append([]RoundSegmentSpan[ID]{
								{
									UniqueStopID:           preceeding_stop_time.GetUniqueStopID(),
									UniqueTripID:           preceeding_stop_time.GetUniqueTripID(),
									StopSequenceInTrip:     preceeding_stop_time.GetStopSequence(),
									ArrivalTimeInSeconds:   preceeding_stop_time.GetArrivalTimeInSeconds(),
									DepartureTimeInSeconds: preceeding_stop_time.GetDepartureTimeInSeconds(),
								},
							}, existing_segment.Spans...)
							latest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()] = RoundSegment[ID]{
								UniqueStopID:         transfer_stop.GetToUniqueStopID(),
								ArrivalTimeInSeconds: arrival_time_at_transfer_stop,
								Spans:                updated_spans,
							}
						}
					}
					/* lastly we can check if this stop is actually one of our origin stops - in which case the segment is corresponding to a complete journe7 */
					stops_which_could_be_origin := []ID{preceeding_stop_time.GetUniqueStopID()}
					for _, transfer_stop := range potential_transfers_for_stop {
						stops_which_could_be_origin = append(stops_which_could_be_origin, transfer_stop.GetToUniqueStopID())
					}

					for _, potential_origin_stop := range stops_which_could_be_origin {
						if _, is_origin_stop := prepared_input.ToStopsByUniqueStopId[potential_origin_stop]; is_origin_stop {
							segment := latest_arrival_time_segments_by_unique_stop_id[potential_origin_stop]
							segment_spans := make([]RoundSegmentSpan[ID], len(segment.Spans))
							copy(segment_spans, segment.Spans)
							potential_journeys_found = append(potential_journeys_found, RoundSegment[ID]{
								UniqueStopID:         segment.UniqueStopID,
								ArrivalTimeInSeconds: segment.ArrivalTimeInSeconds,
								Spans:                segment_spans,
							})

							/* this also means we can stop this loop */
							break preceeding_stop_times_loop
						}
					}
				}
			}
		}
		/* replace stops marked map */
		stops_marked_for_round = stops_marked_for_next_round
	}

	return potential_journeys_found
}

func SimpleRaptor[ID UniqueGtfsIdLike, StopType GtfsStop[ID], TransferType GtfsTransfer[ID], StopTimeType GtfsStopTime[ID]](
	input SimpleRaptorInput[ID, StopType, TransferType, StopTimeType],
) []RoundSegment[ID] {
	if input.Mode == RaptorModeDepartAt {
		return SimpleRaptorDepartAt(input)
	}
	return SimpleRaptorArriveBy(input)
}
