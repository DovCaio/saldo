package transaction

import "time"

type SummaryGroupBy string

const (
    DAY   SummaryGroupBy = "DAY"
    MONTH SummaryGroupBy = "MONTH"
    YEAR  SummaryGroupBy = "YEAR"
)

type TransactionSummaryFilter struct {
    DateFrom *time.Time
    DateTo   *time.Time
    GroupBy  SummaryGroupBy
}