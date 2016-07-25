package main

import (
	"os"
	"testing"
	"time"

	"code.senomas.com/go/util"
	log "github.com/Sirupsen/logrus"
)

// TestRoutine func
func TestRoutine(t *testing.T) {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)

	log.Info("START")

	var checkOnce util.CheckOnce
	out := make(chan int)

	for i := 0; i < 3; i++ {
		ix := i
		go func() {
			log.Info("ROUTINE IN ", ix)
			if !checkOnce.IsDone() {
				log.Info("ROUTINE INSIDE-1 ", ix)
				time.Sleep(time.Duration(ix) * time.Second)
				checkOnce.Done(func() {
					log.Info("ROUTINE INSIDE-2 ", ix)
					out <- ix
				})
			}
			log.Info("ROUTINE OUT ", ix)
		}()
	}

	res := <-out
	log.Info("GET RES ", res)
	close(out)

	log.Info("FOO CHECK")
	time.Sleep(time.Duration(10) * time.Second)
	log.Info("DONE")
}
