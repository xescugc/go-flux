package flux

import (
	"errors"
	"strconv"
	"sync"
)

// List of all the errors
var (
	ErrNotMatchingCallback       = errors.New("does not map to a registered callback")
	ErrAlreadyDispatching        = errors.New("cannot dispatch in the middle of a dispatch")
	ErrWaitForDispatching        = errors.New("wait for must be invoked while dispatching")
	ErrWaitForCircularDependency = errors.New("circular dependency detected")
)

// CallbackFn is the expected type of the callback functions
type CallbackFn func(payload interface{})

// Dispatcher is used to broadcast payloads to registered callbacks.
type Dispatcher interface {
	// Register register the cbFn so it will be
	// executed when Dispatch is executed
	// The returned ID is the internal ID of that function and can be
	// used to WaitFor or Unregister a fn
	Register(cbFn CallbackFn) string

	// Unregister removes the internal cbFn that has
	// assigned id. If the id is not registered
	// it'll return ErrNotMatchingCallback
	Unregister(id string) error

	// WaitFor will wait for all the ids on the list to be
	// executed before running the callback it belongs to.
	// If one of the ids is not registered a ErrNotMatchingCallback
	// will be returned
	// If there is a circular dependency an ErrNotMatchingCallback
	// will be returned
	// If it's called while not Dispatching an ErrWaitForDispatching
	// will be returned
	WaitFor(ids ...string) error

	// Dispatch will send the payload to all the Registered
	// callbacks.
	// If we are already dispatching an ErrAlreadyDispatching
	// will be returned
	Dispatch(payload interface{}) error

	// IsDispatching checks if the Dispatcher is doing any work
	IsDispatching() bool
}

// dispatcher is the  internal implementation of the Dispatcher interface
type dispatcher struct {
	muCallbacks sync.RWMutex
	// callbacks is a list of all
	// the callbacks registered
	callbacks map[string]CallbackFn

	// isHandled keeps track of the callbacks
	// that this already run
	isHandled map[string]struct{}

	// isPending keeps track of all the callbacks
	// that did start running
	isPending map[string]struct{}

	// pendingPayload is the general payload sent
	pendingPayload interface{}

	lastID int

	dispatching bool
}

// NewDispatcher returns a new Dispatcher implementation
func NewDispatcher() Dispatcher {
	return &dispatcher{
		callbacks: make(map[string]CallbackFn),
		isHandled: make(map[string]struct{}),
		isPending: make(map[string]struct{}),
	}
}

func (d *dispatcher) Register(fn CallbackFn) string {
	d.muCallbacks.Lock()
	defer d.muCallbacks.Unlock()

	d.lastID++
	id := strconv.Itoa(d.lastID)
	d.callbacks[id] = fn
	return id
}

func (d *dispatcher) Unregister(id string) error {
	d.muCallbacks.Lock()
	defer d.muCallbacks.Unlock()

	if _, ok := d.callbacks[id]; !ok {
		return ErrNotMatchingCallback
	}

	delete(d.callbacks, id)

	return nil
}

func (d *dispatcher) WaitFor(ids ...string) error {
	if !d.dispatching {
		return ErrWaitForDispatching
	}
	for _, id := range ids {
		if _, ok := d.isPending[id]; ok {
			if _, ok := d.isHandled[id]; ok {
				continue
			}
			return ErrWaitForCircularDependency
		}

		if _, ok := d.callbacks[id]; !ok {
			return ErrNotMatchingCallback
		}

		d.invokeCallback(id)
	}
	return nil
}

func (d *dispatcher) Dispatch(payload interface{}) error {
	d.muCallbacks.RLock()
	defer d.muCallbacks.RUnlock()

	if d.dispatching {
		return ErrAlreadyDispatching
	}

	d.startDispatching(payload)
	defer d.stopDispatching()

	for id := range d.callbacks {
		if _, ok := d.isPending[id]; ok {
			continue
		}
		d.invokeCallback(id)
	}

	return nil
}

func (d *dispatcher) IsDispatching() bool { return d.dispatching }

func (d *dispatcher) invokeCallback(id string) {
	d.isPending[id] = struct{}{}
	d.callbacks[id](d.pendingPayload)
	d.isHandled[id] = struct{}{}
}

func (d *dispatcher) startDispatching(payload interface{}) {
	d.dispatching = true
	d.isHandled = make(map[string]struct{})
	d.isPending = make(map[string]struct{})
	d.pendingPayload = payload
}

func (d *dispatcher) stopDispatching() {
	d.dispatching = false
	d.pendingPayload = nil
}
