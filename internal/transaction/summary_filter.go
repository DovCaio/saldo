package transaction

import "time"

type SummaryGroupBy string

const (
    DAY   SummaryGroupBy = "day"
    MONTH SummaryGroupBy = "month"
    YEAR  SummaryGroupBy = "year"
)

type TransactionSummaryFilter struct {
    DateFrom *time.Time
    DateTo   *time.Time
    GroupBy  SummaryGroupBy
}