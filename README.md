# go-eventbus
simple strong typed event bus from golang generics.

you can use it as a local eventbus or a mediator for [CQRS](https://en.wikipedia.org/wiki/Command%E2%80%93query_separation)

- [x] Publish, Subscribe/SubscribeOnce/Subscribe (with type assertion)
- [x] Dispatch, AddProcessor

## Install
```
go get github.com/goxiaoy/go-eventbus
```

## Usage

### create a eventbus or skip to use Default one
```go
bus := eventbus.New()
```

### Publish and Subscribe
```go
type TestEvent1 struct {
}

func main() {
    ctx := context.Background()
    //Subscribe
    dispose, err := eventbus.Subscribe[*TestEvent1](bus)(func(ctx context.Context, event *TestEvent1) error {
        fmt.Print("do TestEvent1")
        return nil
    })
    if err != nil {
        panic(err)
    }
    //Publish
    err = eventbus.Publish[*TestEvent1](bus)(ctx, &TestEvent1{})
    if err != nil {
        panic(err)
    }
    //UnSubscribe
    err = dispose.Dispose(ctx)
    if err != nil {
        panic(err)
    }

}
```
Subscribe to any

```go
//You can also subscribe to any
dispose, err = eventbus.Subscribe[interface{}](bus)(func(ctx context.Context, event interface{}) error {
    fmt.Println("do any")
    return nil
})
```

Subscribe once
```go
dispose, err = eventbus.SubscribeOnce[interface{}](bus)(func(ctx context.Context, event interface{}) error {
    fmt.Println("do any")
    return nil
})
```

### Dispatch and Process

```go
//Processor
dispose, err = eventbus.AddProcessor[*TestEvent1, *TestResult1](bus)(func(ctx context.Context, event *TestEvent1) (*TestResult1, error) {
    fmt.Println("return result")
    return &TestResult1{}, err
})
//Dispatch
result, err := eventbus.Dispatch[*TestEvent1, *TestResult1](bus)(ctx, &TestEvent1{})
if err != nil {
    panic(err)
}
```

## Limitation

- This eventbus is designed as `O(n)` complexity.
- Do not support Dispatch/Process Result interface type  

