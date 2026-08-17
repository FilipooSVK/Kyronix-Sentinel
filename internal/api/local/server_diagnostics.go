package local

// UpdateDiagnostics changes runtime diagnostics.
func (s *Server) UpdateDiagnostics(
	diagnostics Diagnostics,
) {

	s.mu.Lock()

	defer s.mu.Unlock()

	s.diagnostics = diagnostics
}
