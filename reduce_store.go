package flux

import (
	"reflect"
)

// ReduceFn is the function that would be used to handle dispatched
// actions, it receives the current state and the action and it returns
// a new state applying the action.
type ReduceFn[S, P any] func(state S, payload P) (newState S)

// ReduceAreEqualFn is the function used to compare two
// states and determinate if they are equal or not
type ReduceAreEqualFn[S any] func(one, two S) bool

// ReduceStore is the main struct that should be used to extend/compose Stores
type ReduceStore[S, P any] struct {
	*Store[P]

	state S

	reduceFn   ReduceFn[S, P]
	areEqualFn ReduceAreEqualFn[S]
}

func defaultEqualFn[S any](one, two S) bool {
	return reflect.DeepEqual(one, two)
}

// NewReduceStore will return a new ReduceStore with the given dispatcher that each dispatch will cal the rFn.
// The first state will be the initialState and the opts can overwrite some of the internal logic
// If after the rFn the state has changed a change even will be triggered to the Listeners if any, the change does not
// have to be set manually
func NewReduceStore[S, P any](d *Dispatcher[P], rFn ReduceFn[S, P], initialState S, opts ...ReduceStoreOption[S, P]) *ReduceStore[S, P] {
	rs := &ReduceStore[S, P]{
		state:      initialState,
		reduceFn:   rFn,
		areEqualFn: defaultEqualFn[S],
	}

	rs.Store = NewStore(d, rs.invokeOnDispatch)

	for _, opt := range opts {
		opt(rs)
	}

	return rs
}

// GetState returns the current state
// BUG(xescugc) This is not a copy but the full state, this means that any changes will be
// persisted and that this may cause concurrency issues from the caller.
func (rs *ReduceStore[S, P]) GetState() S { return rs.state }

// AreEqual will compare the object one and two to check if they are equal or not.
// It's used internally to check if the state has changed after the reduce
// function has been executed
func (rs *ReduceStore[S, P]) AreEqual(one, two S) bool {
	return rs.areEqualFn(one, two)
}

func (rs *ReduceStore[S, P]) invokeOnDispatch(payload P) {
	startingState := rs.state

	endignState := rs.reduceFn(startingState, payload)

	if !rs.AreEqual(startingState, endignState) {
		rs.state = endignState
		rs.Store.EmitChange()
	}
}
