package example

import "fmt"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) DoWork(name string) string {
	return fmt.Sprintf("Hello, %s! Work completed.", name)
}
