package common

var PointMultiple float64 = 100000000

// Alert from the Factomd monitor
type FDStatus struct {
	Minute int64
	Dbht   int32
}
