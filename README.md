# go-throttle
Golang function throttler.  
Similar to debounce, but the first call will execute immediately.  
Subsequent calls will always have a minimum duration between executions.

### Examples

Simple usage:
```go
throttler := throttle.New(time.Millisecond * 250)

for i := 0; i < 10; i++ {
    if i == 0 {
        throttler(func() { fmt.Println("hello")})
    } else {
        throttler(func() { fmt.Println("world")})	
    }
    time.Sleep(time.Millisecond * 10)
}

// Output:
// hello (immediately)
// world (after 250 ms)
```


Reloading a cache with a 5m minimum interval:
```go
package cache

import (
    "sync"
    "time"
    
    "github.com/frostbyte73/go-throttle"
)

type DataSource interface {
    Reload() map[string]interface{}
}

type Cache struct {
    DataSource
    
    mu         sync.RWMutex
    data       map[string]interface{}
    throttle   func(func())
}

func NewCache(source DataSource) *Cache {
    c := &Cache{
        DataSource: source,
        throttle:   throttle.New(time.Minute * 5),
    }
    
    // load initial cache data
    c.throttle(c.reloadCache)
    
    return c
}

func (c *Cache) Get(key string) interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    value, ok := c.data[key]
    if !ok {
        // cache miss - queue a reload, max once every 5 minutes
        c.throttle(c.reloadCache)
    }
    return value
}

func (c *Cache) reloadCache() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data = c.DataSource.Reload()
}
```
