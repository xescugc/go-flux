package flux

// ReduceStoreOption are the options that can be passed to the NewReduceStore
// to change some of the internal logic
type ReduceStoreOption[S, P any] func(rs *ReduceStore[S, P])

// ReduceStoreOptionAreEqual will change the way the objects are compared
// to check if they have changed
func ReduceStoreOptionAreEqual[S, P any](aeFn ReduceAreEqualFn[S]) ReduceStoreOption[S, P] {
	return func(rs *ReduceStore[S, P]) {
		rs.areEqualFn = aeFn
	}
}
