package saga

import (
	"context"
	"errors"
)

const (
	StepStateNameInProgress             StepStateName = "in-progress"
	StepStateNameSuccess                StepStateName = "success"
	StepStateNameFailed                 StepStateName = "failed"
	StepStateNameCompensationInProgress StepStateName = "compensation-in-progress"
	StepStateNameCompensationSuccess    StepStateName = "compensation-success"
	StepStateNameCompensationFailed     StepStateName = "compensation-failed"
)

var (
	ErrCurrentStepIsLast     = errors.New("current step is last")
	ErrCurrentStepIsNotFound = errors.New("current step is not found")
)

type (
	CompensationFunc[Data any] func(ctx context.Context, sagaID ID, data *Data) error
	DoFunc[Data any]           func(ctx context.Context, sagaID ID, data *Data) error
	StepName                   string
	StepStateName              string
	StepsDefinition[Data any]  []StepDefinition[Data]
)

type StepDefinition[Data any] struct {
	Name             StepName
	DoFunc           DoFunc[Data]
	CompensationFunc CompensationFunc[Data]
}

func (d StepsDefinition[Data]) GetStepDefinitionByStepName(stepName StepName) (StepDefinition[Data], error) {
	for _, stepDefinition := range d {
		if stepDefinition.Name == stepName {
			return stepDefinition, nil
		}
	}

	return StepDefinition[Data]{}, ErrCurrentStepIsNotFound
}

func (d StepsDefinition[Data]) GetPrevStepDefinitionByCurrentStepName(currentStepName StepName) (StepDefinition[Data], error) {
	for currentStepIndex := len(d) - 1; currentStepIndex >= 0; currentStepIndex-- {
		stepDefinition := d[currentStepIndex]

		if stepDefinition.Name == currentStepName {
			if currentStepIndex == 0 {
				return StepDefinition[Data]{}, ErrCurrentStepIsLast
			}

			return d[currentStepIndex-1], nil
		}
	}

	return StepDefinition[Data]{}, ErrCurrentStepIsNotFound
}

func (d StepsDefinition[Data]) GetNextStepDefinitionByCurrentStepName(currentStepName StepName) (StepDefinition[Data], error) {
	for currentStepIndex, stepDefinition := range d {
		if stepDefinition.Name == currentStepName {
			if currentStepIndex == len(d)-1 {
				return StepDefinition[Data]{}, ErrCurrentStepIsLast
			}

			return d[currentStepIndex+1], nil
		}
	}

	return StepDefinition[Data]{}, ErrCurrentStepIsNotFound
}
