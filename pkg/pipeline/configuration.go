package pipeline

type Configuration interface {
	Id() string
}

func FindConfigurationById[C Configuration](configurations []C, id string) *C {
	for _, configuration := range configurations {
		if configuration.Id() == id {
			return &configuration
		}
	}
	return nil
}
