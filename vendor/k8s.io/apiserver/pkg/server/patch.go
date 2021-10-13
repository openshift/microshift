package server

var skipSystemMastersAuthorizer = false

// SkipSystemMastersAuthorizer disable implicitly added system/master authz, and turn it into another authz mode "SystemMasters", to be added via authorization-mode
func SkipSystemMastersAuthorizer() {
	skipSystemMastersAuthorizer = true
}

func (s *GenericAPIServer) RemoveOpenAPIData() {
	if s.Handler != nil && s.Handler.NonGoRestfulMux != nil {
		s.Handler.NonGoRestfulMux.Unregister("/openapi/v2")
	}
	s.openAPIConfig = nil
}
