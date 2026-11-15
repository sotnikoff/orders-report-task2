package repositories

import (
	"ordersreport/internal/application"
)

type LinesRepository struct {
	adapter application.LineSource
}

func NewLinesRepository(adapter application.LineSource) *LinesRepository {
	return &LinesRepository{adapter: adapter}
}

func (l LinesRepository) Load() ([]string, error) {
	return l.adapter.Load()
}
