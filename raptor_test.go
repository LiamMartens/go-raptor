package go_raptor

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/patrickbr/gtfsparser"
	"github.com/patrickbr/gtfsparser/gtfs"
)

func FormatSecondsSinceMidnight(secs int) string {
	hours := secs / 3600
	minutes := (secs % 3600) / 60
	seconds := secs % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func TestForwardRaptor(t *testing.T) {
	feed := gtfsparser.NewFeed()

	feed.Parse("./gtfs_subway.zip")

	all_stops_by_id := map[string]*gtfs.Stop{}
	parent_child_stations_by_id := map[string][]string{}

	from_stops := []GtfsStopStruct[string]{}
	to_stops := []GtfsStopStruct[string]{}
	transfers := []GtfsTransferStruct[string]{}
	stop_times := []GtfsStopTimeStruct[string]{}

	for _, stop := range feed.Stops {
		all_stops_by_id[stop.Id] = stop
		if stop.Parent_station != nil {
			if _, has_list := parent_child_stations_by_id[stop.Parent_station.Id]; !has_list {
				parent_child_stations_by_id[stop.Parent_station.Id] = []string{}
			}
			parent_child_stations_by_id[stop.Parent_station.Id] = append(parent_child_stations_by_id[stop.Parent_station.Id], stop.Id)
		}

		/* from High St */
		if strings.HasPrefix(stop.Id, "A40") {
			from_stops = append(from_stops, GtfsStopStruct[string]{UniqueID: stop.Id})
		}
		/* 36st Astoria */
		if strings.HasPrefix(stop.Id, "A44") {
			to_stops = append(to_stops, GtfsStopStruct[string]{UniqueID: stop.Id})
		}
	}
	for from_to, transfer := range feed.Transfers {
		/*check if this is a parent station instead of an actual station */
		from_child_stations, from_has_child_stations := parent_child_stations_by_id[from_to.From_stop.Id]
		to_child_stations, to_has_child_stations := parent_child_stations_by_id[from_to.To_stop.Id]

		transfers_possible_from := from_child_stations
		transfers_possible_to := to_child_stations
		if !from_has_child_stations {
			transfers_possible_from = []string{from_to.From_stop.Id}
		}
		if !to_has_child_stations {
			transfers_possible_to = []string{from_to.To_stop.Id}
		}

		for _, from_stop := range transfers_possible_from {
			for _, to_stop := range transfers_possible_to {
				if from_stop != to_stop {
					transfers = append(transfers, GtfsTransferStruct[string]{
						FromUniqueStopID:             from_stop,
						ToUniqueStopID:               to_stop,
						MinimumTransferTimeInSeconds: transfer.Min_transfer_time,
					})
				}
			}
		}
	}
	for _, trip := range feed.Trips {
		if trip.Service.Id() == "Weekday" {
			for _, stop_time := range trip.StopTimes {
				stop_times = append(stop_times, GtfsStopTimeStruct[string]{
					UniqueStopID:           stop_time.Stop().Id,
					UniqueTripID:           trip.Id,
					ArrivalTimeInSeconds:   stop_time.Arrival_time().SecondsSinceMidnight(),
					DepartureTimeInSeconds: stop_time.Departure_time().SecondsSinceMidnight(),
					StopSequence:           stop_time.Sequence(),
				})
			}
		}
	}

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops:        from_stops,
			ToStops:          to_stops,
			Transfers:        transfers,
			StopTimes:        stop_times,
			Mode:             RaptorModeDepartAt,
			DateOfService:    "20250814",
			TimeInSeconds:    9 * 3600,
			MaximumTransfers: 4,
		},
	)

	fmt.Printf("found %d journeys\n", len(journeys))
	sort.Slice(journeys, func(i, j int) bool {
		/* return true if I < J */
		return journeys[i].ArrivalTimeInSeconds < journeys[j].ArrivalTimeInSeconds
	})

	fmt.Printf("the first available journey departs from %s at %s arrives at %s by %s\n", journeys[0].FromUniqueStopID, FormatSecondsSinceMidnight(journeys[0].DepartureTimeInSeconds), journeys[0].ToUniqueStopID, FormatSecondsSinceMidnight(journeys[0].ArrivalTimeInSeconds))
	for i, leg := range journeys[0].Legs {
		if leg.ViaTrip != nil {
			fmt.Printf("the %dnth leg takes the following trip %s at %s from stop %s to stop %s\n", i, leg.ViaTrip.UniqueTripID, FormatSecondsSinceMidnight(leg.DepartureTimeInSecondsFromUniqueStopID), leg.FromUniqueStopID, leg.ToUniqueStopID)
		} else {
			fmt.Printf("the %dnth transfers from %s to %s by walking\n", i, leg.FromUniqueStopID, leg.ToUniqueStopID)
		}
	}
}

