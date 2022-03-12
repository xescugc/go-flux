# go-flux
[![GoDoc](https://godoc.org/github.com/xescugc/go-flux?status.svg)](https://godoc.org/github.com/xescugc/go-flux)

GO implementation/port of the [Flux](https://github.com/facebook/flux) application architecture

## Dispatcher

Dispatcher is used to broadcast payloads to registered callbacks. This is different from generic pub-sub systems in two ways:

* Callbacks are not subscribed to particular events. Every payload is dispatched to every registered callback.
* Callbacks can be deferred in whole or part until other callbacks have been executed.

### Examples

```golang
type CountryStore struct {
  DispatcherToken string
}

type CityStore struct {
  DispatcherToken string
}

// Initialize the Dispatcher
d := flux.NewDispatcher()

// Initialize the Stores
coStore := CountryStore{}
ciStore := CityStore{}

// Register some callbacks to the Dispatcher
coStore.DispatcherToken = d.Register(func(payload interface{}){
  // Do any actions with the payload
})
ciStore.DispatcherToken = d.Register(func(payload interface{}){
  // This will make sure that the
  // callback is executed after the IDs
  // on the WaitFor have already ran the callback 
  // and this will run afer
  d.WaitFor(coStore.DispatcherToken)

  // Do any actions with the payload after
  // he coStore has already dealt with the action
})
```

## Store

Store is an abstraction around a Dispatcher which adds listener functionalities to it

### Examples

```golang
type MyStore struct {
  *flux.Store
}

func NewMyStore(d *Dispatcher) &MyStore {
  ms := &MyStore{}
  s := flux.NewStore(d, ms.OnDispatch)
  ms.Store = s

  return ms
}

// OnDispatch will be called each time the Dispatcher dispatches
// a new action
func (m *MyStore) OnDispatch(payload interface{}) {
  // Do any actions with the payload

  // If I want to notify all the listeners that something has
  // changed I have to use the Store.EmitChange
  m.Storte.EmitChange()
}

d := flux.NewDispatcher()
ms := NewMyStore(d)
rl := ms.AddListener(func() {
  // Will be called when the Store
  // has any new change
})

```

## ReduceStore

Is the main struct to extend/compose any Store, it extends the Store and it has a State
and that State is changed with a reducer and if in that reducer `reduce(state,action) state`
the state has changed it'll automatically trigger a change envent, it's not longer necessary
to manually trigger it.

### Examples

```golang
type MyStore struct {
  *flux.ReduceStore
}

type State struct {
  Value int
}

func NewMyStore(d *Dispatcher) &MyStore {
  ms := &MyStore{}
  rs := flux.NewReduceStore(d, ms.Reduce, State{})
  ms.ReduceStore = rs

  return ms
}

// Reduce will be called each time the Dispatcher dispatches
// a new action and if the staet is changed all the listeners
// will be notified of the change
func (m *MyStore) Reduce(state, payload interface{}) interface{}{
  // Do any actions with the payload
}

d := flux.NewDispatcher()
ms := NewMyStore(d)
rl := ms.AddListener(func() {
  // Will be called when the Store
  // has any new change
})
```
