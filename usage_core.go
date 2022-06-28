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

type UsageBase struct {
	Logger Logger
}

type UsageOption func(*UsageBase)

func WithUsageLogger(logger Logger) UsageOption {
	return func(u *UsageBase) {
		u.Logger = logger
	}
}

func (ub *UsageBase) GetLoggerInstance(opts ...UsageOption) Logger {
	//iterate through each option
	for _, opt := range opts {
		opt(ub)
	}

	if ub.Logger != nil {
		return ub.Logger
	}

	return NewAmberfloDefaultLogger()
}
