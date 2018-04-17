package SteGo

import (
	"fmt"
	"net/http"
	"time"
)

type ServiceOperative interface {
	BeforeCall(original, specific *http.Request) *http.Request
	AfterCall(original *http.Request, res *http.Response) bool
}

type ServiceCall struct {
	Method     string
	Name       string
	TimeoutSec int
	Operative  ServiceOperative
	client     *http.Client
}

type ServiceChainOperative interface {
	Top(r *http.Request)
	Bottom(r *http.Request, w http.ResponseWriter)
}

type ServiceChain struct {
	Services  []*ServiceCall
	Operative ServiceChainOperative
}

func (sc *ServiceCall) Check() error {
	//Check Method is correct
	return nil
}

func (sc *ServiceCall) Call(sv *ServiceVersion, original *http.Request) error {

	if sv == nil {

	}

	//Create the request for the single service
	req, err := http.NewRequest(sc.Method, sv.Address, nil)

	if err != nil {

	}

	newSpecific := sc.Operative.BeforeCall(original, req)

	//if the client is not set
	if sc.client == nil {
		//Set the Timeout
		sc.client = &http.Client{Timeout: time.Duration(sc.TimeoutSec) * time.Second}
	}
	fmt.Printf("address of the request:%v\n", newSpecific.URL)
	res, err := sc.client.Do(newSpecific)

	if err != nil {
		return err
	}

	//the afterCall should set what it needs in the context of the original Requst
	if ok := sc.Operative.AfterCall(original, res); !ok {

	}

	return nil

}

func (sc *ServiceChain) indexOf(name string) int {
	for i, s := range sc.Services {
		if s.Name == name {
			return i
		}
	}
	return -1
}

//The default is to add the Service Call at the bottom
func (sc *ServiceChain) AddServiceCall(s ServiceCall) error {
	if err := s.Check(); err != nil {
		return err
	}
	sc.Services = append(sc.Services, &s)
	return nil
}

func (sc *ServiceChain) AddServiceCallAt(s ServiceCall, pos int) error {
	if err := s.Check(); err != nil {
		return err
	}
	sc.Services = append(append(sc.Services[:pos], &s), sc.Services[pos:]...)
	return nil
}
func (sc *ServiceChain) AddServiceCallAfter(s ServiceCall, serviceName string) error {
	if err := s.Check(); err != nil {
		return err
	}
	pos := sc.indexOf(serviceName)
	if pos == -1 {

	}
	sc.AddServiceCallAt(s, pos+1)
	return nil
}

func (sc *ServiceChain) AddServiceCallBefore(s ServiceCall, serviceName string) error {
	if err := s.Check(); err != nil {
		return err
	}
	pos := sc.indexOf(serviceName)
	if pos == -1 {

	}
	sc.AddServiceCallAt(s, pos)
	return nil
}

func (sChain *ServiceChain) walkTheChain(srv map[string]*Service, version int, req *http.Request, w http.ResponseWriter) {
	if sChain.Operative != nil {
		sChain.Operative.Top(req)
	}

	for _, sc := range sChain.Services {
		// Get the service from the name
		s, ok := srv[sc.Name]
		if !ok {

		}

		sv := s.Version(version)

		if sv == nil {

		}

		err := sc.Call(sv, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			break
		}
	}

	if sChain.Operative != nil {
		sChain.Operative.Bottom(req, w)
	}
}
