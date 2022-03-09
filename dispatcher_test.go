package flux_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/go-flux"
)

func TestDispatcher(t *testing.T) {
	d := flux.NewDispatcher()
	var fnCalled bool
	pl := "some action"

	id := d.Register(func(payload interface{}) {
		fnCalled = true
		assert.Equal(t, pl, payload)
		assert.True(t, d.IsDispatching())

		err := d.Dispatch(payload)
		assert.EqualError(t, err, flux.ErrAlreadyDispatching.Error())
	})

	assert.Equal(t, "1", id)
	assert.False(t, fnCalled)
	assert.False(t, d.IsDispatching())

	err := d.Dispatch(pl)
	require.NoError(t, err)

	assert.True(t, fnCalled)
	assert.False(t, d.IsDispatching())

	err = d.Unregister(id)
	require.NoError(t, err)

	err = d.Unregister(id)
	assert.EqualError(t, err, flux.ErrNotMatchingCallback.Error())

	fnCalled = false

	d.Dispatch(pl)
	assert.False(t, fnCalled)
}

func TestWaitFor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		d := flux.NewDispatcher()

		results := []string{}
		eresults := []string{"3", "2", "1"}

		// The ids used on this examples are deduced
		// as the order of execution of the Register
		// is sequential and the ids have an order,
		// if not we would need to use the result of
		// the d.Register
		d.Register(func(payload interface{}) {
			err := d.WaitFor("2", "3")
			require.NoError(t, err)

			results = append(results, "1")
		})
		d.Register(func(payload interface{}) {
			err := d.WaitFor("3")
			require.NoError(t, err)

			results = append(results, "2")
		})
		d.Register(func(payload interface{}) {
			results = append(results, "3")
		})

		err := d.Dispatch("")
		require.NoError(t, err)

		assert.Equal(t, eresults, results)
	})
	t.Run("ErrWaitForCircularDependency", func(t *testing.T) {
		d := flux.NewDispatcher()

		var wferr1, wferr2 error
		// The ids used on this examples are deduced
		// as the order of execution of the Register
		// is sequential and the ids have an order,
		// if not we would need to use the result of
		// the d.Register
		d.Register(func(payload interface{}) {
			wferr1 = d.WaitFor("2")
		})
		d.Register(func(payload interface{}) {
			wferr2 = d.WaitFor("1")
		})

		err := d.Dispatch("")
		require.NoError(t, err)
		// Depending on the execution order the 1 or the 2 could get the error
		// but not the 2 at the same time
		if wferr1 != nil {
			assert.EqualError(t, wferr1, flux.ErrWaitForCircularDependency.Error())
		} else if wferr2 != nil {
			assert.EqualError(t, wferr2, flux.ErrWaitForCircularDependency.Error())
		} else {
			assert.Fail(t, "Required ErrWaitForCircularDependency")
		}
	})
	t.Run("ErrWaitForDispatching", func(t *testing.T) {
		d := flux.NewDispatcher()

		err := d.WaitFor()
		assert.EqualError(t, err, flux.ErrWaitForDispatching.Error())
	})
	t.Run("ErrWaitForCircularDependency", func(t *testing.T) {
		d := flux.NewDispatcher()

		var wferr error
		d.Register(func(payload interface{}) {
			wferr = d.WaitFor("2")
		})

		err := d.Dispatch("")
		require.NoError(t, err)
		assert.EqualError(t, wferr, flux.ErrNotMatchingCallback.Error())
	})
}
