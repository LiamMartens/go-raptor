package go_raptor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func FormatSecondsSinceMidnight(secs int64) string {
	hours := secs / 3600
	minutes := (secs % 3600) / 60
	seconds := secs % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func TestSimpleForwardRaptor(t *testing.T) {
	var epoch_20250822_120000_edt int64 = 1755878400
	var epoch_20250823_120000_edt int64 = 1755964800
	var epoch_20250824_120000_edt int64 = 1756051200

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops: []GtfsStopStruct[string]{
				{UniqueID: "High St"},
			},
			ToStops: []GtfsStopStruct[string]{
				{UniqueID: "Franklin Av"},
			},
			Transfers: []GtfsTransferStruct[string]{},
			StopTimes: []GtfsStopTimeStruct[string]{
				{UniqueStopID: "High St", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250822_120000_edt - 10, DepartureTimeInSeconds: epoch_20250822_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 120, DepartureTimeInSeconds: epoch_20250822_120000_edt + 130},

				{UniqueStopID: "High St", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250823_120000_edt - 10, DepartureTimeInSeconds: epoch_20250823_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 120, DepartureTimeInSeconds: epoch_20250823_120000_edt + 130},

				{UniqueStopID: "High St", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250824_120000_edt - 10, DepartureTimeInSeconds: epoch_20250824_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 120, DepartureTimeInSeconds: epoch_20250824_120000_edt + 130},
			},
			Mode: RaptorModeDepartAt,
			/* 2025/08/23 12:00:00PM EDT */
			TimeInSeconds:        epoch_20250823_120000_edt,
			MaximumTransfers:     4,
			AllowTransferHopping: false,
		},
	)

	if len(journeys) == 0 {
		t.Fatalf(`did not find any journeys for stop times`)
	}

	if journeys[0].ArrivalTimeInSeconds != epoch_20250823_120000_edt+120 {
		t.Fatalf(`expected raptor to find arrival time %v but got %v`, epoch_20250823_120000_edt+120, journeys[0].ArrivalTimeInSeconds)
	}
}

func TestSimpleReverseRaptor(t *testing.T) {
	var epoch_20250822_120000_edt int64 = 1755878400
	var epoch_20250823_120000_edt int64 = 1755964800
	var epoch_20250824_120000_edt int64 = 1756051200

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops: []GtfsStopStruct[string]{
				{UniqueID: "High St"},
			},
			ToStops: []GtfsStopStruct[string]{
				{UniqueID: "Franklin Av"},
			},
			Transfers: []GtfsTransferStruct[string]{},
			StopTimes: []GtfsStopTimeStruct[string]{
				{UniqueStopID: "High St", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250822_120000_edt - 10, DepartureTimeInSeconds: epoch_20250822_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 120, DepartureTimeInSeconds: epoch_20250822_120000_edt + 130},

				{UniqueStopID: "High St", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250823_120000_edt - 10, DepartureTimeInSeconds: epoch_20250823_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 120, DepartureTimeInSeconds: epoch_20250823_120000_edt + 130},

				{UniqueStopID: "High St", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250824_120000_edt - 10, DepartureTimeInSeconds: epoch_20250824_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 120, DepartureTimeInSeconds: epoch_20250824_120000_edt + 130},
			},
			Mode: RaptorModeArriveBy,
			/* 2025/08/23 12:00:00PM EDT */
			TimeInSeconds:        epoch_20250823_120000_edt + 120,
			MaximumTransfers:     4,
			AllowTransferHopping: false,
		},
	)

	if len(journeys) == 0 {
		t.Fatalf(`did not find any journeys for stop times`)
	}

	if journeys[0].DepartureTimeInSeconds != epoch_20250823_120000_edt+10 {
		t.Fatalf(`expected raptor to find departure time %v but got %v`, epoch_20250823_120000_edt+10, journeys[0].DepartureTimeInSeconds)
	}
}

