package example

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mv-kan/go-data-handling-in-concurrent-env/pkg/store"
)

const (
	INFO                = "INFO"
	WARNING             = "WARNING"
	ERROR               = "ERROR"
	increaseSpeedPeriod = time.Second * 5
)

var (
	ErrConnection = errors.New("connection error")
)

func RunCarExample(
	logger *log.Logger,
	done <-chan struct{},
) error {
	carDev := newCarDev()

	carDataStore, carStatusStore := fireupCarCommunication(
		logger,
		carDev,
		done,
	)
	carDataUpdateChan := make(chan carData)
	carDataUpdateChan2 := make(chan carData)
	go store.Fanout(done, carDataStore.NotifyChanReceive(), carDataUpdateChan, carDataUpdateChan2)
	// just read from chan and throw out the values
	go store.Fanout(done, carStatusStore.NotifyChanReceive())

	fireupIncreaseSpeed(
		logger,
		carDataStore,
		done,
	)

	fireupPrintOutData(
		logger,
		carDataUpdateChan,
		done,
	)

	err := fireupWriteDataIntoFile(
		logger,
		carDataUpdateChan2,
		done,
	)
	if err != nil {
		return err
	}
	return nil
}

func fireupCarCommunication(
	logger *log.Logger,
	dev carDevice,
	done <-chan struct{},
) (store.DataStore[carData], store.DataStoreGetNotif[carStatus]) {
	carDataStore := store.NewDataStore[carData](carData{})
	carStatusStore := store.NewDataStore[carStatus](carStatus{})

	go func() {
		logPrefix := "fireupCarCommunication"
		ticker := time.NewTicker(time.Second * 2)

		getDataFromDevice := func(dev carDevice) carData {
			data := carData{
				Location: dev.GetLocation(),
				Speed:    dev.GetSpeed(),
			}
			return data
		}
		getStatusFromDevice := func(dev carDevice) carStatus {
			data := carStatus{
				IsConnected: dev.IsConnected(),
			}
			return data
		}
		for {
			if !carStatusStore.GetDataUnsafe().IsConnected {
				carStatusStore.Lock()
				carStatusStore.SetDataUnsafe(getStatusFromDevice(dev))
				carStatusStore.Unlock()
				carStatusStore.NotifyChanSend() <- carStatusStore.GetDataUnsafe()
				time.Sleep(time.Second)
				logger.Printf("%s %s: car status is disconnected, wait until connected", logPrefix, ERROR)
				continue
			}
			select {
			case <-done:
				logger.Printf("%s %s: got done signal, exit", logPrefix, INFO)
				return
			case setReq := <-carDataStore.SetReqChanReceive():
				carDataStore.Lock()
				carStatusStore.Lock()
				logger.Printf("%s %s: locked data store, status store", logPrefix, INFO)
				ok := true
				// no checks for negative speed, because it is not production code lol
				err := dev.SetSpeed(setReq.Data.Speed)
				if err != nil {
					ok = false
					carStatusStore.SetDataUnsafe(carStatus{IsConnected: false})
				}
				carDataStore.SetDataUnsafe(getDataFromDevice(dev))
				carDataStore.Unlock()
				carStatusStore.Unlock()
				setReq.SetResponseChan <- err
				logger.Printf("%s %s: unlocked data store, status store", logPrefix, INFO)
				if !ok {
					carStatusStore.NotifyChanSend() <- carStatusStore.GetDataUnsafe()
				} else {
					carDataStore.NotifyChanSend() <- carDataStore.GetDataUnsafe()
				}
			case <-ticker.C:
				carDataStore.Lock()
				carStatusStore.Lock()
				// Inside this lock unlock block
				// NEVER EVER send data to channel and other goroutine communication shenanigans
				// WHY? it is really easy to deadlock code like that, but exceptions always exist
				carDataStore.SetDataUnsafe(getDataFromDevice(dev))
				carDataStore.Unlock()
				carStatusStore.Unlock()

				carDataStore.NotifyChanSend() <- carDataStore.GetDataUnsafe()
			}
		}
	}()

	return carDataStore, carStatusStore
}

func fireupIncreaseSpeed(
	logger *log.Logger,
	carDataStore store.DataStoreGetSet[carData],
	done <-chan struct{},
) {
	go func() {
		logPrefix := "fireupIncreaseSpeed"

		ticker := time.NewTicker(increaseSpeedPeriod)
		for {
			select {
			case <-done:
				logger.Printf("%s %s: got done signal, exit", logPrefix, INFO)
				return
			case <-ticker.C:
				speed := carDataStore.GetData().Speed
				carDataStore.SetData(carData{
					Speed: speed + 1,
				})
			}
		}
	}()
}

func fireupPrintOutData(
	logger *log.Logger,
	carDataUpdateChan <-chan carData,
	done <-chan struct{},
) {
	go func() {
		logPrefix := "fireupPrintOutData"
		for {
			select {
			case <-done:
				logger.Printf("%s %s: got done signal, exit", logPrefix, INFO)
				return
			case data := <-carDataUpdateChan:
				fmt.Printf("Speed: %f, Location: %f\n", data.Speed, data.Location)
			}
		}
	}()
}

func fireupWriteDataIntoFile(
	logger *log.Logger,
	carDataUpdateChan <-chan carData,
	done <-chan struct{},
) error {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	go func() {
		logPrefix := "fireupPrintOutData"
		defer file.Close()
		for {
			select {
			case <-done:
				logger.Printf("%s %s: got done signal, exit", logPrefix, INFO)
				return
			case data := <-carDataUpdateChan:
				currentTime := time.Now()
				strDate := currentTime.Format("2006-01-02 15:04:05")
				str := fmt.Sprintf("Date: %s, Speed: %f, Location: %f\n", strDate, data.Speed, data.Location)
				_, err := file.WriteString(str)
				if err != nil {
					logger.Printf("%s %s: got error=%v", logPrefix, ERROR, err)
				}
			}
		}
	}()
	return nil
}

func newCarDev() carDevice {
	return &carDeviceS{
		status: carStatus{IsConnected: true},
	}
}

func (c *carDeviceS) IsConnected() bool {
	return c.status.IsConnected
}

func (c *carDeviceS) GetSpeed() float64 {
	return c.data.Speed
}

func (c *carDeviceS) SetSpeed(v float64) error {
	c.data.Speed = v
	return nil
}

func (c *carDeviceS) GetLocation() float64 {
	return c.data.Speed * 2
}
