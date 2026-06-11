package bridge

import (
	"log/slog"
	"strconv"
	"time"

	bmwcardata "github.com/tjamet/bmw-cardata"
	"github.com/tjamet/bmw-cardata/cardataapi"

	"github.com/wkulhane/bmw-loxone-bridge/internal/names"
	"github.com/wkulhane/bmw-loxone-bridge/internal/store"
)

type Bridge struct {
	Store  *store.Store
	Logger *slog.Logger
}

func New(s *store.Store, logger *slog.Logger) *Bridge {
	return &Bridge{Store: s, Logger: logger}
}

func (b *Bridge) HandleMessage(msg bmwcardata.StreamedMessage) {
	for name, details := range msg.Data {
		ts := parseTimestamp(details.Timestamp)
		val, numeric := extractValue(details.Value)

		dp := store.DataPoint{
			Value:     val,
			Numeric:   numeric,
			Unit:      details.Unit,
			Timestamp: ts,
			RawName:   name,
		}

		friendlyName := names.ToFriendly(name)
		b.Store.Set(friendlyName, dp)

		b.Logger.Debug("data point updated",
			"name", friendlyName,
			"raw", name,
			"value", val,
			"unit", details.Unit,
		)
	}
}

func extractValue(v bmwcardata.StreamedDataValue) (string, *float64) {
	if v.String != nil {
		return *v.String, nil
	}
	if v.Float != nil {
		f := *v.Float
		return strconv.FormatFloat(f, 'f', -1, 64), &f
	}
	if v.Int != nil {
		f := float64(*v.Int)
		return strconv.FormatInt(*v.Int, 10), &f
	}
	if v.Bool != nil {
		if *v.Bool {
			f := 1.0
			return "true", &f
		}
		f := 0.0
		return "false", &f
	}
	return "", nil
}

func (b *Bridge) HandleTelematicData(data map[string]cardataapi.TelematicDataEntryDto) {
	for name, entry := range data {
		var ts time.Time
		if entry.Timestamp != nil {
			ts = parseTimestamp(*entry.Timestamp)
		} else {
			ts = time.Now()
		}

		val := ""
		if entry.Value != nil {
			val = *entry.Value
		}
		unit := ""
		if entry.Unit != nil {
			unit = *entry.Unit
		}

		var numeric *float64
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			numeric = &f
		}

		dp := store.DataPoint{
			Value:     val,
			Numeric:   numeric,
			Unit:      unit,
			Timestamp: ts,
			RawName:   name,
		}

		friendlyName := names.ToFriendly(name)
		b.Store.Set(friendlyName, dp)

		b.Logger.Debug("REST data point loaded",
			"name", friendlyName,
			"raw", name,
			"value", val,
			"unit", unit,
		)
	}
}

func parseTimestamp(ts string) time.Time {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return time.Now()
		}
	}
	return t
}
