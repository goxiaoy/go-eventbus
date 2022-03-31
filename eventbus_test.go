package eventbus

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestEvent1 struct {
}

type TestEvent2 struct {
}

type TestResult1 struct {
}

type TestResult2 struct {
}

type TestInterface interface {
	GetTest()
}

func TestPublishAndSubscribe(t *testing.T) {
	bus := New()
	ctx := context.Background()

	event1 := &TestEvent1{}
	event1_1 := &TestEvent1{}
	event2_1 := &TestEvent2{}

	counter1 := 0
	dispose1, err := Subscribe[interface{}](bus)(ctx, func(ctx context.Context, event interface{}) error {

		if counter1 == 0 {
			assert.Equal(t, event1, event)
		}
		if counter1 == 1 {
			assert.Equal(t, event1_1, event)
		}
		if counter1 == 2 {
			assert.Equal(t, event2_1, event)
		}
		counter1++
		return nil
	})

	assert.NoError(t, err)

	counter1_2 := 0
	_, err = Subscribe[*TestEvent1](bus)(ctx, func(ctx context.Context, event *TestEvent1) error {
		if counter1_2 == 0 {
			assert.Equal(t, event1, event)
		}
		if counter1_2 == 1 {
			assert.Equal(t, event1_1, event)
		}
		counter1_2++
		return nil
	})
	assert.NoError(t, err)

	counter2_1 := 0
	_, err = Subscribe[*TestEvent2](bus)(ctx, func(ctx context.Context, event *TestEvent2) error {
		assert.Equal(t, event2_1, event)
		counter2_1++
		return nil
	})

	assert.NoError(t, err)

	err = Publish[*TestEvent1](bus)(ctx, event1)
	assert.NoError(t, err)
	err = Publish[*TestEvent1](bus)(ctx, event1_1)
	assert.NoError(t, err)
	err = Publish[*TestEvent2](bus)(ctx, event2_1)
	assert.NoError(t, err)

	assert.Equal(t, counter1, 3)
	assert.Equal(t, counter1_2, 2)
	assert.Equal(t, counter2_1, 1)

	//publish after dispose
	err = dispose1.Dispose(ctx)
	assert.NoError(t, err)
	err = Publish[*TestEvent1](bus)(ctx, event1)
	assert.NoError(t, err)
	assert.Equal(t, counter1, 3)
}

func TestDispatchAndProcess(t *testing.T) {
	bus := New()
	ctx := context.Background()
	event1 := &TestEvent1{}

	result := &TestResult1{}

	counter := 0
	dispose1, err := AddProcessor[*TestEvent1, *TestResult1](bus)(ctx, func(ctx context.Context, event *TestEvent1) (*TestResult1, error) {
		counter++
		return result, nil
	})
	assert.NoError(t, err)

	dispose2, err := AddProcessor[*TestEvent1, *TestResult2](bus)(ctx, func(ctx context.Context, event *TestEvent1) (*TestResult2, error) {
		counter++
		return &TestResult2{}, nil
	})
	assert.NoError(t, err)

	result1, err := Dispatch[*TestEvent1, *TestResult1](bus)(ctx, event1)
	assert.NoError(t, err)
	assert.Equal(t, result, result1)

	assert.Equal(t, 1, counter)

	err = dispose1.Dispose(ctx)
	assert.NoError(t, err)
	result1, err = Dispatch[*TestEvent1, *TestResult1](bus)(ctx, event1)
	assert.Error(t, ErrNotProcessor, err)
	assert.Nil(t, result1)
	assert.Equal(t, 1, counter)

	result2, err := Dispatch[*TestEvent1, *TestResult2](bus)(ctx, event1)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Equal(t, 2, counter)

	dispose2.Dispose(ctx)

	//Any Processor
	//dispose3, err := AddProcessor[interface{}, interface{}](bus)(ctx, func(ctx context.Context, event interface{}) (interface{}, error) {
	//	return &TestResult1{}, err
	//})
	//var resultAny interface{}
	//assert.NoError(t, err)
	//resultAny, err = Dispatch[*TestEvent1, interface{}](bus)(ctx, &TestEvent1{})
	//assert.NoError(t, err)
	//assert.NotNil(t, resultAny)
	//dispose3.Dispose(ctx)
	//
	//resultAny, err = Dispatch[*TestEvent1, interface{}](bus)(ctx, &TestEvent1{})
	//assert.ErrorIs(t, err, ErrNotProcessor)
	//assert.Nil(t, resultAny)
	//bizErr := fmt.Errorf("biz")
	//dispose, err := AddProcessor[interface{}, TestInterface](bus)(ctx, func(ctx context.Context, event interface{}) (TestInterface, error) {
	//	return nil, bizErr
	//})
	//assert.NoError(t, err)
	//_, err = Dispatch[*TestEvent1, interface{}](bus)(ctx, &TestEvent1{})
	//assert.ErrorIs(t, err, ErrNotProcessor)
	//result3, err := Dispatch[*TestEvent1, TestInterface](bus)(ctx, &TestEvent1{})
	//assert.ErrorIs(t, err, bizErr)
	//assert.Nil(t, result3)
	//dispose.Dispose(ctx)

}
