// Code generated by http://github.com/gojuno/minimock (v3.4.0). DO NOT EDIT.

package mock

//go:generate minimock -i route256/loms/internal/service/loms.ITxManager -o transaction_manager_mock.go -n ITxManagerMock -p mock

import (
	"context"
	mm_service "route256/loms/internal/service/loms"
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// ITxManagerMock implements mm_service.ITxManager
type ITxManagerMock struct {
	t          minimock.Tester
	finishOnce sync.Once

	funcWithTx          func(ctx context.Context, fn mm_service.WithTxFunc) (err error)
	funcWithTxOrigin    string
	inspectFuncWithTx   func(ctx context.Context, fn mm_service.WithTxFunc)
	afterWithTxCounter  uint64
	beforeWithTxCounter uint64
	WithTxMock          mITxManagerMockWithTx
}

// NewITxManagerMock returns a mock for mm_service.ITxManager
func NewITxManagerMock(t minimock.Tester) *ITxManagerMock {
	m := &ITxManagerMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.WithTxMock = mITxManagerMockWithTx{mock: m}
	m.WithTxMock.callArgs = []*ITxManagerMockWithTxParams{}

	t.Cleanup(m.MinimockFinish)

	return m
}

type mITxManagerMockWithTx struct {
	optional           bool
	mock               *ITxManagerMock
	defaultExpectation *ITxManagerMockWithTxExpectation
	expectations       []*ITxManagerMockWithTxExpectation

	callArgs []*ITxManagerMockWithTxParams
	mutex    sync.RWMutex

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// ITxManagerMockWithTxExpectation specifies expectation struct of the ITxManager.WithTx
type ITxManagerMockWithTxExpectation struct {
	mock               *ITxManagerMock
	params             *ITxManagerMockWithTxParams
	paramPtrs          *ITxManagerMockWithTxParamPtrs
	expectationOrigins ITxManagerMockWithTxExpectationOrigins
	results            *ITxManagerMockWithTxResults
	returnOrigin       string
	Counter            uint64
}

// ITxManagerMockWithTxParams contains parameters of the ITxManager.WithTx
type ITxManagerMockWithTxParams struct {
	ctx context.Context
	fn  mm_service.WithTxFunc
}

// ITxManagerMockWithTxParamPtrs contains pointers to parameters of the ITxManager.WithTx
type ITxManagerMockWithTxParamPtrs struct {
	ctx *context.Context
	fn  *mm_service.WithTxFunc
}

// ITxManagerMockWithTxResults contains results of the ITxManager.WithTx
type ITxManagerMockWithTxResults struct {
	err error
}

// ITxManagerMockWithTxOrigins contains origins of expectations of the ITxManager.WithTx
type ITxManagerMockWithTxExpectationOrigins struct {
	origin    string
	originCtx string
	originFn  string
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmWithTx *mITxManagerMockWithTx) Optional() *mITxManagerMockWithTx {
	mmWithTx.optional = true
	return mmWithTx
}

// Expect sets up expected params for ITxManager.WithTx
func (mmWithTx *mITxManagerMockWithTx) Expect(ctx context.Context, fn mm_service.WithTxFunc) *mITxManagerMockWithTx {
	if mmWithTx.mock.funcWithTx != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by Set")
	}

	if mmWithTx.defaultExpectation == nil {
		mmWithTx.defaultExpectation = &ITxManagerMockWithTxExpectation{}
	}

	if mmWithTx.defaultExpectation.paramPtrs != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by ExpectParams functions")
	}

	mmWithTx.defaultExpectation.params = &ITxManagerMockWithTxParams{ctx, fn}
	mmWithTx.defaultExpectation.expectationOrigins.origin = minimock.CallerInfo(1)
	for _, e := range mmWithTx.expectations {
		if minimock.Equal(e.params, mmWithTx.defaultExpectation.params) {
			mmWithTx.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmWithTx.defaultExpectation.params)
		}
	}

	return mmWithTx
}

// ExpectCtxParam1 sets up expected param ctx for ITxManager.WithTx
func (mmWithTx *mITxManagerMockWithTx) ExpectCtxParam1(ctx context.Context) *mITxManagerMockWithTx {
	if mmWithTx.mock.funcWithTx != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by Set")
	}

	if mmWithTx.defaultExpectation == nil {
		mmWithTx.defaultExpectation = &ITxManagerMockWithTxExpectation{}
	}

	if mmWithTx.defaultExpectation.params != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by Expect")
	}

	if mmWithTx.defaultExpectation.paramPtrs == nil {
		mmWithTx.defaultExpectation.paramPtrs = &ITxManagerMockWithTxParamPtrs{}
	}
	mmWithTx.defaultExpectation.paramPtrs.ctx = &ctx
	mmWithTx.defaultExpectation.expectationOrigins.originCtx = minimock.CallerInfo(1)

	return mmWithTx
}

