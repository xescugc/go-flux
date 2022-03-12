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
type Store struct {
	dispatcher      *Dispatcher
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
func NewStore(d *Dispatcher, cbFn CallbackFn) *Store {
	s := &Store{
		dispatcher: d,
		listeners:  make(map[string]func()),
	}
	s.dispatcherToken = d.Register(func(payload interface{}) {
		s.invokeCallbackFn(cbFn, payload)
	})
	return s
}

// AddListener will add a listener fn to the store, when the store
// change the given fn will be called.
// It returns a rmFn to remove the listener from the store
func (s *Store) AddListener(fn func()) func() {
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
func (s *Store) GetDispatcher() *Dispatcher { return s.dispatcher }

// GetDispatcherToken returns the dispatcher token assigned
// to this store
func (s *Store) GetDispatcherToken() string { return s.dispatcherToken }

// HasChanged evaluates if the store has chnaged.
// It can only be called during a dispatch if not
// it returns ErrRequiresDispatching
func (s *Store) HasChanged() (bool, error) {
	if !s.dispatcher.IsDispatching() {
		return false, ErrRequiresDispatching
	}

	return s.changed, nil
}

// EmitChange will notify all the listeners that the store has changed,
// notifications will be done at the end of the dispatch. It can only
// be called during a dispatch if not
// it returns ErrRequiresDispatching
func (s *Store) EmitChange() error {
	if !s.dispatcher.IsDispatching() {
		return ErrRequiresDispatching
	}

	s.changed = true
	return nil
}

func (s *Store) invokeCallbackFn(cbFn CallbackFn, payload interface{}) {
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
