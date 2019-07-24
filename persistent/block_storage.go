package persistent

import (
	"context"
	"fmt"
)

type simpleBlock struct {
	base *BufferedStorage
}

// NewSimpleBlock turns a BufferedStorage implementation into a BlockStorage
// implementation. It simply converts the block pointer into a hex string and
// uses that as the key.
func NewSimpleBlock(base *BufferedStorage) BlockStorage {
	return simpleBlock{base}
}

func (sb simpleBlock) Start(ctx context.Context) error {
	return sb.base.Start(ctx)
}

func (sb simpleBlock) Get(ctx context.Context, ptr uint64) ([]byte, error) {
	return sb.base.Get(ctx, fmt.Sprintf("%x", ptr))
}

func (sb simpleBlock) GetMany(ctx context.Context, ptrs []uint64) (map[uint64][]byte, error) {
	keys := make([]string, 0, len(ptrs))
	conversion := make(map[string]uint64)
	for _, ptr := range ptrs {
		key := fmt.Sprintf("%x", ptr)
		keys = append(keys, key)
		conversion[key] = ptr
	}

	data, err := sb.base.GetMany(ctx, keys)
	if err != nil {
		return nil, err
	}

	out := make(map[uint64][]byte)
	for key, val := range data {
		ptr, ok := conversion[key]
		if !ok {
			return nil, fmt.Errorf("given value for unexpected key")
		}
		out[ptr] = val
	}
	return out, nil
}

func (sb simpleBlock) Set(ctx context.Context, ptr uint64, data []byte) error {
	return sb.base.Set(ctx, fmt.Sprintf("%x", ptr), data)
}

func (sb simpleBlock) Commit(ctx context.Context) error {
	return sb.base.Commit(ctx)
}

func (sb simpleBlock) Rollback(ctx context.Context) {
	sb.base.Rollback(ctx)
}

type blockMemory map[uint64][]byte

// NewBlockMemory returns an implementation of BlockStorage that simply stores
// data in-memory.
func NewBlockMemory() BlockStorage { return make(blockMemory) }

func (bm blockMemory) Start(ctx context.Context) error { return nil }

func (bm blockMemory) Get(ctx context.Context, ptr uint64) ([]byte, error) {
	d, ok := bm[ptr]
	if !ok {
		return nil, ErrObjectNotFound
	}
	return dup(d), nil
}

func (bm blockMemory) GetMany(ctx context.Context, ptrs []uint64) (map[uint64][]byte, error) {
	out := make(map[uint64][]byte)
	for _, ptr := range ptrs {
		val, ok := bm[ptr]
		if !ok {
			continue
		}
		out[ptr] = dup(val)
	}
	return out, nil
}

func (bm blockMemory) Set(ctx context.Context, ptr uint64, data []byte) error {
	bm[ptr] = dup(data)
	return nil
}

func (bm blockMemory) Commit(ctx context.Context) error { return nil }
func (bm blockMemory) Rollback(ctx context.Context)     {}
