package plugin

// ServiceSet represents a collection of services, providing mechanisms to
// register and manage them.
type ServiceSet struct {
	services map[string]Service
}

// NewServiceSet creates and returns a new instance of ServiceSet with an
// initialized services map.
func NewServiceSet() *ServiceSet {
	return &ServiceSet{
		services: make(map[string]Service),
	}
}

// Register adds a service to the ServiceSet if it has a valid name and is not
// yet registered.
func (s *ServiceSet) Register(svc Service) {
	if name := svc.Name(); name != "" {
		if _, ok := s.services[name]; !ok {
			s.services[name] = svc
		}
	}
}

// Services returns a copy of the registered services within the ServiceSet.
func (s *ServiceSet) Services() map[string]Service {
	svc := make(map[string]Service)
	for k, v := range s.services {
		svc[k] = v
	}

	return svc
}

// Append adds services from another ServiceSet to the current one if they are
// not already present.
func (s *ServiceSet) Append(services *ServiceSet) {
	for k, v := range services.Services() {
		if _, ok := s.services[k]; !ok {
			s.services[k] = v
		}
	}
}
