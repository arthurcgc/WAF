package manager

import "context"

type Manager interface {
	CreateInstance(context.Context, CreateArgs) error
	DeleteInstance(context.Context, DeleteArgs) error
}
