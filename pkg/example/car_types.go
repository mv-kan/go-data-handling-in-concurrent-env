package example

type carDevice interface {
	IsConnected() bool
	GetSpeed() float64
	SetSpeed(float64) error
	GetLocation() float64
}

// implements carDevice interface
type carDeviceS struct {
	data   carData
	status carStatus
}

type carData struct {
	Speed    float64
	Location float64
}

type carStatus struct {
	IsConnected bool
}
