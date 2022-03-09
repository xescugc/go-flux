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

