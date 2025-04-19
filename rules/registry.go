package rules

import "sync"

type CustomFunc func(args ...interface{}) bool

var (
	customFuncRegistry = make(map[string]CustomFunc)
	registryLock       sync.RWMutex
)

func RegisterFunc(name string, fn CustomFunc) {
	registryLock.Lock()
	defer registryLock.Unlock()
	customFuncRegistry[name] = fn
}

func GetFunc(name string) (CustomFunc, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()
	fn, ok := customFuncRegistry[name]
	return fn, ok
}
