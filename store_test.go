package flux_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/go-flux"
)

func TestStore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cbFnInvoked := false
		d := flux.NewDispatcher()
		var s flux.Store
		cbFn := func(p interface{}) {
			cbFnInvoked = true
			err := s.EmitChange()
			require.NoError(t, err)

			hc, err := s.HasChanged()
			require.NoError(t, err)
			assert.True(t, hc)
		}
		s = flux.NewStore(d, cbFn)

		d.Dispatch("")
		assert.True(t, cbFnInvoked)

		t.Run("BasicDispatcherGetters", func(t *testing.T) {
			assert.Equal(t, d, s.GetDispatcher())
			assert.NotEmpty(t, s.GetDispatcherToken())
		})

		t.Run("ErrRequiresDispatching", func(t *testing.T) {
			err := s.EmitChange()
			assert.EqualError(t, err, flux.ErrRequiresDispatching.Error())

			hc, err := s.HasChanged()
			assert.False(t, hc)
			assert.EqualError(t, err, flux.ErrRequiresDispatching.Error())
		})
	})
	t.Run("SuccessWithListeners", func(t *testing.T) {
		cbFnInvoked := false
		d := flux.NewDispatcher()
		var s flux.Store
		cbFn := func(p interface{}) {
			cbFnInvoked = true
			err := s.EmitChange()
			require.NoError(t, err)
		}
		lFnInvoked := false
		lFn := func() {
			lFnInvoked = true
		}
		lFn2Invoked := false
		lFn2 := func() {
			lFn2Invoked = true
		}

		s = flux.NewStore(d, cbFn)
		rlFn := s.AddListener(lFn)
		_ = s.AddListener(lFn2)

		t.Run("WhenAllAdded", func(t *testing.T) {
			d.Dispatch("")
			assert.True(t, cbFnInvoked)
			assert.True(t, lFnInvoked)
			assert.True(t, lFn2Invoked)
		})

		t.Run("WhenRemovingTheFirstOne", func(t *testing.T) {
			rlFn()

			cbFnInvoked, lFnInvoked, lFn2Invoked = false, false, false
			d.Dispatch("")
			assert.True(t, cbFnInvoked)
			assert.False(t, lFnInvoked)
			assert.True(t, lFn2Invoked)
		})

		t.Run("WhenReaddingAfterDelete", func(t *testing.T) {
			_ = s.AddListener(lFn)
			cbFnInvoked, lFnInvoked, lFn2Invoked = false, false, false
			d.Dispatch("")
			assert.True(t, cbFnInvoked)
			assert.True(t, lFnInvoked)
			assert.True(t, lFn2Invoked)
		})
	})
	t.Run("SuccessWithNoChanges", func(t *testing.T) {
		cbFnInvoked := false
		d := flux.NewDispatcher()
		var s flux.Store
		cbFn := func(p interface{}) {
			cbFnInvoked = true
		}
		lFnInvoked := false
		lFn := func() {
			lFnInvoked = true
		}
		lFn2Invoked := false
		lFn2 := func() {
			lFn2Invoked = true
		}

		s = flux.NewStore(d, cbFn)
		s.AddListener(lFn)
		s.AddListener(lFn2)

		d.Dispatch("")
		assert.True(t, cbFnInvoked)
		assert.False(t, lFnInvoked)
		assert.False(t, lFn2Invoked)
	})
}
