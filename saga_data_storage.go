package saga

import "context"

type DataStorage[Data any] interface {
	Load(ctx context.Context, sagaID ID) (*Data, error)
	Save(ctx context.Context, sagaID ID, data *Data) error
}
