package manager

import "context"

type Manager interface {
	CreateInstance(context.Context, CreateArgs) error
	UpdateInstance(context.Context, UpdateArgs) error
	DeleteInstance(context.Context, DeleteArgs) error
}
