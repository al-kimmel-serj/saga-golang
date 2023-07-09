package saga

type (
	ID string
)

type Definition[Data any] struct {
	Steps StepsDefinition[Data]
}

func NewDefinition[Data any](steps StepsDefinition[Data]) Definition[Data] {
	return Definition[Data]{
		Steps: steps,
	}
}
