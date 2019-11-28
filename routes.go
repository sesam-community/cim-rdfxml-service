package main

// Routes sets up server URL routes with corresponding handlers
func (s *Server) Routes() {
	s.router.POST("/", s.HandleDefault)
	s.router.POST("/:field", s.HandleField)
	s.router.POST("/:field/:namespace", s.HandleFieldNamespace)
	s.router.RedirectTrailingSlash = false // enables special route semantic handling below with trailing slash
	s.router.POST("/:field/", s.HandleFieldNamespace)
}
