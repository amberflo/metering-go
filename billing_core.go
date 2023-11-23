package metering

type BillingPeriod struct {
	Interval       BillingPeriodInterval `json:"interval"`
	IntervalsCount int32                 `json:"intervalsCount"`
}

type BillingPeriodInterval string

const (
	DAY   BillingPeriodInterval = "DAY"
	MONTH BillingPeriodInterval = "MONTH"
	YEAR  BillingPeriodInterval = "YEAR"
)

type PrepaidPriority string

const (
	HIGHEST_PRIORITY PrepaidPriority = "HIGHEST_PRIORITY"
	LOWEST_PRIORITY  PrepaidPriority = "LOWEST_PRIORITY"
)

type PaymentStatus string

const (
	PRE_PAYMENT     PaymentStatus = "PRE_PAYMENT"
	REQUIRES_ACTION PaymentStatus = "REQUIRES_ACTION"
	PENDING         PaymentStatus = "PENDING"
	FAILED          PaymentStatus = "FAILED"
	SETTLED         PaymentStatus = "SETTLED"
	NOT_NEEDED      PaymentStatus = "NOT_NEEDED"
	UNKNOWN         PaymentStatus = "UNKNOWN"
)
