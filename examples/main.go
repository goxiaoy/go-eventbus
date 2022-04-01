package main

import (
	"context"
	"fmt"
	eventbus "github.com/goxiaoy/go-eventbus"
)

type TestEvent1 struct {
}

type TestResult1 struct {
}

func main() {
	bus := eventbus.New()
	ctx := context.Background()
	//Subscribe
	dispose, err := eventbus.Subscribe[*TestEvent1](bus)(func(ctx context.Context, event *TestEvent1) error {
		fmt.Println("do TestEvent1")
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
	err = dispose.Dispose()
	if err != nil {
		panic(err)
	}

	//You can also subscribe to any
	dispose, err = eventbus.Subscribe[interface{}](bus)(func(ctx context.Context, event interface{}) error {
		fmt.Println("do any")
		return nil
	})
	if err != nil {
		panic(err)
	}

	dispose, err = eventbus.SubscribeOnce[interface{}](bus)(func(ctx context.Context, event interface{}) error {
		fmt.Println("do any once")
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
	//Publish twice
	err = eventbus.Publish[*TestEvent1](bus)(ctx, &TestEvent1{})
	if err != nil {
		panic(err)
	}

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
	if result == nil {
		panic("no result")
	}
	dispose.Dispose()

	////Any Processor
	//dispose, err = eventbus.AddProcessor[interface{}, interface{}](bus)(ctx, func(ctx context.Context, event interface{}) (interface{}, error) {
	//	return &TestResult1{}, err
	//})
	//
	//resultAny, err := eventbus.Dispatch[*TestEvent1, interface{}](bus)(ctx, &TestEvent1{})
	//if err != nil {
	//	panic(err)
	//}
	//if resultAny == nil {
	//	panic("no result")
	//}
	//result, ok := resultAny.(*TestResult1)
	//if !ok {
	//	panic("result not match")
	//}
	//if result == nil {
	//	panic("no result")
	//}
}
