package rest

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func (s *Server) Shutdown() {
	if s != nil {

		log.Printf("Shutting down http server")
		err := s.listener.Close()
		if err != nil {
			log.Print(err)
		}
	}
}

// RunServer starts the FastML Engine-API server
func RunServer(shutdownCh chan struct{}) error {

	httpServer, err := NewServer(shutdownCh)
	if err != nil {
		close(shutdownCh)
		return err
	}
	defer httpServer.Shutdown()

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	for {
		var sig os.Signal
		shutdownChClosed := false
		select {
		case s := <-signalCh:
			sig = s
		case <-shutdownCh:
			sig = os.Interrupt
			shutdownChClosed = true
		}
		// Check if this is a SIGHUP
		if sig == syscall.SIGHUP {
			// TODO reload
		} else {
			if !shutdownChClosed {
				close(shutdownCh)
			}
			return nil
		}
	}

}
