package saga

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrSagaIsBroken        = errors.New("saga is broken")
	ErrSagaIsInProgress    = errors.New("saga is in progress")
	ErrSagaHasUnknownState = errors.New("saga has unknown state")
	ErrSagaIsDone          = errors.New("saga is done")
)

type Executor[Data any] struct {
	sagaDefinition    Definition[Data]
	sagaDataStorage   DataStorage[Data]
	transactionLogger TransactionLogger
}

func NewExecutor[Data any](
	sagaDefinition Definition[Data],
	sagaDataStorage DataStorage[Data],
	transactionLogger TransactionLogger,
) *Executor[Data] {
	return &Executor[Data]{
		sagaDefinition:    sagaDefinition,
		sagaDataStorage:   sagaDataStorage,
		transactionLogger: transactionLogger,
	}
}

func (e *Executor[Data]) ExecuteNextStep(ctx context.Context, sagaID ID) error {
	stepDefinition, isCompensation, err := e.getNextStepDefinitionForExecution(ctx, sagaID)
	if err != nil {
		return err
	}

	data, err := e.sagaDataStorage.Load(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("load saga data error: %w", err)
	}

	err = e.transactionLogger.LogStepStart(ctx, sagaID, stepDefinition.Name, isCompensation)
	if err != nil {
		return fmt.Errorf("step '%s': log step start error: %w", stepDefinition.Name, err)
	}

	if isCompensation {
		if stepDefinition.CompensationFunc != nil {
			err = stepDefinition.CompensationFunc(ctx, sagaID, data)
		}
	} else {
		err = stepDefinition.DoFunc(ctx, sagaID, data)
	}
	if err != nil {
		err = e.transactionLogger.LogFailedStep(ctx, sagaID, stepDefinition.Name, isCompensation, err)
		if err != nil {
			return fmt.Errorf("step '%s': log step unsuccessful end error: %w", stepDefinition.Name, err)
		}
		return fmt.Errorf("step '%s': execution error: %w", stepDefinition.Name, err)
	}

	err = e.sagaDataStorage.Save(ctx, sagaID, data)
	if err != nil {
		return fmt.Errorf("save saga data error: %w", err)
	}

	err = e.transactionLogger.LogSuccessStep(ctx, sagaID, stepDefinition.Name, isCompensation)
	if err != nil {
		return fmt.Errorf("step '%s': log step successful end error: %w", stepDefinition.Name, err)
	}

	return nil
}

func (e *Executor[Data]) getNextStepDefinitionForExecution(ctx context.Context, sagaID ID) (StepDefinition[Data], bool, error) {
	var (
		isCompensation    bool
		proceedToNextStep bool
		stepDefinition    StepDefinition[Data]
		stepName          StepName
	)

	lastLogEntry, err := e.transactionLogger.GetLastLogEntry(ctx, sagaID)
	if err != nil {
		if err == ErrLastTransactionLogEntryIsNotFound {
			return e.sagaDefinition.Steps[0], false, nil
		}
		return StepDefinition[Data]{}, false, fmt.Errorf("get transaction last log entry error: %w", err)
	}

	switch lastLogEntry.StepStateName {
	case StepStateNameInProgress:
		return StepDefinition[Data]{}, false, ErrSagaIsInProgress
	case StepStateNameSuccess:
		stepName = lastLogEntry.StepName
		proceedToNextStep = true
	case StepStateNameFailed:
		stepName = lastLogEntry.StepName
		isCompensation = true
	case StepStateNameCompensationInProgress:
		return StepDefinition[Data]{}, false, ErrSagaIsInProgress
	case StepStateNameCompensationSuccess:
		stepName = lastLogEntry.StepName
		isCompensation = true
		proceedToNextStep = true
	case StepStateNameCompensationFailed:
		return StepDefinition[Data]{}, false, ErrSagaIsBroken
	default:
		return StepDefinition[Data]{}, false, ErrSagaHasUnknownState
	}

	if proceedToNextStep {
		if isCompensation {
			stepDefinition, err = e.sagaDefinition.Steps.GetPrevStepDefinitionByCurrentStepName(stepName)
		} else {
			stepDefinition, err = e.sagaDefinition.Steps.GetNextStepDefinitionByCurrentStepName(stepName)
		}
		if err != nil {
			if err == ErrCurrentStepIsLast {
				return StepDefinition[Data]{}, false, ErrSagaIsDone
			}
			return StepDefinition[Data]{}, false, fmt.Errorf("get prev/next step definiton by current step name '%s' error", stepName)
		}
	} else {
		stepDefinition, err = e.sagaDefinition.Steps.GetStepDefinitionByStepName(stepName)
		if err != nil {
			return StepDefinition[Data]{}, false, fmt.Errorf("get step definiton by step name '%s' error", stepName)
		}
	}

	return stepDefinition, isCompensation, nil
}
