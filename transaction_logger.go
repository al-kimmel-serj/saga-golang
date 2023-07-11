package saga

import (
	"context"
	"errors"
)

var (
	ErrTransactionLastLogEntryIsNotFound = errors.New("transaction last log entry is not found")
)

type TransactionLastLogEntry struct {
	StepName      StepName
	StepStateName StepStateName
}

type TransactionLogger interface {
	GetLastLogEntry(ctx context.Context, sagaID ID) (TransactionLastLogEntry, error)
	LogStepStart(ctx context.Context, sagaID ID, stepName StepName, isCompensation bool) error
	LogSuccessStep(ctx context.Context, sagaID ID, stepName StepName, isCompensation bool) error
	LogFailedStep(ctx context.Context, sagaID ID, stepName StepName, isCompensation bool, err error) error
}
