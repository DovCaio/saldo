package transaction

import "errors"

var (
    ErrEmptyImport       = errors.New("import contains no transactions")
    ErrDuplicateInBatch  = errors.New("duplicate transaction in import batch")
)