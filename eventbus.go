package eventbus

import (
	"context"
	"fmt"
	"sync"
)

var (
	ErrNotProcessor = fmt.Errorf("no processor registerd")
)

type Handler[TEvent any] func(ctx context.Context, event TEvent) error
type Processor[TEvent any, TResult any] func(ctx context.Context, event TEvent) (TResult, error)

type IDisposable interface {
	Dispose(ctx context.Context) error
}

type DisposeFunc func(ctx context.Context) error

func (d DisposeFunc) Dispose(ctx context.Context) error {
	return d(ctx)
}

type PublisherFunc[TEvent any] func(ctx context.Context, event TEvent) error

type DispatcherFunc[TEvent any, TResult any] func(ctx context.Context, event TEvent) (TResult, error)

type SubscribeFunc[TEvent any] func(ctx context.Context, handler Handler[TEvent]) (IDisposable, error)

type SubscribeOnceFunc[TEvent any] func(ctx context.Context, handler Handler[TEvent]) (IDisposable, error)

type ProcessableFunc[TEvent any, TResult any] func(ctx context.Context, processor Processor[TEvent, TResult]) (IDisposable, error)

type handler interface {
	CanHandle(ctx context.Context, event interface{}) bool
	Handle(ctx context.Context, event interface{}) error
}

type handlerImpl struct {
	canHandlerFunc func(ctx context.Context, event interface{}) bool
	handleFunc     func(ctx context.Context, event interface{}) error
}

func (h *handlerImpl) CanHandle(ctx context.Context, event interface{}) bool {
	return h.canHandlerFunc(ctx, event)
}

func (h *handlerImpl) Handle(ctx context.Context, event interface{}) error {
	return h.handleFunc(ctx, event)
}

type processor interface {
	CanProcess(ctx context.Context, event interface{}, result interface{}) bool
	Process(ctx context.Context, event interface{}) (result interface{}, err error)
}

type processorImpl struct {
	canProcessFunc func(ctx context.Context, event interface{}, result interface{}) bool
	processFunc    func(ctx context.Context, event interface{}) (result interface{}, err error)
}

func (p *processorImpl) CanProcess(ctx context.Context, event interface{}, result interface{}) bool {
	return p.canProcessFunc(ctx, event, result)
}

func (p *processorImpl) Process(ctx context.Context, event interface{}) (result interface{}, err error) {
	return p.processFunc(ctx, event)
}

type EventBus struct {
	handlers     []handler
	processors   []processor
	handleLock   sync.Mutex
	dispatchLock sync.Mutex
}

func New() *EventBus {
	return &EventBus{}
}

func (e *EventBus) publish(ctx context.Context, event interface{}) error {
	e.handleLock.Lock()
	defer e.handleLock.Unlock()
	for _, h := range e.handlers {
		if h.CanHandle(ctx, event) {
			err := h.Handle(ctx, event)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *EventBus) dispatch(ctx context.Context, event interface{}, result interface{}) (interface{}, error) {
	e.dispatchLock.Lock()
	defer e.dispatchLock.Unlock()
	for _, h := range e.processors {
		if h.CanProcess(ctx, event, result) {
			result, err := h.Process(ctx, event)
			return result, err
		}
	}
	return nil, ErrNotProcessor
}

func (e *EventBus) subscribe(ctx context.Context, h handler) (IDisposable, error) {
	e.handleLock.Lock()
	defer e.handleLock.Unlock()
	e.handlers = append(e.handlers, h)
	return DisposeFunc(func(ctx context.Context) error {
		e.handleLock.Lock()
		defer e.handleLock.Unlock()
		e.handlers = removeHandler(e.handlers, h)
		return nil
	}), nil
}

func (e *EventBus) subscriberOnce(ctx context.Context, h handler) (IDisposable, error) {
	e.handleLock.Lock()
	defer e.handleLock.Unlock()

	wrapper := &handlerImpl{
		canHandlerFunc: h.CanHandle,
	}

	dispose := DisposeFunc(func(ctx context.Context) error {
		//do not need to lock due to locked by publish
		e.handlers = removeHandler(e.handlers, handler(wrapper))
		return nil
	})

	wrapper.handleFunc = func(ctx context.Context, event interface{}) error {
		err := h.Handle(ctx, event)
		if err != nil {
			return err
		}
		return dispose(ctx)
	}

	e.handlers = append(e.handlers, wrapper)
	//dispose
	return dispose, nil
}

func (e *EventBus) addProcessor(ctx context.Context, p processor) (IDisposable, error) {
	e.dispatchLock.Lock()
	defer e.dispatchLock.Unlock()
	e.processors = append(e.processors, p)
	return DisposeFunc(func(ctx context.Context) error {
		e.dispatchLock.Lock()
		defer e.dispatchLock.Unlock()
		e.processors = removeProcessor(e.processors, p)
		return nil
	}), nil
}

func Publish[TEvent any](e *EventBus) PublisherFunc[TEvent] {
	return func(ctx context.Context, event TEvent) error {
		return e.publish(ctx, event)
	}
}

//Dispatch return processed result, ErrNotProcessor returned if no matching processor
func Dispatch[TEvent any, TResult any](e *EventBus) DispatcherFunc[TEvent, TResult] {
	return func(ctx context.Context, event TEvent) (TResult, error) {
		var r TResult
		result, err := e.dispatch(ctx, event, r)
		if err != nil {
			return r, err
		}
		return result.(TResult), err
	}
}

func Subscribe[TEvent any](e *EventBus) SubscribeFunc[TEvent] {
	return func(ctx context.Context, handler Handler[TEvent]) (IDisposable, error) {
		wrapper := func(ctx context.Context, event interface{}) error {
			return handler(ctx, event.(TEvent))
		}
		return e.subscribe(ctx, &handlerImpl{
			canHandlerFunc: func(ctx context.Context, event interface{}) bool {
				_, ok := event.(TEvent)
				return ok
			},
			handleFunc: wrapper,
		})
	}
}

func SubscribeOnce[TEvent any](e *EventBus) SubscribeOnceFunc[TEvent] {
	return func(ctx context.Context, handler Handler[TEvent]) (IDisposable, error) {
		wrapper := func(ctx context.Context, event interface{}) error {
			return handler(ctx, event.(TEvent))
		}
		return e.subscriberOnce(ctx, &handlerImpl{
			canHandlerFunc: func(ctx context.Context, event interface{}) bool {
				_, ok := event.(TEvent)
				return ok
			},
			handleFunc: wrapper,
		})
	}
}

func AddProcessor[TEvent any, TResult any](e *EventBus) ProcessableFunc[TEvent, TResult] {
	return func(ctx context.Context, processor Processor[TEvent, TResult]) (IDisposable, error) {
		wrapper := func(ctx context.Context, event interface{}) (interface{}, error) {
			result, err := processor(ctx, event.(TEvent))
			return result, err
		}
		return e.addProcessor(ctx, &processorImpl{
			canProcessFunc: func(ctx context.Context, event interface{}, result interface{}) bool {
				_, ok := event.(TEvent)
				if !ok {
					return false
				}
				//check TResult interface{}
				if result == nil {
					return true
				}

				_, ok = result.(TResult)
				if !ok {
					return false
				}
				return true
			},
			processFunc: wrapper,
		})
	}
}

func removeHandler(l []handler, item handler) []handler {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func removeProcessor(l []processor, item processor) []processor {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}
