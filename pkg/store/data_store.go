package store

func NewDataStore[T any](data T) DataStoreOwner[T] {
	return &dataStore[T]{
		data:       data,
		notifyChan: make(chan T),
		setReqChan: make(chan SetRequest[T]),
		setResChan: make(chan error),
	}
}

func (s *dataStore[T]) GetData() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (s *dataStore[T]) GetDataUnsafe() T {
	return s.data
}

func (s *dataStore[T]) SetDataUnsafe(data T) {
	s.data = data
}

func (s *dataStore[T]) NotifyChanReceive() <-chan T {
	return s.notifyChan
}

func (s *dataStore[T]) NotifyChanSend() chan<- T {
	return s.notifyChan
}

func (s *dataStore[T]) SetData(data T) error {
	req := SetRequest[T]{
		Data:            data,
		SetResponseChan: s.setResChan,
	}

	s.setReqChan <- req
	// Waits until we get response that set data is ok or not ok
	// This blocks code and forces it to wait until set request finishes
	err := <-s.setResChan
	return err
}
func (s *dataStore[T]) SetReqChanReceive() <-chan SetRequest[T] {
	return s.setReqChan
}

func (s *dataStore[T]) Lock() {
	s.mu.Lock()
}

func (s *dataStore[T]) Unlock() {
	s.mu.Unlock()
}
