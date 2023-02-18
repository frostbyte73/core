# Core

Things I wish were built in.

## Fuse

A thread-safe one-way switch, used for permanent state changes.

```go
type Server struct {
    msgs chan *Message
	shutdown core.Fuse
}

func NewServer() *Server {
    return &Server{
        shutdown: core.NewFuse()	
    }
}

func (s *Server) Run() {
    for {
        select {
        case msg := <-s.msgs:
            s.Process(msg)
        case <-s.shutdown.Wire():
            return
        }   
    }
}

func (s *Server) DoSomething() {
    if s.shutdown.IsClosed() {
        return errors.New("server closed")
    }
    ...
}

func (s *Server) Shutdown() {
    s.shutdown.Close()
}
```

## Throttle

Golang function throttler.  
Similar to debounce, but the first call will execute immediately.  
Subsequent calls will always have a minimum duration between executions.

```go
func main() {
    throttle := core.NewThrottle(time.Millisecond * 250)
    
    for i := 0; i < 10; i++ {
        if i == 0 {
            throttle(func() { fmt.Println("hello") })
        } else {
            throttle(func() { fmt.Println("world") })
        }
        time.Sleep(time.Millisecond * 10)
    }
}
```

Output:

```
hello (immediately)
world (after 250 ms)
```
