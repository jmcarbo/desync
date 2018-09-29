package desync

import (
	"strings"

	"github.com/pkg/errors"
)

// StoreRouter is used to route requests to multiple stores. When a chunk is
// requested from the router, it'll query the first store and if that returns
// ChunkMissing, it'll move on to the next.
type StoreRouter struct {
	Stores []Store
}

// NewStoreRouter returns an initialized router
func NewStoreRouter(stores ...Store) StoreRouter {
	var l []Store
	for _, s := range stores {
		l = append(l, s)
	}
	return StoreRouter{l}
}

// GetChunk queries the available stores in order and moves to the next if
// it gets a ChunkMissing. Fails if any store returns a different error.
func (r StoreRouter) GetChunk(id ChunkID) (*Chunk, error) {
	for _, s := range r.Stores {
		chunk, err := s.GetChunk(id)
		switch err.(type) {
		case nil:
			return chunk, nil
		case ChunkMissing:
			continue
		default:
			return nil, errors.Wrap(err, s.String())
		}
	}
	return nil, ChunkMissing{id}
}

// HasChunk returns true if one of the containing stores has the chunk. It
// goes through the stores in order and returns as soon as the chunk is found.
func (r StoreRouter) HasChunk(id ChunkID) bool {
	for _, s := range r.Stores {
		if s.HasChunk(id) {
			return true
		}
	}
	return false
}

func (r StoreRouter) String() string {
	var a []string
	for _, s := range r.Stores {
		a = append(a, s.String())
	}
	return strings.Join(a, ",")
}

// Close calls the Close() method on every store in the router. Returns
// only the first error encountered.
func (r StoreRouter) Close() error {
	var sErr error
	for _, s := range r.Stores {
		if err := s.Close(); err != nil {
			if sErr == nil {
				sErr = err
			}
		}
	}
	return sErr
}
