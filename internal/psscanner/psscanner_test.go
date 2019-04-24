package psscanner

import (
	"testing"
	"time"
)

const timeout = 100 * time.Millisecond

// refresh

func TestRun(t *testing.T) {
	tests := []struct {
		pids   []int
		events []string
	}{
		{pids: []int{1, 2, 3}, events: []string{
			"UID=???  PID=3      | the-command",
			"UID=???  PID=2      | the-command",
			"UID=???  PID=1      | the-command",
		}},
	}

	for _, tt := range tests {
		restoreGetPIDs := mockPidList(tt.pids)
		restoreCmdLineReader := mockCmdLineReader([]byte("the-command"), nil)
		restoreProcStatusReader := mockProcStatusReader([]byte(""), nil) // don't mock read value since it's not worth it

		pss := NewPSScanner()
		triggerCh := make(chan struct{})
		eventCh, errCh := pss.Run(triggerCh)

		// does nothing without triggering
		select {
		case e := <-eventCh:
			t.Errorf("Received event before trigger: %s", e)
		case err := <-errCh:
			t.Errorf("Received error before trigger: %v", err)
		case <-time.After(timeout):
			// ok
		}

		triggerCh <- struct{}{}

		// received event after the trigger
		for i := 0; i < 3; i++ {
			select {
			case <-time.After(timeout):
				t.Errorf("did not receive event in time")
			case e := <-eventCh:
				if e.String() != tt.events[i] {
					t.Errorf("Wrong event received: got '%s' but wanted '%s'", e, tt.events[i])
				}
			case err := <-errCh:
				t.Errorf("Received unexpected error: %v", err)
			}
		}

		restoreProcStatusReader()
		restoreCmdLineReader()
		restoreGetPIDs()
	}
}
