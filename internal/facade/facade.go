package facade

import (
	"database/sql"

	"final/internal/factory"
	"final/internal/store"
	"final/internal/strategy"
)

type Facade struct {
	posts     *store.PostStore
	factory   *factory.Factory
	moderator *strategy.Moderator
}

func New(db *sql.DB) *Facade {
	strict, _ := strategy.ByName("strict")
	return &Facade{
		posts:     store.NewPostStore(db),
		factory:   factory.NewFactory(),
		moderator: strategy.NewModerator(strict),
	}
}