func TestSimpleForwardRaptor2(t *testing.T) {
	now := time.Now()
	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops: []GtfsStopStruct[string]{
				{UniqueID: "Franklin Ave"},
			},
			ToStops: []GtfsStopStruct[string]{
				{UniqueID: "Jay Street"},
			},
			Transfers: []GtfsTransferStruct[string]{},
			StopTimes: []GtfsStopTimeStruct[string]{
				{
					UniqueStopID:           "Franklin Ave",
					UniqueTripID:           "C_NORTH",
					UniqueTripServiceID:    "C_NORTH",
					StopSequence:           5,
					ArrivalTimeInSeconds:   now.Add(10 * time.Second).Unix(),
					DepartureTimeInSeconds: now.Add(15 * time.Second).Unix(),
				},
				{
					UniqueStopID:           "Jay Street",
					UniqueTripID:           "C_NORTH",
					UniqueTripServiceID:    "C_NORTH",
					StopSequence:           6,
					ArrivalTimeInSeconds:   now.Add(60 * time.Second).Unix(),
					DepartureTimeInSeconds: now.Add(65 * time.Second).Unix(),
				},

				{
					UniqueStopID:           "Franklin Ave",
					UniqueTripID:           "C_SOUTH",
					UniqueTripServiceID:    "C_SOUTH",
					StopSequence:           5,
					ArrivalTimeInSeconds:   now.Add(15 * time.Second).Unix(),
					DepartureTimeInSeconds: now.Add(20 * time.Second).Unix(),
				},
				{
					UniqueStopID:           "Nostrand",
					UniqueTripID:           "C_SOUTH",
					UniqueTripServiceID:    "C_SOUTH",
					StopSequence:           6,
					ArrivalTimeInSeconds:   now.Add(30 * time.Second).Unix(),
					DepartureTimeInSeconds: now.Add(35 * time.Second).Unix(),
				},
				{
					UniqueStopID:           "Nostrand",
					UniqueTripID:           "A_NORTH",
					UniqueTripServiceID:    "A_NORTH",
					StopSequence:           6,
					ArrivalTimeInSeconds:   now.Add(40 * time.Second).Unix(),
					DepartureTimeInSeconds: now.Add(45 * time.Second).Unix(),
				},
				{
					UniqueStopID:           "Jay Street",
					UniqueTripID:           "A_NORTH",
					UniqueTripServiceID:    "A_NORTH",
					StopSequence:           7,
					ArrivalTimeInSeconds:   now.Add(55 * time.Second).Unix(),
					DepartureTimeInSeconds: now.Add(60 * time.Second).Unix(),
				},
			},
			Mode:                 RaptorModeDepartAt,
			TimeInSeconds:        now.Unix(),
			MaximumTransfers:     4,
			AllowTransferHopping: false,
		},
	)

	assert.Len(t, journeys, 2, "should return both journey options")
}

