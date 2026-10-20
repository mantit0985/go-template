package health

type Status struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Check() Status {
	return Status{
		Status: "OK",
		Details: map[string]string{
			"version": "1.0.0",
		},
	}
}

func (s *Service) Ready() Status {
	return Status{Status: "OK"}
}

type Version struct {
	Version string `json:"version"`
}

func (s *Service) Version() Version {
	return Version{Version: "1.0.0"}
}
