package saga

import (
	"context"
	"errors"
)

var (
	ErrLastTransactionLogEntryIsNotFound = errors.New("last transaction log entry is not found")
)

type TransactionLogEntry struct {
	StepName      StepName
	StepStateName StepStateName
}

type TransactionLogger interface {
	GetLastLogEntry(ctx context.Context, sagaID ID) (TransactionLogEntry, error)
	LogStepStart(ctx context.Context, sagaID ID, stepName StepName, isCompensation bool) error
	LogSuccessStep(ctx context.Context, sagaID ID, stepName StepName, isCompensation bool) error
	LogFailedStep(ctx context.Context, sagaID ID, stepName StepName, isCompensation bool, err error) error
}
