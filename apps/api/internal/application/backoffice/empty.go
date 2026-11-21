package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

// EmptyWorkspace is the live initializer. It contains no users, money, images or sample catalog.
func EmptyWorkspace() Workspace {
	return Workspace{HomeSettings: DefaultHomeSettings(), DisplaySettings: catalog.DisplaySettings{CardAttributeLimit: 3, Version: 1}, Access: &AccessDirectory{}, Commerce: &commerce.State{Settings: commerce.DefaultSettings()}}
}