func TestForwardRaptorLIRR(t *testing.T) {
	feed := gtfsparser.NewFeed()

	feed.Parse("./gtfslirr.zip")

	for _, service := range feed.Services {
		fmt.Printf("%#v\n", service.RawDaymap() == 0)
	}

	all_stops_by_id := map[string]*gtfs.Stop{}
	parent_child_stations_by_id := map[string][]string{}

	from_stops := []GtfsStopStruct[string]{}
	to_stops := []GtfsStopStruct[string]{}
	transfers := []GtfsTransferStruct[string]{}
	stop_times := []GtfsStopTimeStruct[string]{}

	for _, stop := range feed.Stops {
		all_stops_by_id[stop.Id] = stop
		if stop.Parent_station != nil {
			if _, has_list := parent_child_stations_by_id[stop.Parent_station.Id]; !has_list {
				parent_child_stations_by_id[stop.Parent_station.Id] = []string{}
			}
			parent_child_stations_by_id[stop.Parent_station.Id] = append(parent_child_stations_by_id[stop.Parent_station.Id], stop.Id)
		}

		/* from Nostrand Avenue */
		if strings.HasPrefix(stop.Id, "148") {
			from_stops = append(from_stops, GtfsStopStruct[string]{UniqueID: stop.Id})
		}
		/* valley stream */
		if strings.HasPrefix(stop.Id, "211") {
			to_stops = append(to_stops, GtfsStopStruct[string]{UniqueID: stop.Id})
		}
	}
	for from_to, transfer := range feed.Transfers {
		/*check if this is a parent station instead of an actual station */
		from_child_stations, from_has_child_stations := parent_child_stations_by_id[from_to.From_stop.Id]
		to_child_stations, to_has_child_stations := parent_child_stations_by_id[from_to.To_stop.Id]

		transfers_possible_from := from_child_stations
		transfers_possible_to := to_child_stations
		if !from_has_child_stations {
			transfers_possible_from = []string{from_to.From_stop.Id}
		}
		if !to_has_child_stations {
			transfers_possible_to = []string{from_to.To_stop.Id}
		}

		for _, from_stop := range transfers_possible_from {
			for _, to_stop := range transfers_possible_to {
				if from_stop != to_stop {
					transfers = append(transfers, GtfsTransferStruct[string]{
						FromUniqueStopID:             from_stop,
						ToUniqueStopID:               to_stop,
						MinimumTransferTimeInSeconds: transfer.Min_transfer_time,
					})
				}
			}
		}
	}
	for _, trip := range feed.Trips {
		for _, stop_time := range trip.StopTimes {
			stop_times = append(stop_times, GtfsStopTimeStruct[string]{
				UniqueStopID:           stop_time.Stop().Id,
				UniqueTripID:           trip.Id,
				ArrivalTimeInSeconds:   stop_time.Arrival_time().SecondsSinceMidnight(),
				DepartureTimeInSeconds: stop_time.Departure_time().SecondsSinceMidnight(),
				StopSequence:           stop_time.Sequence(),
			})
		}
	}

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops:        from_stops,
			ToStops:          to_stops,
			Transfers:        transfers,
			StopTimes:        stop_times,
			Mode:             RaptorModeDepartAt,
			DateOfService:    "20250814",
			TimeInSeconds:    9 * 3600,
			MaximumTransfers: 4,
		},
	)

	fmt.Printf("found %d journeys\n", len(journeys))
	sort.Slice(journeys, func(i, j int) bool {
		/* return true if I < J */
		return journeys[i].ArrivalTimeInSeconds < journeys[j].ArrivalTimeInSeconds
	})

	fmt.Printf("the first available journey departs from %s at %s arrives at %s by %s\n", journeys[0].FromUniqueStopID, FormatSecondsSinceMidnight(journeys[0].DepartureTimeInSeconds), journeys[0].ToUniqueStopID, FormatSecondsSinceMidnight(journeys[0].ArrivalTimeInSeconds))
	for i, leg := range journeys[0].Legs {
		if leg.ViaTrip != nil {
			fmt.Printf("the %dnth leg takes the following trip %s at %s from stop %s to stop %s\n", i, leg.ViaTrip.UniqueTripID, FormatSecondsSinceMidnight(leg.DepartureTimeInSecondsFromUniqueStopID), leg.FromUniqueStopID, leg.ToUniqueStopID)
		} else {
			fmt.Printf("the %dnth transfers from %s to %s by walking\n", i, leg.FromUniqueStopID, leg.ToUniqueStopID)
		}
	}
}

