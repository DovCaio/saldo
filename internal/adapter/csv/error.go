package csv

import (
	"errors"
)

var (
	InvalidCSVHeader = errors.New("invalid CSV header")
	CSVIsEmpty = errors.New("CSV is empty")
)