package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"

	"doc_service/domain"
	entity "doc_service/domain"
)

const defaultTTL = 15 * time.Minute

type InMemoryCache struct {
	c *cache.Cache
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		c: cache.New(defaultTTL, time.Minute),
	}
}

func (m *InMemoryCache) Set(doc *entity.Doc) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	m.c.Set(doc.ID, data, defaultTTL)
	return nil
}

func (m *InMemoryCache) Get(id string) (*entity.Doc, error) {
	raw, found := m.c.Get(id)
	if !found {
		return nil, domain.ErrNotFound
	}

	data, ok := raw.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid type assertion for cache value")
	}

	var doc entity.Doc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (m *InMemoryCache) Flush() {
	m.c.Flush()
}
