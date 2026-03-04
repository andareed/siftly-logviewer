package hostlog

import (
	"github.com/andareed/siftly-hostlog/internal/siftly"
)

type Model struct {
	*siftly.Model // embed the shared model
}

// func NewModel() (*Model, error) {
// 	items, err := loadItemsSync() // sync load
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Model{CoreModel: core.New(items)}, nil
// }
