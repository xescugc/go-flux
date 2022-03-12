package flux

import (
	"reflect"
)

// ReduceFn is the function that would be used to handle dispatched
// actions, it receives the current state and the action and it returns
// a new state applying the action.
type ReduceFn func(state, action interface{}) (newState interface{})

// ReduceAreEqualFn is the function used to compare two
// states and determinate if they are equal or not
type ReduceAreEqualFn func(one, two interface{}) bool

// ReduceStore is the main struct that should be used to extend/compose Stores
type ReduceStore struct {
	*Store

	state interface{}

	reduceFn   ReduceFn
	areEqualFn ReduceAreEqualFn
}

// NewReduceStore will return a new ReduceStore with the given dispatcher that each dispatch wil cal the rFn.
// The first state will be the initialState and the opts can overwrite some of the internal logic
// If after the rFn the state has changed a change even will be triggered to the Listeners if any, the change does not
// have to be set manually
func NewReduceStore(d *Dispatcher, rFn ReduceFn, initialState interface{}, opts ...ReduceStoreOption) *ReduceStore {
	rs := &ReduceStore{
		state:      initialState,
		reduceFn:   rFn,
		areEqualFn: reflect.DeepEqual,
	}

	rs.Store = NewStore(d, rs.invokeOnDispatch)

	for _, opt := range opts {
		opt(rs)
	}

	return rs
}

// GetState returns the current state
func (rs *ReduceStore) GetState() interface{} { return rs.state }

// AreEqual will compare the object one and two to check if they are equal or not.
// It's used internally to check if the state has changed after the reduce
// function has been executed
func (rs *ReduceStore) AreEqual(one, two interface{}) bool {
	return rs.areEqualFn(one, two)
}

func (rs *ReduceStore) invokeOnDispatch(payload interface{}) {
	startingState := rs.state

	endignState := rs.reduceFn(startingState, payload)

	if !rs.AreEqual(startingState, endignState) {
		rs.state = endignState
		rs.Store.EmitChange()
	}
}