func TestSimpleForwardRaptor_MultiTrip(t *testing.T) {
	var epoch_20250822_120000_edt int64 = 1755878400
	var epoch_20250823_120000_edt int64 = 1755964800
	var epoch_20250824_120000_edt int64 = 1756051200

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops: []GtfsStopStruct[string]{
				{UniqueID: "High St"},
			},
			ToStops: []GtfsStopStruct[string]{
				{UniqueID: "Franklin Av"},
			},
			Transfers: []GtfsTransferStruct[string]{},
			StopTimes: []GtfsStopTimeStruct[string]{
				{UniqueStopID: "High St", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250822_120000_edt - 10, DepartureTimeInSeconds: epoch_20250822_120000_edt + 10},
				{UniqueStopID: "Hoyt St", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 120, DepartureTimeInSeconds: epoch_20250822_120000_edt + 130},
				{UniqueStopID: "Hoyt St", UniqueTripID: "C_20250822", UniqueTripServiceID: "C_20250822", StopSequence: 8, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 125, DepartureTimeInSeconds: epoch_20250822_120000_edt + 135},
				{UniqueStopID: "Franklin Av", UniqueTripID: "C_20250822", UniqueTripServiceID: "C_20250822", StopSequence: 9, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 200, DepartureTimeInSeconds: epoch_20250822_120000_edt + 210},

				{UniqueStopID: "High St", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250823_120000_edt - 10, DepartureTimeInSeconds: epoch_20250823_120000_edt + 10},
				{UniqueStopID: "Hoyt St", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 120, DepartureTimeInSeconds: epoch_20250823_120000_edt + 130},
				{UniqueStopID: "Hoyt St", UniqueTripID: "C_20250823", UniqueTripServiceID: "C_20250823", StopSequence: 8, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 125, DepartureTimeInSeconds: epoch_20250823_120000_edt + 135},
				{UniqueStopID: "Franklin Av", UniqueTripID: "C_20250823", UniqueTripServiceID: "C_20250823", StopSequence: 9, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 200, DepartureTimeInSeconds: epoch_20250823_120000_edt + 210},

				{UniqueStopID: "High St", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250824_120000_edt - 10, DepartureTimeInSeconds: epoch_20250824_120000_edt + 10},
				{UniqueStopID: "Hoyt St", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 120, DepartureTimeInSeconds: epoch_20250824_120000_edt + 130},
				{UniqueStopID: "Hoyt St", UniqueTripID: "C_20250824", UniqueTripServiceID: "C_20250824", StopSequence: 8, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 125, DepartureTimeInSeconds: epoch_20250824_120000_edt + 135},
				{UniqueStopID: "Franklin Av", UniqueTripID: "C_20250824", UniqueTripServiceID: "C_20250824", StopSequence: 9, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 200, DepartureTimeInSeconds: epoch_20250824_120000_edt + 210},
			},
			Mode: RaptorModeDepartAt,
			/* 2025/08/23 12:00:00PM EDT */
			TimeInSeconds:        epoch_20250823_120000_edt,
			MaximumTransfers:     4,
			AllowTransferHopping: false,
		},
	)

	if len(journeys) == 0 {
		t.Fatalf(`did not find any journeys for stop times`)
	}

	if journeys[0].ArrivalTimeInSeconds != epoch_20250823_120000_edt+200 {
		t.Fatalf(`expected raptor to find arrival time %v but got %v`, epoch_20250823_120000_edt+200, journeys[0].ArrivalTimeInSeconds)
	}
}

func TestSimpleForwardRaptor_ManualTransfer(t *testing.T) {
	var epoch_20250822_120000_edt int64 = 1755878400
	var epoch_20250823_120000_edt int64 = 1755964800
	var epoch_20250824_120000_edt int64 = 1756051200

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops: []GtfsStopStruct[string]{
				{UniqueID: "High St"},
			},
			ToStops: []GtfsStopStruct[string]{
				{UniqueID: "Franklin Av"},
			},
			Transfers: []GtfsTransferStruct[string]{
				{FromUniqueStopID: "Jay St", ToUniqueStopID: "Hoyt St", MinimumTransferTimeInSeconds: 0},
			},
			StopTimes: []GtfsStopTimeStruct[string]{
				{UniqueStopID: "High St", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250822_120000_edt - 10, DepartureTimeInSeconds: epoch_20250822_120000_edt + 10},
				{UniqueStopID: "Jay St", UniqueTripID: "A_20250822", UniqueTripServiceID: "A_20250822", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 120, DepartureTimeInSeconds: epoch_20250822_120000_edt + 130},
				{UniqueStopID: "Hoyt St", UniqueTripID: "C_20250822", UniqueTripServiceID: "C_20250822", StopSequence: 8, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 125, DepartureTimeInSeconds: epoch_20250822_120000_edt + 135},
				{UniqueStopID: "Franklin Av", UniqueTripID: "C_20250822", UniqueTripServiceID: "C_20250822", StopSequence: 9, ArrivalTimeInSeconds: epoch_20250822_120000_edt + 200, DepartureTimeInSeconds: epoch_20250822_120000_edt + 210},

				{UniqueStopID: "High St", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250823_120000_edt - 10, DepartureTimeInSeconds: epoch_20250823_120000_edt + 10},
				{UniqueStopID: "Jay St", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 120, DepartureTimeInSeconds: epoch_20250823_120000_edt + 130},
				{UniqueStopID: "Hoyt St", UniqueTripID: "C_20250823", UniqueTripServiceID: "C_20250823", StopSequence: 8, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 125, DepartureTimeInSeconds: epoch_20250823_120000_edt + 135},
				{UniqueStopID: "Franklin Av", UniqueTripID: "C_20250823", UniqueTripServiceID: "C_20250823", StopSequence: 9, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 200, DepartureTimeInSeconds: epoch_20250823_120000_edt + 210},

				{UniqueStopID: "High St", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250824_120000_edt - 10, DepartureTimeInSeconds: epoch_20250824_120000_edt + 10},
				{UniqueStopID: "Jay St", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 120, DepartureTimeInSeconds: epoch_20250824_120000_edt + 130},
				{UniqueStopID: "Hoyt St", UniqueTripID: "C_20250824", UniqueTripServiceID: "C_20250824", StopSequence: 8, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 125, DepartureTimeInSeconds: epoch_20250824_120000_edt + 135},
				{UniqueStopID: "Franklin Av", UniqueTripID: "C_20250824", UniqueTripServiceID: "C_20250824", StopSequence: 9, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 200, DepartureTimeInSeconds: epoch_20250824_120000_edt + 210},
			},
			Mode: RaptorModeDepartAt,
			/* 2025/08/23 12:00:00PM EDT */
			TimeInSeconds:        epoch_20250823_120000_edt,
			MaximumTransfers:     4,
			AllowTransferHopping: false,
		},
	)

	if len(journeys) == 0 {
		t.Fatalf(`did not find any journeys for stop times`)
	}

	if journeys[0].ArrivalTimeInSeconds != epoch_20250823_120000_edt+200 {
		t.Fatalf(`expected raptor to find arrival time %v but got %v`, epoch_20250823_120000_edt+200, journeys[0].ArrivalTimeInSeconds)
	}
}

