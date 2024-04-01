package common

type HttpServer interface {
	RegisterRoutes()
	Run()
}

type Server struct {
	name        string
	router      Router
}

type Router interface {
	Run()
}

func NewServer(name string, router Router) *Server {
	return &Server{name, router}
}

func (s *Server) Serve() {
	s.router.Run()
}