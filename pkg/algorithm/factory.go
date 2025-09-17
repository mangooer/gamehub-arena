package algorithm

import (
	"fmt"
	"sync"
)

type AlgorithmFactory struct {
	algorithms map[string]func() MatchingAlgorithm
	instances  map[string]MatchingAlgorithm
	mu         sync.RWMutex
}

var (
	defaultFactory *AlgorithmFactory
	once           sync.Once
)

func InitFactory() *AlgorithmFactory {
	once.Do(func() {
		defaultFactory = &AlgorithmFactory{
			algorithms: make(map[string]func() MatchingAlgorithm),
			instances:  make(map[string]MatchingAlgorithm),
		}
	})
	return defaultFactory
}

func (f *AlgorithmFactory) RegisterAlgorithm(name string, algorithm func() MatchingAlgorithm) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.algorithms[name] = algorithm
}

func (f *AlgorithmFactory) GetAlgorithm(name string) (MatchingAlgorithm, error) {
	// 🔍 快速路径：使用读锁检查已存在的实例
	f.mu.RLock()
	instance, exists := f.instances[name]
	f.mu.RUnlock()

	if exists {
		return instance, nil
	}

	// 🔒 慢速路径：使用写锁创建实例
	f.mu.Lock()
	defer f.mu.Unlock()

	// 🔍 再次检查，防止竞态条件
	if instance, exists := f.instances[name]; exists {
		return instance, nil
	}

	// 🏭 创建新实例
	fn, exists := f.algorithms[name]
	if !exists {
		return nil, fmt.Errorf("algorithm %s not found", name)
	}

	instance = fn()
	f.instances[name] = instance
	return instance, nil
}
