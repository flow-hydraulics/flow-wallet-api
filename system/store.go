package system

type Store interface {
	GetSettings() (*Settings, error)
	SaveSettings(*Settings) error
}
