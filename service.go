package SteGo

import ()

type errorServiceVersion struct {
	message string
}

func newErrorServiceVersion(message string) errorServiceVersion {
	return errorServiceVersion{
		message: message,
	}
}

func (e errorServiceVersion) Error() string {
	return "Service Version Fail: " + e.message
}

type ServiceVersion struct {
	Address string
	Version int
}

type Service struct {
	Versions []*ServiceVersion // Versions are inserted at correspondent place in the slice
}

func (s *Service) checkVersion(sv ServiceVersion) error {
	// Check Address is filled
	if sv.Address == `` {
		return newErrorServiceVersion("Address can't be empty")
	}
	return nil
}

func (s *Service) Version(number int) *ServiceVersion {
	// get the max{ versions | version_number <= number}
	if number > len(s.Versions) {
		// get the latest
		return s.Versions[len(s.Versions)-1]
	}

	for i := number - 1; i >= 0; i-- {

		if candidate := s.Versions[i]; candidate != nil {
			return candidate
		}
	}

	return nil

}

func (s *Service) AddVersion(sv ServiceVersion) error {
	if err := s.checkVersion(sv); err != nil {
		return err
	}

	// if there is a gap between the provided version and the last registered,
	// we need to fill such gap
	if len(s.Versions) < sv.Version-1 {
		s.Versions = append(s.Versions, make([]*ServiceVersion, sv.Version-(len(s.Versions)+1))...)
		//Insert the Service Version x at place x-1
		s.Versions = append(s.Versions, &sv)
		return nil
	}

	if len(s.Versions) == sv.Version-1 {
		//Insert the Service Version x at place x-1
		s.Versions = append(s.Versions, &sv)
		return nil
	}

	if s.Versions[sv.Version-1] != nil {
		return newErrorServiceVersion("Version already exists and can't be replaced")
	}

	s.Versions[sv.Version-1] = &sv
	return nil
}
