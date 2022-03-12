package flux_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/go-flux"
)

type testState struct {
	Value int
}

func TestReduceStore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		d := flux.NewDispatcher()
		is := testState{}
		eis := testState{Value: 1}
		rFnInvoked := false

		rFn := func(state, action interface{}) interface{} {
			rFnInvoked = true
			ts := state.(testState)
			ts.Value = 1
			return ts
		}

		lFnInvoked := false
		lFn := func() {
			lFnInvoked = true
		}

		rs := flux.NewReduceStore(d, rFn, is)

		rs.AddListener(lFn)

		t.Run("WithChange", func(t *testing.T) {
			assert.Equal(t, is, rs.GetState())
			assert.True(t, rs.AreEqual(is, rs.GetState()))
			assert.False(t, rFnInvoked)
			assert.False(t, lFnInvoked)

			d.Dispatch("")

			assert.Equal(t, eis, rs.GetState())
			assert.True(t, rFnInvoked)
			assert.True(t, lFnInvoked)
		})

		t.Run("WithNoChange", func(t *testing.T) {
			rFnInvoked, lFnInvoked = false, false

			assert.Equal(t, eis, rs.GetState())
			assert.True(t, rs.AreEqual(eis, rs.GetState()))
			assert.False(t, rFnInvoked)
			assert.False(t, lFnInvoked)

			d.Dispatch("")

			assert.Equal(t, eis, rs.GetState())
			assert.True(t, rFnInvoked)
			assert.False(t, lFnInvoked)

		})
	})
	t.Run("OverwriteAreEqual", func(t *testing.T) {
		d := flux.NewDispatcher()
		is := testState{}
		rFnInvoked := false

		rFn := func(state, action interface{}) interface{} {
			rFnInvoked = true
			ts := state.(testState)
			ts.Value = 1
			return ts
		}

		lFnInvoked := false
		lFn := func() {
			lFnInvoked = true
		}

		aeFn := flux.ReduceStoreOptionAreEqual(func(one, two interface{}) bool {
			return true
		})
		rs := flux.NewReduceStore(d, rFn, is, aeFn)

		rs.AddListener(lFn)

		assert.Equal(t, is, rs.GetState())
		assert.False(t, rFnInvoked)
		assert.False(t, lFnInvoked)

		d.Dispatch("")

		assert.Equal(t, is, rs.GetState())
		assert.True(t, rFnInvoked)
		assert.False(t, lFnInvoked)
	})
}
