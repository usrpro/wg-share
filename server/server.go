package server

import (
	"context"
	"log"
	"net/http"
)

// Server implements a RPC server.
type Server struct {
	listeners []*http.Server
}

func (s *Server) listen() <-chan error {
	ec := make(chan error)
	for _, l := range s.listeners {
		go func(l *http.Server) {
			ec <- l.ListenAndServe()
		}(l)
	}
	return ec
}

// Close the server now
// All errors are send to "log" and only the last error is returned.
func (s *Server) Close() error {
	var err error
	for i, l := range s.listeners {
		if err = l.Close(); err != nil {
			log.Printf("Close %d on %s error: %v", i, l.Addr, err)
		}
	}
	return err
}

// Shutdown the server gracefully
// All errors are send to "log" and only the last error is returned.
func (s *Server) Shutdown(ctx context.Context) error {
	ec := make(chan error)
	for _, l := range s.listeners {
		go func(s *http.Server) {
			ec <- s.Shutdown(ctx)
		}(l)
	}
	var err error
	for i := 0; i < len(s.listeners); i++ {
		err = <-ec
		if err != nil {
			log.Printf("Shutdown %d on %s error: %v", i, s.listeners[i].Addr, err)
		}
	}
	return err
}

// ListenAndServe the RPC servers on all configured addresses.
// Blocks while there are open listeners.
//
// In case of a error other then http.ErrServerClosed on one of the listeners,
// all remaining listeners are closed immediatly.
//
// All errors are send to "log" and only the last error is returned.
func (s *Server) ListenAndServe() error {
	ec := s.listen()
	var err error
	for i, l := range s.listeners {
		err = <-ec
		log.Printf("Listener %d on %s error: %v", i, l.Addr, err)
		if err != http.ErrServerClosed {
			if ce := s.Close(); ce != nil {
				err = ce
			}
			break
		}
	}
	return err
}
