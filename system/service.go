package system

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store}
}

func (svc *Service) GetSettings() (*Settings, error) {
	return svc.store.GetSettings()
}

func (svc *Service) SaveSettings(settings *Settings) error {
	return svc.store.SaveSettings(settings)
}

func (svc *Service) IsMaintenanceMode() bool {
	settings, err := svc.GetSettings()
	// TODO: handle error, log
	return err == nil && settings.MaintenanceMode
}
