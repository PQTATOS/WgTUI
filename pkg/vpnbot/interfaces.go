package vpnbot


type Processor interface {
	Process(Event) error
	AddHandler(string, func(Event, Processor)) error
	Start()
}


type Event struct {
	Text string
	Meta interface{}
}