package all

import (
	// Import all generators to trigger their registration
	_ "github.com/valentinesamuel/mockcraft/internal/generators/industries/base"
	_ "github.com/valentinesamuel/mockcraft/internal/generators/industries/aviation"
	_ "github.com/valentinesamuel/mockcraft/internal/generators/industries/health"
	// Add future generators here:
	// _ "github.com/valentinesamuel/mockcraft/internal/generators/industries/health"
)
