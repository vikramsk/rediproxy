package mocks

import "github.com/vikramsk/rediproxy/pkg/cache"

// ensure that the mocks satisfy the interfaces.
var _ = cache.Getter(&Getter{})
var _ = cache.Setter(&Setter{})

// Getter is a mock implementation of
// cache.Getter
type Getter struct {
	GetFn        func(key string) (string, error)
	GetFnInvoked bool
}

// Setter is a mock implementation of
// cache.Writer
type Setter struct {
	SetFn        func(key, value string)
	SetFnInvoked bool
}

// Get is a mock implementation of the Get func.
func (cr *Getter) Get(key string) (string, error) {
	cr.GetFnInvoked = true
	return cr.GetFn(key)
}

// Set is a mock implementation of the Set func.
func (cw *Setter) Set(key, value string) {
	cw.SetFnInvoked = true
	cw.SetFn(key, value)
}
