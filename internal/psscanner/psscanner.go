package psscanner

type PSScanner struct{}

func NewPSScanner() *PSScanner {
	return &PSScanner{}
}

func (p *PSScanner) Run(triggerCh chan struct{}) (chan string, chan error) {
	eventCh := make(chan string, 100)
	errCh := make(chan error)
	pl := make(procList)

	go func() {
		for {
			<-triggerCh
			pl.refresh(eventCh)
		}
	}()
	return eventCh, errCh
}
