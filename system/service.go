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