func TestSimpleForwardRaptor_NoTransferStart(t *testing.T) {
	var epoch_20250823_120000_edt int64 = 1755964800
	var epoch_20250824_120000_edt int64 = 1756051200

	journeys := SimpleRaptor(
		SimpleRaptorInput[string, GtfsStopStruct[string], GtfsTransferStruct[string], GtfsStopTimeStruct[string]]{
			FromStops: []GtfsStopStruct[string]{
				{UniqueID: "SANDS ST/PEARL ST "},
				{UniqueID: "High St"},
			},
			ToStops: []GtfsStopStruct[string]{
				{UniqueID: "Franklin Av"},
			},
			Transfers: []GtfsTransferStruct[string]{
				{
					FromUniqueStopID:             "SANDS ST/PEARL ST ",
					ToUniqueStopID:               "High St",
					MinimumTransferTimeInSeconds: 0,
				},
			},
			StopTimes: []GtfsStopTimeStruct[string]{
				{UniqueStopID: "High St", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250823_120000_edt - 10, DepartureTimeInSeconds: epoch_20250823_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250823", UniqueTripServiceID: "A_20250823", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250823_120000_edt + 120, DepartureTimeInSeconds: epoch_20250823_120000_edt + 130},

				{UniqueStopID: "High St", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 5, ArrivalTimeInSeconds: epoch_20250824_120000_edt - 10, DepartureTimeInSeconds: epoch_20250824_120000_edt + 10},
				{UniqueStopID: "Franklin Av", UniqueTripID: "A_20250824", UniqueTripServiceID: "A_20250824", StopSequence: 6, ArrivalTimeInSeconds: epoch_20250824_120000_edt + 120, DepartureTimeInSeconds: epoch_20250824_120000_edt + 130},
			},
			Mode: RaptorModeDepartAt,
			/* 2025/08/23 12:00:00PM EDT */
			TimeInSeconds:        epoch_20250823_120000_edt,
			MaximumTransfers:     4,
			AllowTransferHopping: false,
		},
	)

	if len(journeys) == 0 {
		t.Fatalf(`did not find any journeys for stop times`)
	}

	if len(journeys) > 1 {
		t.Fatalf(`expected to find 1 journey - should not allow starting at Pearl St and then walking to High St`)
	}

	if journeys[0].ArrivalTimeInSeconds != epoch_20250823_120000_edt+120 {
		t.Fatalf(`expected raptor to find arrival time %v but got %v`, epoch_20250823_120000_edt+120, journeys[0].ArrivalTimeInSeconds)
	}
}
