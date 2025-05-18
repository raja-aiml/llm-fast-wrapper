package ui

import (
   "bytes"
   "io"
   "os"
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
// captureOutput redirects stdout and returns captured output.
func captureOutput(f func()) string {
   old := os.Stdout
   r, w, _ := os.Pipe()
   os.Stdout = w
   f()
   w.Close()
   var buf bytes.Buffer
   io.Copy(&buf, r)
   os.Stdout = old
   return buf.String()
}

func TestSpinnerOutput(t *testing.T) {
   sp := NewSpinner()
   out := captureOutput(func() {
       sp.Start()
       time.Sleep(50 * time.Millisecond)
       sp.Stop()
   })
   require.Contains(t, out, "Thinking...")
}