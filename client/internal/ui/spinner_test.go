package ui

import (
   "testing"
   "time"

   "github.com/stretchr/testify/require"
)

func TestSpinnerStartStop(t *testing.T) {
   sp := NewSpinner()
   require.NotNil(t, sp)
   sp.Start()
   // allow spinner goroutine to run at least one iteration
   time.Sleep(10 * time.Millisecond)
   sp.Stop()
   // done channel should be closed
   select {
   case <-sp.done:
       // ok
   default:
       t.Error("expected done channel to be closed")
   }
}