package store

import "sync"

type DataStoreGetter[T any] interface {
	GetData() T
}

type DataStoreSetter[T any] interface {
	SetData(T) error
}

/*
This interface for consumer go routine
IMPORTANT if you are not using notifier you still HAVE TO create some go routine
that continuously reads from notify channel, otherwise we will get blocked
*/
type DataStoreNotifier[T any] interface {
	NotifyChanReceive() <-chan T
}

type DataStoreGetSet[T any] interface {
	DataStoreGetter[T]
	DataStoreSetter[T]
}

type DataStoreGetNotif[T any] interface {
	DataStoreGetter[T]
	DataStoreNotifier[T]
}

type DataStore[T any] interface {
	DataStoreGetter[T]
	DataStoreSetter[T]
	DataStoreNotifier[T]
}

type DataStoreOwner[T any] interface {
	DataStoreGetter[T]
	DataStoreSetter[T]
	DataStoreNotifier[T]

	GetDataUnsafe() T
	SetDataUnsafe(T)

	SetReqChanReceive() <-chan SetRequest[T]
	NotifyChanSend() chan<- T

	Lock()
	Unlock()
}

type SetRequest[T any] struct {
	Data            T
	SetResponseChan chan<- error
}

type dataStore[T any] struct {
	mu         sync.Mutex
	data       T
	notifyChan chan T
	setReqChan chan SetRequest[T]
	setResChan chan error
}

// VERY IMPORTANT for notifier functionality
func Fanout[T any](done <-chan struct{}, in <-chan T, out ...chan<- T) {
	for {
		select {
		case <-done:
			return
		case v := <-in:
			for _, ch := range out {
				ch <- v
			}
		}
	}
}
