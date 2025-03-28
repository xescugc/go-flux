package flux_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/go-flux/v2"
)

func TestDispatcher(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		d := flux.NewDispatcher[string]()
		var fnCalled bool
		pl := "some action"

		id := d.Register(func(payload string) {
			fnCalled = true
			assert.Equal(t, pl, payload)
			assert.True(t, d.IsDispatching())

			go func() {
				err := d.Dispatch(payload)
				require.NoError(t, err)
			}()
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
	})

	t.Run("RegisteringAndUnregisteringMultipleFunctions", func(t *testing.T) {
		calls := make([]int, 0, 0)
		cbFn := func(n int) flux.CallbackFn[string] {
			return func(payload string) {
				calls = append(calls, n)
			}
		}

		d := flux.NewDispatcher[string]()
		d.Register(cbFn(1))
		id2 := d.Register(cbFn(2))
		d.Register(cbFn(3))
		d.Dispatch("")
		sort.Ints(calls)

		assert.Equal(t, []int{1, 2, 3}, calls)
		calls = make([]int, 0, 0)

		d.Unregister(id2)
		d.Register(cbFn(4))
		d.Dispatch("")
		sort.Ints(calls)

		assert.Equal(t, []int{1, 3, 4}, calls)
	})
}

func TestWaitFor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		d := flux.NewDispatcher[string]()

		results := []string{}
		eresults := []string{"3", "2", "1"}

		// The ids used on this examples are deduced
		// as the order of execution of the Register
		// is sequential and the ids have an order,
		// if not we would need to use the result of
		// the d.Register
		d.Register(func(payload string) {
			err := d.WaitFor("2", "3")
			require.NoError(t, err)

			results = append(results, "1")
		})
		d.Register(func(payload string) {
			err := d.WaitFor("3")
			require.NoError(t, err)

			results = append(results, "2")
		})
		d.Register(func(payload string) {
			results = append(results, "3")
		})

		err := d.Dispatch("")
		require.NoError(t, err)

		assert.Equal(t, eresults, results)
	})
	t.Run("ErrWaitForCircularDependency", func(t *testing.T) {
		d := flux.NewDispatcher[string]()

		var wferr1, wferr2 error
		// The ids used on this examples are deduced
		// as the order of execution of the Register
		// is sequential and the ids have an order,
		// if not we would need to use the result of
		// the d.Register
		d.Register(func(payload string) {
			wferr1 = d.WaitFor("2")
		})
		d.Register(func(payload string) {
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
		d := flux.NewDispatcher[string]()

		err := d.WaitFor()
		assert.EqualError(t, err, flux.ErrWaitForDispatching.Error())
	})
	t.Run("ErrWaitForCircularDependency", func(t *testing.T) {
		d := flux.NewDispatcher[string]()

		var wferr error
		d.Register(func(payload string) {
			wferr = d.WaitFor("2")
		})

		err := d.Dispatch("")
		require.NoError(t, err)
		assert.EqualError(t, wferr, flux.ErrNotMatchingCallback.Error())
	})
}
