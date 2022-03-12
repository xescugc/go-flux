package flux

// ReduceStoreOption are the options that can be passed to the NewReduceStore
// to change some of the internal logic
type ReduceStoreOption func(rs *ReduceStore)

// ReduceStoreOptionAreEqual will change the way the objects are compared
// to check if they have changed
func ReduceStoreOptionAreEqual(aeFn ReduceAreEqualFn) ReduceStoreOption {
	return func(rs *ReduceStore) {
		rs.areEqualFn = aeFn
	}
}
