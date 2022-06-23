package metering

type AggregationInterval string

const (
	Hour  AggregationInterval = "HOUR"
	Day   AggregationInterval = "DAY"
	Week  AggregationInterval = "WEEK"
	Month AggregationInterval = "MONTH"
)

type Take struct {
	Limit       int64 `json:"limit"`
	IsAscending bool  `json:"isAscending,omitempty"`
}

type TimeRange struct {
	StartTimeInSeconds int64 `json:"startTimeInSeconds"`
	EndTimeInSeconds   int64 `json:"endTimeInSeconds,omitempty"`
}
