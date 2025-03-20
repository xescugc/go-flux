package flux

import (
	"errors"
	"strconv"
	"sync"
)

// List of all the errors
var (
	ErrRequiresDispatching = errors.New("must be invoked while dispatching")
)

// Store represents the basic implementation of a FluxStore
type Store[P any] struct {
	dispatcher      *Dispatcher[P]
	dispatcherToken string

	muListeners sync.RWMutex
	listeners   map[string]func()

	changed bool

	lastID int
}

// NewStore initialized a Store with the dispatcher d and the main
// callback function cbFn that will be the handler of all the
// dispatched payloads.
// If the cbFn needs to trigger a change event it should call
// EmitChange inside of it
func NewStore[P any](d *Dispatcher[P], cbFn CallbackFn[P]) *Store[P] {
	s := &Store[P]{
		dispatcher: d,
		listeners:  make(map[string]func()),
	}
	s.dispatcherToken = d.Register(func(payload P) {
		s.invokeCallbackFn(cbFn, payload)
	})
	return s
}

// AddListener will add a listener fn to the store, when the store
// change the given fn will be called.
// It returns a rmFn to remove the listener from the store
func (s *Store[P]) AddListener(fn func()) func() {
	s.muListeners.Lock()
	defer s.muListeners.Unlock()

	s.lastID++
	id := strconv.Itoa(s.lastID)
	s.listeners[id] = fn
	return func(id string) func() {
		return func() {
			s.muListeners.Lock()
			defer s.muListeners.Unlock()

			delete(s.listeners, id)
		}
	}(id)
}

// GetDispatcher returns the internal dispatcher
func (s *Store[P]) GetDispatcher() *Dispatcher[P] { return s.dispatcher }

// GetDispatcherToken returns the dispatcher token assigned
// to this store
func (s *Store[P]) GetDispatcherToken() string { return s.dispatcherToken }

// HasChanged evaluates if the store has chnaged.
// It can only be called during a dispatch if not
// it returns ErrRequiresDispatching
func (s *Store[P]) HasChanged() (bool, error) {
	if !s.dispatcher.IsDispatching() {
		return false, ErrRequiresDispatching
	}

	return s.changed, nil
}

// EmitChange will notify all the listeners that the store has changed,
// notifications will be done at the end of the dispatch. It can only
// be called during a dispatch if not
// it returns ErrRequiresDispatching
func (s *Store[P]) EmitChange() error {
	if !s.dispatcher.IsDispatching() {
		return ErrRequiresDispatching
	}

	s.changed = true
	return nil
}

func (s *Store[P]) invokeCallbackFn(cbFn CallbackFn[P], payload P) {
	s.changed = false
	cbFn(payload)
	if s.changed {
		s.muListeners.RLock()
		defer s.muListeners.RUnlock()

		for _, fn := range s.listeners {
			fn()
		}
	}
}
