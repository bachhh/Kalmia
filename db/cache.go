package db

import (
	"errors"
	"strings"
	"sync"

	"git.difuse.io/Difuse/kalmia/logger"
)

//nolint:gochecknoglobals
var Cache *sync.Map

type CacheEntry struct {
	Data        []byte
	ContentType string
}

func InitCache() {
	Cache = &sync.Map{}
	logger.Info("Cache initialized")
}

func SetKey(key []byte, value []byte, contentType string) error {
	Cache.Store(string(key), CacheEntry{
		Data:        value,
		ContentType: contentType,
	})
	return nil
}

func GetValue(key []byte) (CacheEntry, error) {
	if entry, ok := Cache.Load(string(key)); ok {
		if cacheEntry, ok := entry.(CacheEntry); ok {
			return cacheEntry, nil
		}
	}
	return CacheEntry{}, ErrKeyNotFound
}

func ClearCacheByPrefix(prefix string) error {
	Cache.Range(func(k, v interface{}) bool {
		if kStr, ok := k.(string); ok {
			if strings.HasPrefix(kStr, prefix) {
				Cache.Delete(k)
			}
		}
		return true
	})
	return nil
}

func GetCacheByPrefix(prefix string) (map[string]string, error) {
	result := make(map[string]string)
	Cache.Range(func(k, v interface{}) bool {
		if kStr, ok := k.(string); ok {
			if strings.HasPrefix(kStr, prefix) {
				if entry, ok := v.(CacheEntry); ok {
					result[kStr] = string(entry.Data)
				}
			}
		}
		return true
	})
	return result, nil
}

var ErrKeyNotFound = errors.New("key not found")
