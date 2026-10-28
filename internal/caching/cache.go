// Package caching composes optional data caches; authority stays in Postgres.
package caching

import (
	"github.com/plutoframework/pluto-example/internal/config"

	"github.com/plutoframework/pluto-core/cache"
	"github.com/plutoframework/pluto-core/cache/memory"
	valkey "github.com/plutoframework/pluto-core/stores/valkey"
)

func New(settings config.CacheConfig) (cache.Store, func(), error) {
	switch settings.Backend {
	case "", "none":
		return cache.Disabled{}, func() {}, nil
	case "memory":
		store, err := memory.New(memory.DefaultConfig())
		return store, func() {}, err
	case "valkey":
		options := valkey.DefaultConfig()
		options.URL, options.AllowInsecure = settings.URL, settings.AllowInsecure

		store, err := valkey.New(options)
		if err != nil {
			return nil, nil, err
		}

		return store, store.Close, nil
	default:
		return nil, nil, cache.ErrInvalid
	}
}