// ExpectFnParam2 sets up expected param fn for ITxManager.WithTx
func (mmWithTx *mITxManagerMockWithTx) ExpectFnParam2(fn mm_service.WithTxFunc) *mITxManagerMockWithTx {
	if mmWithTx.mock.funcWithTx != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by Set")
	}

	if mmWithTx.defaultExpectation == nil {
		mmWithTx.defaultExpectation = &ITxManagerMockWithTxExpectation{}
	}

	if mmWithTx.defaultExpectation.params != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by Expect")
	}

	if mmWithTx.defaultExpectation.paramPtrs == nil {
		mmWithTx.defaultExpectation.paramPtrs = &ITxManagerMockWithTxParamPtrs{}
	}
	mmWithTx.defaultExpectation.paramPtrs.fn = &fn
	mmWithTx.defaultExpectation.expectationOrigins.originFn = minimock.CallerInfo(1)

	return mmWithTx
}

// Inspect accepts an inspector function that has same arguments as the ITxManager.WithTx
func (mmWithTx *mITxManagerMockWithTx) Inspect(f func(ctx context.Context, fn mm_service.WithTxFunc)) *mITxManagerMockWithTx {
	if mmWithTx.mock.inspectFuncWithTx != nil {
		mmWithTx.mock.t.Fatalf("Inspect function is already set for ITxManagerMock.WithTx")
	}

	mmWithTx.mock.inspectFuncWithTx = f

	return mmWithTx
}

// Return sets up results that will be returned by ITxManager.WithTx
func (mmWithTx *mITxManagerMockWithTx) Return(err error) *ITxManagerMock {
	if mmWithTx.mock.funcWithTx != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by Set")
	}

	if mmWithTx.defaultExpectation == nil {
		mmWithTx.defaultExpectation = &ITxManagerMockWithTxExpectation{mock: mmWithTx.mock}
	}
	mmWithTx.defaultExpectation.results = &ITxManagerMockWithTxResults{err}
	mmWithTx.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmWithTx.mock
}

// Set uses given function f to mock the ITxManager.WithTx method
func (mmWithTx *mITxManagerMockWithTx) Set(f func(ctx context.Context, fn mm_service.WithTxFunc) (err error)) *ITxManagerMock {
	if mmWithTx.defaultExpectation != nil {
		mmWithTx.mock.t.Fatalf("Default expectation is already set for the ITxManager.WithTx method")
	}

	if len(mmWithTx.expectations) > 0 {
		mmWithTx.mock.t.Fatalf("Some expectations are already set for the ITxManager.WithTx method")
	}

	mmWithTx.mock.funcWithTx = f
	mmWithTx.mock.funcWithTxOrigin = minimock.CallerInfo(1)
	return mmWithTx.mock
}

// When sets expectation for the ITxManager.WithTx which will trigger the result defined by the following
// Then helper
func (mmWithTx *mITxManagerMockWithTx) When(ctx context.Context, fn mm_service.WithTxFunc) *ITxManagerMockWithTxExpectation {
	if mmWithTx.mock.funcWithTx != nil {
		mmWithTx.mock.t.Fatalf("ITxManagerMock.WithTx mock is already set by Set")
	}

	expectation := &ITxManagerMockWithTxExpectation{
		mock:               mmWithTx.mock,
		params:             &ITxManagerMockWithTxParams{ctx, fn},
		expectationOrigins: ITxManagerMockWithTxExpectationOrigins{origin: minimock.CallerInfo(1)},
	}
	mmWithTx.expectations = append(mmWithTx.expectations, expectation)
	return expectation
}

// Then sets up ITxManager.WithTx return parameters for the expectation previously defined by the When method
func (e *ITxManagerMockWithTxExpectation) Then(err error) *ITxManagerMock {
	e.results = &ITxManagerMockWithTxResults{err}
	return e.mock
}

