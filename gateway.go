package SteGo

import (
	"net/http"
)

type APIgateway struct {
	chains   map[string]*ServiceChain
	services map[string]*Service
}

func New() *APIgateway {
	return &APIgateway{}
}

func (ag *APIgateway) NewService(name string) *Service {
	if name == `` {
		return nil
	}

	s := new(Service)
	ag.services[name] = s
	return s
}

func (ag *APIgateway) Service(name string) *Service {
	return ag.services[name]
}

func (ag *APIgateway) Chain(name string) *ServiceChain {
	return ag.chains[name]
}

func (ag *APIgateway) NewChain(name string) *ServiceChain {
	sc := new(ServiceChain)
	ag.chains[name] = sc
	return sc
}

func (ag *APIgateway) ServeRequestWithChain(name string, version int) func(http.ResponseWriter, *http.Request) {

	//Get chain to use
	chain, ok := ag.chains[name]
	if !ok {

	}

	return func(w http.ResponseWriter, r *http.Request) {
		//Copy the Request
		req := new(http.Request)
		*req = *r
		//Walk the chain, checking for every step if any errors occurred
		chain.walkTheChain(ag.services, version, req, w)
	}

}