func TestReverseRaptor(t *testing.T) {
	feed := gtfsparser.NewFeed()

	feed.Parse("./gtfs_subway.zip")

	all_stops_by_id := map[string]*gtfs.Stop{}
	parent_child_stations_by_id := map[string][]string{}

	from_stops := []GtfsStopStruct[string]{}
	to_stops := []GtfsStopStruct[string]{}
	transfers := []GtfsTransferStruct[string]{}
	stop_times := []GtfsStopTimeStruct[string]{}

	for _, stop := range feed.Stops {
		all_stops_by_id[stop.Id] = stop
		if stop.Parent_station != nil {
			if _, has_list := parent_child_stations_by_id[stop.Parent_station.Id]; !has_list {
				parent_child_stations_by_id[stop.Parent_station.Id] = []string{}
			}
			parent_child_stations_by_id[stop.Parent_station.Id] = append(parent_child_stations_by_id[stop.Parent_station.Id], stop.Id)
		}

		/* from High St */
		if strings.HasPrefix(stop.Id, "A40") {
			from_stops = append(from_stops, GtfsStopStruct[string]{UniqueID: stop.Id})
		}
		/* 36st Astoria */
		if strings.HasPrefix(stop.Id, "R36") {
			// if strings.HasPrefix(stop.Id, "A41") {
			to_stops = append(to_stops, GtfsStopStruct[string]{UniqueID: stop.Id})
		}
	}
	for from_to, transfer := range feed.Transfers {
		/*check if this is a parent station instead of an actual station */
		from_child_stations, from_has_child_stations := parent_child_stations_by_id[from_to.From_stop.Id]
		to_child_stations, to_has_child_stations := parent_child_stations_by_id[from_to.To_stop.Id]

		transfers_possible_from := from_child_stations
		transfers_possible_to := to_child_stations
		if !from_has_child_stations {
			transfers_possible_from = []string{from_to.From_stop.Id}
		}
		if !to_has_child_stations {
			transfers_possible_to = []string{from_to.To_stop.Id}
		}

		for _, from_stop := range transfers_possible_from {
			for _, to_stop := range transfers_possible_to {
				if from_stop != to_stop {
					transfers = append(transfers, GtfsTransferStruct[string]{
						FromUniqueStopID:             from_stop,
						ToUniqueStopID:               to_stop,
						MinimumTransferTimeInSeconds: transfer.Min_transfer_time,
					})
				}
			}
		}
	}
	for _, trip := range feed.Trips {
		if trip.Service.Id() == "Weekday" {
			for _, stop_time := range trip.StopTimes {
				stop_times = append(stop_times, GtfsStopTimeStruct[string]{
					UniqueStopID:           stop_time.Stop().Id,
					UniqueTripID:           trip.Id,
					ArrivalTimeInSeconds:   stop_time.Arrival_time().SecondsSinceMidnight(),
					DepartureTimeInSeconds: stop_time.Departure_time().SecondsSinceMidnight(),
					StopSequence:           stop_time.Sequence(),
				})
			}
		}
	}

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops:        from_stops,
			ToStops:          to_stops,
			Transfers:        transfers,
			StopTimes:        stop_times,
			Mode:             RaptorModeArriveBy,
			DateOfService:    "20250814",
			TimeInSeconds:    9 * 3600,
			MaximumTransfers: 4,
		},
	)

	fmt.Printf("found %d journeys\n", len(journeys))
	sort.Slice(journeys, func(i, j int) bool {
		return journeys[i].ArrivalTimeInSeconds > journeys[j].ArrivalTimeInSeconds
	})
	fmt.Printf("the last possible journey departs from %s at %s arrives at %s by %s\n", journeys[0].FromUniqueStopID, FormatSecondsSinceMidnight(journeys[0].DepartureTimeInSeconds), journeys[0].ToUniqueStopID, FormatSecondsSinceMidnight(journeys[0].ArrivalTimeInSeconds))
	for i, leg := range journeys[0].Legs {
		if leg.ViaTrip != nil {
			fmt.Printf("the %dnth leg takes the following trip %s at %s from stop %s to stop %s\n", i, leg.ViaTrip.UniqueTripID, FormatSecondsSinceMidnight(leg.DepartureTimeInSecondsFromUniqueStopID), leg.FromUniqueStopID, leg.ToUniqueStopID)
		} else {
			fmt.Printf("the %dnth transfers from %s to %s by walking\n", i, leg.FromUniqueStopID, leg.ToUniqueStopID)
		}
	}
}