// Times sets number of times ITxManager.WithTx should be invoked
func (mmWithTx *mITxManagerMockWithTx) Times(n uint64) *mITxManagerMockWithTx {
	if n == 0 {
		mmWithTx.mock.t.Fatalf("Times of ITxManagerMock.WithTx mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmWithTx.expectedInvocations, n)
	mmWithTx.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmWithTx
}

func (mmWithTx *mITxManagerMockWithTx) invocationsDone() bool {
	if len(mmWithTx.expectations) == 0 && mmWithTx.defaultExpectation == nil && mmWithTx.mock.funcWithTx == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmWithTx.mock.afterWithTxCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmWithTx.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// WithTx implements mm_service.ITxManager
func (mmWithTx *ITxManagerMock) WithTx(ctx context.Context, fn mm_service.WithTxFunc) (err error) {
	mm_atomic.AddUint64(&mmWithTx.beforeWithTxCounter, 1)
	defer mm_atomic.AddUint64(&mmWithTx.afterWithTxCounter, 1)

	mmWithTx.t.Helper()

	if mmWithTx.inspectFuncWithTx != nil {
		mmWithTx.inspectFuncWithTx(ctx, fn)
	}

	mm_params := ITxManagerMockWithTxParams{ctx, fn}

	// Record call args
	mmWithTx.WithTxMock.mutex.Lock()
	mmWithTx.WithTxMock.callArgs = append(mmWithTx.WithTxMock.callArgs, &mm_params)
	mmWithTx.WithTxMock.mutex.Unlock()

	for _, e := range mmWithTx.WithTxMock.expectations {
		if minimock.Equal(*e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.err
		}
	}

	if mmWithTx.WithTxMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmWithTx.WithTxMock.defaultExpectation.Counter, 1)
		mm_want := mmWithTx.WithTxMock.defaultExpectation.params
		mm_want_ptrs := mmWithTx.WithTxMock.defaultExpectation.paramPtrs

		mm_got := ITxManagerMockWithTxParams{ctx, fn}

		if mm_want_ptrs != nil {

			if mm_want_ptrs.ctx != nil && !minimock.Equal(*mm_want_ptrs.ctx, mm_got.ctx) {
				mmWithTx.t.Errorf("ITxManagerMock.WithTx got unexpected parameter ctx, expected at\n%s:\nwant: %#v\n got: %#v%s\n",
					mmWithTx.WithTxMock.defaultExpectation.expectationOrigins.originCtx, *mm_want_ptrs.ctx, mm_got.ctx, minimock.Diff(*mm_want_ptrs.ctx, mm_got.ctx))
			}

			if mm_want_ptrs.fn != nil && !minimock.Equal(*mm_want_ptrs.fn, mm_got.fn) {
				mmWithTx.t.Errorf("ITxManagerMock.WithTx got unexpected parameter fn, expected at\n%s:\nwant: %#v\n got: %#v%s\n",
					mmWithTx.WithTxMock.defaultExpectation.expectationOrigins.originFn, *mm_want_ptrs.fn, mm_got.fn, minimock.Diff(*mm_want_ptrs.fn, mm_got.fn))
			}

		} else if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmWithTx.t.Errorf("ITxManagerMock.WithTx got unexpected parameters, expected at\n%s:\nwant: %#v\n got: %#v%s\n",
				mmWithTx.WithTxMock.defaultExpectation.expectationOrigins.origin, *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmWithTx.WithTxMock.defaultExpectation.results
		if mm_results == nil {
			mmWithTx.t.Fatal("No results are set for the ITxManagerMock.WithTx")
		}
		return (*mm_results).err
	}
	if mmWithTx.funcWithTx != nil {
		return mmWithTx.funcWithTx(ctx, fn)
	}
	mmWithTx.t.Fatalf("Unexpected call to ITxManagerMock.WithTx. %v %v", ctx, fn)
	return
}

// WithTxAfterCounter returns a count of finished ITxManagerMock.WithTx invocations
func (mmWithTx *ITxManagerMock) WithTxAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmWithTx.afterWithTxCounter)
}

// WithTxBeforeCounter returns a count of ITxManagerMock.WithTx invocations
func (mmWithTx *ITxManagerMock) WithTxBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmWithTx.beforeWithTxCounter)
}

// Calls returns a list of arguments used in each call to ITxManagerMock.WithTx.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmWithTx *mITxManagerMockWithTx) Calls() []*ITxManagerMockWithTxParams {
	mmWithTx.mutex.RLock()

	argCopy := make([]*ITxManagerMockWithTxParams, len(mmWithTx.callArgs))
	copy(argCopy, mmWithTx.callArgs)

	mmWithTx.mutex.RUnlock()

	return argCopy
}

// MinimockWithTxDone returns true if the count of the WithTx invocations corresponds
// the number of defined expectations
func (m *ITxManagerMock) MinimockWithTxDone() bool {
	if m.WithTxMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.WithTxMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.WithTxMock.invocationsDone()
}

// MinimockWithTxInspect logs each unmet expectation
func (m *ITxManagerMock) MinimockWithTxInspect() {
	for _, e := range m.WithTxMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to ITxManagerMock.WithTx at\n%s with params: %#v", e.expectationOrigins.origin, *e.params)
		}
	}

	afterWithTxCounter := mm_atomic.LoadUint64(&m.afterWithTxCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.WithTxMock.defaultExpectation != nil && afterWithTxCounter < 1 {
		if m.WithTxMock.defaultExpectation.params == nil {
			m.t.Errorf("Expected call to ITxManagerMock.WithTx at\n%s", m.WithTxMock.defaultExpectation.returnOrigin)
		} else {
			m.t.Errorf("Expected call to ITxManagerMock.WithTx at\n%s with params: %#v", m.WithTxMock.defaultExpectation.expectationOrigins.origin, *m.WithTxMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcWithTx != nil && afterWithTxCounter < 1 {
		m.t.Errorf("Expected call to ITxManagerMock.WithTx at\n%s", m.funcWithTxOrigin)
	}

	if !m.WithTxMock.invocationsDone() && afterWithTxCounter > 0 {
		m.t.Errorf("Expected %d calls to ITxManagerMock.WithTx at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.WithTxMock.expectedInvocations), m.WithTxMock.expectedInvocationsOrigin, afterWithTxCounter)
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *ITxManagerMock) MinimockFinish() {
	m.finishOnce.Do(func() {
		if !m.minimockDone() {
			m.MinimockWithTxInspect()
		}
	})
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *ITxManagerMock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *ITxManagerMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockWithTxDone()
}