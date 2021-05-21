package server


func (s *GenericAPIServer) RemoveOpenAPIData() {
	if s.Handler != nil && s.Handler.NonGoRestfulMux != nil {
		s.Handler.NonGoRestfulMux.Unregister("/openapi/v2")
	}
	s.openAPIConfig = nil
}
