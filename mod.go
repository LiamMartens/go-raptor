package go_raptor

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
	transfers_by_unique_stop_id := map[ID][]int{}
	if input.TransfersByUniqueStopId != nil {
		transfers_by_unique_stop_id = *input.TransfersByUniqueStopId
	} else {
		for index, transfer := range input.Transfers {
			if _, has_key := transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()]; !has_key {
				transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()] = []int{}
			}
			transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()] = append(transfers_by_unique_stop_id[transfer.GetFromUniqueStopID()], index)
		}
	}

	/** create a map of stop times by stop id and by trip id for easy lookup */
	has_prepared_stop_times_by_unique_stop_id := input.StopTimesByUniqueStopId != nil
	has_prepared_stop_times_by_unique_trip_service_id := input.StopTimesByUniqueTripServiceId != nil
	stop_times_by_unique_stop_id := map[ID][]int{}
	stop_times_by_unique_trip_service_id := map[ID][]int{}
	if has_prepared_stop_times_by_unique_stop_id {
		stop_times_by_unique_stop_id = *input.StopTimesByUniqueStopId
	}
	if has_prepared_stop_times_by_unique_trip_service_id {
		stop_times_by_unique_trip_service_id = *input.StopTimesByUniqueTripServiceId
	}
	if !has_prepared_stop_times_by_unique_stop_id || !has_prepared_stop_times_by_unique_trip_service_id {
		for index, stop_time := range input.StopTimes {
			if !has_prepared_stop_times_by_unique_stop_id {
				if _, has_key := stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()]; !has_key {
					stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()] = []int{}
				}
				stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()] = append(stop_times_by_unique_stop_id[stop_time.GetUniqueStopID()], index)
			}

			if !has_prepared_stop_times_by_unique_trip_service_id {
				if _, has_key := stop_times_by_unique_trip_service_id[stop_time.GetUniqueTripServiceID()]; !has_key {
					stop_times_by_unique_trip_service_id[stop_time.GetUniqueTripServiceID()] = []int{}
				}
				stop_times_by_unique_trip_service_id[stop_time.GetUniqueTripServiceID()] = append(stop_times_by_unique_trip_service_id[stop_time.GetUniqueTripServiceID()], index)
			}
		}
	}

	return PreparedRaptorInput[ID, StopType, TransferType, StopTimeType]{
		Input:                          &input,
		FromStopsByUniqueStopId:        from_stops_by_unique_stop_id,
		ToStopsByUniqueStopId:          to_stops_by_unique_stop_id,
		TransfersByUniqueStopId:        transfers_by_unique_stop_id,
		StopTimesByUniqueStopId:        stop_times_by_unique_stop_id,
		StopTimesByUniqueTripServiceId: stop_times_by_unique_trip_service_id,
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
) []Journey[ID] {
	prepared_input := PrepareRaptorInput(input)

	/* below is the start of the raptor based algorithm */
	/* this map contains the earliest arrival time at each stop across rounds - keeping track of all the segments */
	earliest_arrival_time_segments_by_unique_stop_id := map[ID]RoundSegment[ID]{}
	/* this is the result slice which contains all the potential journeys (meaning segments which reach the end destination) */
	potential_journeys_found := []Journey[ID]{}
	potential_journey_fingerprints := map[string]bool{}

	/* we will also initialize the initial segments for the from_stops -> essentially saying we have arrived at said stops at the depart_at time */
	for _, from_stop := range input.FromStops {
		earliest_arrival_time_segments_by_unique_stop_id[from_stop.GetUniqueID()] = RoundSegment[ID]{
			UniqueStopID:         from_stop.GetUniqueID(),
			ArrivalTimeInSeconds: input.TimeInSeconds,
			/* we arrived here "as-is" so no spans yet */
			Spans: []RoundSegmentSpan[ID]{},
		}
	}

	/* to start we need to mark which stops we are going to check during the current round - at the start this will only be the from stops */
	/* this will be replaced between rounds because we will be checking the next set of transferred to stops */
	stops_marked_for_round := make(map[ID]RaptorMarkedStop[ID], len(input.FromStops))
	for _, stop := range input.FromStops {
		stops_marked_for_round[stop.GetUniqueID()] = RaptorMarkedStop[ID]{
			ID:     stop.GetUniqueID(),
			Source: RaptorMarkedStopSourceArrival,
		}
	}

	/* now we can start the rounds up until N transfers */
	trips_scanned_from_sequence := map[ID]int{}
	for range input.MaximumTransfers {
		/* this will be the set of next stops to check for the next round */
		stops_marked_for_next_round := map[ID]RaptorMarkedStop[ID]{}
		/* in each round we will check all marked stops for the trips we could take - we will do this by going through the stop times */
		for _, marked_stop := range stops_marked_for_round {
			/* this should always exist because any marked stop should have been added to the segment list */
			current_segment_for_stop := earliest_arrival_time_segments_by_unique_stop_id[marked_stop.ID]
			stop_times_for_marked_stop := prepared_input.StopTimesByUniqueStopId[marked_stop.ID]
			/* we will go through the stop times and find the departures we can make based on our current earliest arrival time at the  "marked_stop" */
			stop_times_for_marked_stop_it := NewSliceIterator(stop_times_for_marked_stop, false)
			for stop_times_for_marked_stop_it.HasNext() {
				stop_time_for_marked_stop := prepared_input.Input.StopTimes[stop_times_for_marked_stop_it.Next()]
				trip_already_scanned_from_sequence, has_already_scanned_trip_from_sequence := trips_scanned_from_sequence[stop_time_for_marked_stop.GetUniqueTripID()]
				/* skip scanning if trip was already forward scanned past or from this sequence */
				if stop_time_for_marked_stop.GetDepartureTimeInSeconds() < current_segment_for_stop.ArrivalTimeInSeconds ||
					has_already_scanned_trip_from_sequence && stop_time_for_marked_stop.GetStopSequence() >= trip_already_scanned_from_sequence {
					/* if the departure time of this stop time happens before my earliest arrival time - I won't be able to make it -> skipping */
					continue
				}

				/* mark trip as scanned from sequence */
				trips_scanned_from_sequence[stop_time_for_marked_stop.GetUniqueTripID()] = stop_time_for_marked_stop.GetStopSequence()

				/*
				 * if we CAN make it we will want to look up the stop times after the current one in the trip.
				 * we're essentially just going down the line and storing each stop time if the arrival time is earlier than the currently stored one
				 * (meaning I could get to this stop earlier than initially expected)
				 */

				/* we want to only take the required slice ; ie if we already scanned some stop times after the current sequence we only need to check the missing ones */
				/* we'll also check the sequence offset because the input may have omitted a number of irrelevant stop times at the start */
				var stop_times_for_unique_trip_id_after_current_stop_it *SliceIterator[int]
				stop_times_for_unique_trip_id_it := NewSliceIterator(prepared_input.StopTimesByUniqueTripServiceId[stop_time_for_marked_stop.GetUniqueTripServiceID()], false)
				trip_stop_times_sequence_offset := prepared_input.Input.StopTimes[stop_times_for_unique_trip_id_it.First()].GetStopSequence()
				/* we want to subtract the first stop time sequence and add 1 to skip the current one if the current one is the same */
				stop_times_start_offset := stop_time_for_marked_stop.GetStopSequence() - trip_stop_times_sequence_offset + 1
				stop_times_end_offset := trip_already_scanned_from_sequence - trip_stop_times_sequence_offset
				if !has_already_scanned_trip_from_sequence {
					stop_times_for_unique_trip_id_after_current_stop_it = stop_times_for_unique_trip_id_it.SliceIterator(stop_times_start_offset, stop_times_for_unique_trip_id_it.Length())
				} else {
					stop_times_for_unique_trip_id_after_current_stop_it = stop_times_for_unique_trip_id_it.SliceIterator(stop_times_start_offset, stop_times_end_offset)
				}

				/* the stop times are expected to be in order of sequence ascending */
			following_stop_times_loop:
				for stop_times_for_unique_trip_id_after_current_stop_it.HasNext() {
					following_stop_time := prepared_input.Input.StopTimes[stop_times_for_unique_trip_id_after_current_stop_it.Next()]
					existing_segment, has_existing_segment := earliest_arrival_time_segments_by_unique_stop_id[following_stop_time.GetUniqueStopID()]
					is_improvement_to_existing_arrival_time := !has_existing_segment || existing_segment.ArrivalTimeInSeconds > following_stop_time.GetArrivalTimeInSeconds()
					/* if this stop was not arrived at yet OR if this arrival is before the recorded arrival */
					if is_improvement_to_existing_arrival_time {
						updated_spans := make([]RoundSegmentSpan[ID], len(current_segment_for_stop.Spans)+1)
						/* copy current segment spans + add a new span for how to get to this stop */
						copy(updated_spans, current_segment_for_stop.Spans)
						updated_spans[len(updated_spans)-1] = RoundSegmentSpan[ID]{
							FromUniqueStopID: stop_time_for_marked_stop.GetUniqueStopID(),
							ToUniqueStopID:   following_stop_time.GetUniqueStopID(),
							ViaTrip: &ViaTrip[ID]{
								UniqueTripID:           following_stop_time.GetUniqueTripID(),
								UniqueTripServiceID:    following_stop_time.GetUniqueTripServiceID(),
								FromStopSequenceInTrip: stop_time_for_marked_stop.GetStopSequence(),
								ToStopSequenceInTrip:   following_stop_time.GetStopSequence(),
							},
							DepartureTimeInSecondsFromUniqueStopID: stop_time_for_marked_stop.GetDepartureTimeInSeconds(),
							ArrivalTimeInSecondsToUniqueStopID:     following_stop_time.GetArrivalTimeInSeconds(),
						}
						earliest_arrival_time_segments_by_unique_stop_id[following_stop_time.GetUniqueStopID()] = RoundSegment[ID]{
							UniqueStopID:         following_stop_time.GetUniqueStopID(),
							ArrivalTimeInSeconds: following_stop_time.GetArrivalTimeInSeconds(),
							Spans:                updated_spans,
						}
						/* update existing segment in place for later */
						existing_segment = earliest_arrival_time_segments_by_unique_stop_id[following_stop_time.GetUniqueStopID()]

						/* only allow looking for transfers again if transfer hopping is allowed or the currently marked stop was arrived at by a trip not by a transfer */
						if input.AllowTransferHopping || marked_stop.Source == RaptorMarkedStopSourceArrival {
							potential_transfers_for_stop := prepared_input.TransfersByUniqueStopId[following_stop_time.GetUniqueStopID()]
							for _, transfer_stop_index := range potential_transfers_for_stop {
								transfer_stop := prepared_input.Input.Transfers[transfer_stop_index]
								/* we don't want to override a direct arrival marked stop */
								if _, has_already_marked_stop := stops_marked_for_next_round[transfer_stop.GetToUniqueStopID()]; !has_already_marked_stop {
									stops_marked_for_next_round[transfer_stop.GetToUniqueStopID()] = RaptorMarkedStop[ID]{
										ID:     transfer_stop.GetToUniqueStopID(),
										Source: RaptorMarkedStopSourceTransfer,
									}
								}
								/* for each transferrable station we'll also add an earliest arrival segment which is the current arrival time + the minimum transfer time (if the arrival is earlier than the previously recorded one) */
								arrival_time_at_transfer_stop := following_stop_time.GetArrivalTimeInSeconds() + int64(transfer_stop.GetMinimumTransferTimeInSeconds())

								existing_transfer_segment, has_existing_transfer_segment := earliest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()]
								if !has_existing_transfer_segment || existing_transfer_segment.ArrivalTimeInSeconds > arrival_time_at_transfer_stop {
									/* copy current segment spans from the original arrival station + add a new one for the transfer itself */
									updated_spans := make([]RoundSegmentSpan[ID], len(existing_segment.Spans)+1)
									copy(updated_spans, existing_segment.Spans)
									updated_spans[len(updated_spans)-1] = RoundSegmentSpan[ID]{
										FromUniqueStopID:                       following_stop_time.GetUniqueStopID(),
										ToUniqueStopID:                         transfer_stop.GetToUniqueStopID(),
										ViaTrip:                                nil,
										DepartureTimeInSecondsFromUniqueStopID: following_stop_time.GetArrivalTimeInSeconds(),
										ArrivalTimeInSecondsToUniqueStopID:     arrival_time_at_transfer_stop,
									}
									earliest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()] = RoundSegment[ID]{
										UniqueStopID:         transfer_stop.GetToUniqueStopID(),
										ArrivalTimeInSeconds: arrival_time_at_transfer_stop,
										Spans:                updated_spans,
									}
								}
							}
						}
					}
					/* next we can mark this stop to check in the next round AND add any potential transfers from this stop to mark */
					stops_marked_for_next_round[following_stop_time.GetUniqueStopID()] = RaptorMarkedStop[ID]{
						ID:     following_stop_time.GetUniqueStopID(),
						Source: RaptorMarkedStopSourceArrival,
					}

					/* lastly we can check if this stop is actually one of our destination stops - in which case the segment is corresponding to a complete journe7 */
					if _, is_destination_stop := prepared_input.ToStopsByUniqueStopId[following_stop_time.GetUniqueStopID()]; is_destination_stop {
						segment := earliest_arrival_time_segments_by_unique_stop_id[following_stop_time.GetUniqueStopID()]
						segment_fingerprint := segment.GetFingerPrint()
						if _, has_same_trip := potential_journey_fingerprints[segment_fingerprint]; !has_same_trip && len(segment.Spans) > 0 && segment.Spans[0].ViaTrip != nil && segment.Spans[len(segment.Spans)-1].ViaTrip != nil {
							/* if the spans are 0 it means we were already at our stop in the first place */
							segment_spans := make([]RoundSegmentSpan[ID], len(segment.Spans))
							copy(segment_spans, segment.Spans)
							first_segment_span := segment_spans[0]
							last_segment_span := segment_spans[len(segment_spans)-1]
							journey := Journey[ID]{
								FromUniqueStopID:       first_segment_span.FromUniqueStopID,
								ToUniqueStopID:         last_segment_span.ToUniqueStopID,
								DepartureTimeInSeconds: first_segment_span.DepartureTimeInSecondsFromUniqueStopID,
								ArrivalTimeInSeconds:   last_segment_span.ArrivalTimeInSecondsToUniqueStopID,
								Legs:                   segment_spans,
							}

							potential_journeys_found = append(potential_journeys_found, journey)
							potential_journey_fingerprints[segment_fingerprint] = true

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
) []Journey[ID] {
	/* !! stop times input should be in reverse */
	prepared_input := PrepareRaptorInput(input)

	/* below is the start of the raptor based algorithm */
	/* this map contains the latest possible arrival time at each stop across rounds (nearest to the arrive by time) - keeping track of all the segments */
	latest_arrival_time_segments_by_unique_stop_id := map[ID]RoundSegment[ID]{}
	/* this is the result slice which contains all the potential journeys (meaning segments which reach the end destination) */
	potential_journeys_found := []Journey[ID]{}
	potential_journey_fingerprints := map[string]bool{}

	/* to start we need to mark which stops we are going to check during the current round - at the start this will only be the destinations stops */
	/* this will be replaced between rounds because we will be checking the next set of transferred to stops */
	stops_marked_for_round := make(map[ID]RaptorMarkedStop[ID], len(input.ToStops))
	for _, stop := range input.ToStops {
		stops_marked_for_round[stop.GetUniqueID()] = RaptorMarkedStop[ID]{
			ID:     stop.GetUniqueID(),
			Source: RaptorMarkedStopSourceArrival,
		}
	}

	/* we will also initialize the initial segments for the to_stops -> essentially saying we have not been able to arrive yet */
	for _, to_stop := range input.ToStops {
		latest_arrival_time_segments_by_unique_stop_id[to_stop.GetUniqueID()] = RoundSegment[ID]{
			UniqueStopID:         to_stop.GetUniqueID(),
			ArrivalTimeInSeconds: input.TimeInSeconds,
			/* no spans yet since we need to calculate the arrival route */
			Spans: []RoundSegmentSpan[ID]{},
		}
	}

	/* now we can start the rounds up until N transfers */
	trips_scanned_from_sequence := map[ID]int{}
	for range input.MaximumTransfers {
		/* this will be the set of next stops to check for the next round */
		stops_marked_for_next_round := map[ID]RaptorMarkedStop[ID]{}
		/* in each round we will check all marked stops for the trips we could take - we will do this by going through the stop times in  reverse */
		for _, marked_stop := range stops_marked_for_round {
			/* this should always exist because any marked stop should have been added to the segment list */
			current_segment_for_stop := latest_arrival_time_segments_by_unique_stop_id[marked_stop.ID]
			stop_times_for_marked_stop := prepared_input.StopTimesByUniqueStopId[marked_stop.ID]
			/*
				we will go through the stop times and find the latest arrivals which are still before my expected can make based on our current earliest arrival time at the  "marked_stop"
				in the arrive by implementation we will iterate in reverse
			*/
			stop_times_for_marked_stop_it := NewSliceIterator(stop_times_for_marked_stop, true)
			for stop_times_for_marked_stop_it.HasNext() {
				stop_time_for_marked_stop := prepared_input.Input.StopTimes[stop_times_for_marked_stop_it.Next()]
				trip_already_scanned_from_sequence, has_already_scanned_trip_from_sequence := trips_scanned_from_sequence[stop_time_for_marked_stop.GetUniqueTripID()]
				/* we don't want to scan the preceeding stops if they were already scanned before -> unless this stop sequence is after the already scanned sequence in which case we are missing a few */
				if stop_time_for_marked_stop.GetArrivalTimeInSeconds() > current_segment_for_stop.ArrivalTimeInSeconds ||
					has_already_scanned_trip_from_sequence && stop_time_for_marked_stop.GetStopSequence() <= trip_already_scanned_from_sequence {
					/* if the arrival time of this stop time happens after the current segment arrival time then we are too late */
					continue
				}

				/* mark trip as scanned from sequence */
				trips_scanned_from_sequence[stop_time_for_marked_stop.GetUniqueTripID()] = stop_time_for_marked_stop.GetStopSequence()

				/*
				 * if we CAN make it we will want to look up the stop times before the current one in the trip.
				 * we're essentially just going down the line in reverse and storing each stop time if the arrival time is later than the currently stored one
				 * (meaning I could get to this stop later than initially expected)
				 */
				/* to get these we want to reverse the stop sequence and skip one to exclude my current stop which I already checked */
				var stop_times_for_unique_trip_id_after_current_stop_it *SliceIterator[int]
				stop_times_for_unique_trip_id_it := NewSliceIterator(prepared_input.StopTimesByUniqueTripServiceId[stop_time_for_marked_stop.GetUniqueTripServiceID()], true)
				stop_times_last_sequence := prepared_input.Input.StopTimes[stop_times_for_unique_trip_id_it.First()].GetStopSequence()
				stop_times_start_offset := stop_times_last_sequence - stop_time_for_marked_stop.GetStopSequence() + 1
				stop_times_end_offset := stop_times_last_sequence - trip_already_scanned_from_sequence
				if !has_already_scanned_trip_from_sequence {
					stop_times_for_unique_trip_id_after_current_stop_it = stop_times_for_unique_trip_id_it.SliceIterator(stop_times_start_offset, stop_times_for_unique_trip_id_it.Length())
				} else {
					stop_times_for_unique_trip_id_after_current_stop_it = stop_times_for_unique_trip_id_it.SliceIterator(stop_times_start_offset, stop_times_end_offset)
				}

				/* the stop times are expected to be in order of sequence descending */
			preceeding_stop_times_loop:
				for stop_times_for_unique_trip_id_after_current_stop_it.HasNext() {
					preceeding_stop_time := prepared_input.Input.StopTimes[stop_times_for_unique_trip_id_after_current_stop_it.Next()]
					existing_segment, has_existing_segment := latest_arrival_time_segments_by_unique_stop_id[preceeding_stop_time.GetUniqueStopID()]
					is_improvement_to_existing_arrival_time := !has_existing_segment || preceeding_stop_time.GetArrivalTimeInSeconds() > existing_segment.ArrivalTimeInSeconds
					/* if this stop was not arrived at yet OR if this arrival is after the recorded arrival */
					if is_improvement_to_existing_arrival_time {
						/* we'll want to update the segment spans of the current marked stop NOT the preceeding stop since we don't know yet how we can arrive at the preceeding */
						/* however we do now now how we could arrive at the current marked stop which is through this stop time */
						updated_spans := append([]RoundSegmentSpan[ID]{
							{
								FromUniqueStopID: preceeding_stop_time.GetUniqueStopID(),
								ToUniqueStopID:   stop_time_for_marked_stop.GetUniqueStopID(),
								ViaTrip: &ViaTrip[ID]{
									UniqueTripID:           preceeding_stop_time.GetUniqueTripID(),
									UniqueTripServiceID:    preceeding_stop_time.GetUniqueTripServiceID(),
									FromStopSequenceInTrip: preceeding_stop_time.GetStopSequence(),
									ToStopSequenceInTrip:   stop_time_for_marked_stop.GetStopSequence(),
								},
								DepartureTimeInSecondsFromUniqueStopID: preceeding_stop_time.GetDepartureTimeInSeconds(),
								ArrivalTimeInSecondsToUniqueStopID:     stop_time_for_marked_stop.GetArrivalTimeInSeconds(),
							},
						}, current_segment_for_stop.Spans...)
						latest_arrival_time_segments_by_unique_stop_id[preceeding_stop_time.GetUniqueStopID()] = RoundSegment[ID]{
							UniqueStopID:         preceeding_stop_time.GetUniqueStopID(),
							ArrivalTimeInSeconds: preceeding_stop_time.GetArrivalTimeInSeconds(),
							Spans:                updated_spans,
						}
						/* update existing segment in place for later */
						existing_segment = latest_arrival_time_segments_by_unique_stop_id[preceeding_stop_time.GetUniqueStopID()]

						/* only allow looking for transfers again if transfer hopping is allowed or the currently marked stop was arrived at by a trip not by a transfer */
						if input.AllowTransferHopping || marked_stop.Source == RaptorMarkedStopSourceArrival {
							potential_transfers_for_stop := prepared_input.TransfersByUniqueStopId[preceeding_stop_time.GetUniqueStopID()]
							for _, transfer_stop_index := range potential_transfers_for_stop {
								transfer_stop := prepared_input.Input.Transfers[transfer_stop_index]
								/* we don't want to override a direct arrival mark */
								if _, has_already_marked_stop := stops_marked_for_next_round[transfer_stop.GetToUniqueStopID()]; !has_already_marked_stop {
									stops_marked_for_next_round[transfer_stop.GetToUniqueStopID()] = RaptorMarkedStop[ID]{
										ID:     transfer_stop.GetToUniqueStopID(),
										Source: RaptorMarkedStopSourceTransfer,
									}
								}
								/* for each transferrable station we'll also add a latest arrival segment which is the current arrival time - the minimum transfer time (if the arrival is later than the previously recorded one) */
								departure_time_from_transfer_stop := preceeding_stop_time.GetArrivalTimeInSeconds() - int64(transfer_stop.GetMinimumTransferTimeInSeconds())
								existing_transfer_segment, has_existing_transfer_segment := latest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()]
								if !has_existing_transfer_segment || departure_time_from_transfer_stop > existing_transfer_segment.ArrivalTimeInSeconds {
									/* copy current segment spans from the original arrival station + add a new one for the transfer itself */
									updated_spans := append([]RoundSegmentSpan[ID]{
										{
											FromUniqueStopID:                       transfer_stop.GetToUniqueStopID(),
											ToUniqueStopID:                         preceeding_stop_time.GetUniqueStopID(),
											ViaTrip:                                nil,
											DepartureTimeInSecondsFromUniqueStopID: departure_time_from_transfer_stop,
											ArrivalTimeInSecondsToUniqueStopID:     preceeding_stop_time.GetArrivalTimeInSeconds(),
										},
									}, existing_segment.Spans...)
									latest_arrival_time_segments_by_unique_stop_id[transfer_stop.GetToUniqueStopID()] = RoundSegment[ID]{
										UniqueStopID:         transfer_stop.GetToUniqueStopID(),
										ArrivalTimeInSeconds: departure_time_from_transfer_stop,
										Spans:                updated_spans,
									}
								}
							}
						}
					}
					/* next we can mark this stop to check in the next round AND add any potential transfers from this stop to mark */
					stops_marked_for_next_round[preceeding_stop_time.GetUniqueStopID()] = RaptorMarkedStop[ID]{
						ID:     preceeding_stop_time.GetUniqueStopID(),
						Source: RaptorMarkedStopSourceArrival,
					}

					/* lastly we can check if this stop is actually one of our origin stops - in which case the segment is corresponding to a complete journe7 */
					if _, is_origin_stop := prepared_input.FromStopsByUniqueStopId[preceeding_stop_time.GetUniqueStopID()]; is_origin_stop {
						segment := latest_arrival_time_segments_by_unique_stop_id[preceeding_stop_time.GetUniqueStopID()]
						segment_fingerprint := segment.GetFingerPrint()
						if _, has_same_trip := potential_journey_fingerprints[segment_fingerprint]; !has_same_trip && len(segment.Spans) > 0 && segment.Spans[0].ViaTrip != nil && segment.Spans[len(segment.Spans)-1].ViaTrip != nil {
							/* if the spans are 0 it means we were already at our stop in the first place */
							segment_spans := make([]RoundSegmentSpan[ID], len(segment.Spans))
							copy(segment_spans, segment.Spans)
							first_segment_span := segment_spans[0]
							last_segment_span := segment_spans[len(segment_spans)-1]
							journey := Journey[ID]{
								FromUniqueStopID:       first_segment_span.FromUniqueStopID,
								ToUniqueStopID:         last_segment_span.ToUniqueStopID,
								DepartureTimeInSeconds: first_segment_span.DepartureTimeInSecondsFromUniqueStopID,
								ArrivalTimeInSeconds:   last_segment_span.ArrivalTimeInSecondsToUniqueStopID,
								Legs:                   segment_spans,
							}

							potential_journeys_found = append(potential_journeys_found, journey)
							potential_journey_fingerprints[segment_fingerprint] = true

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
) []Journey[ID] {
	if input.Mode == RaptorModeDepartAt {
		return SimpleRaptorDepartAt(input)
	}
	return SimpleRaptorArriveBy(input)
}
