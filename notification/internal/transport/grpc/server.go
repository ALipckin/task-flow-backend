package grpc

import "net"

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Register() {}

func (s *Server) Serve(lis net.Listener) error {
	_ = lis
	return nil
}

func (s *Server) GracefulStop() {}
