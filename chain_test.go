package SteGo

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

const (
	SERVICE_NAME = "testService"
)

//Context-Required Type Definition
type testKey int

var testingKey testKey
var serviceAddress string

//Tester Service
func serviceTest(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)
	fmt.Println(r.URL)
	r.ParseForm()

	// Get the sequence number
	seq := r.PostFormValue("sequence")
	if seq == `` {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failure: Sequence not set")
	}
	fmt.Println(seq)

	if seq == "1" {

		// Value set during first call
		fv := r.PostFormValue("first")
		if fv == "true" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Success")
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Failure: %v not expected", fv)
		}
	}

	if seq == "2" {

		// Value set during second call
		sv := r.PostFormValue("second")

		if sv == "true" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Success Second")
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Failure: %v not expected", sv)
		}

	}
}

//Operative of the Tester Service
type tsOperative struct {
	seq  int
	flag string
}

func (tso tsOperative) BeforeCall(original, specific *http.Request) *http.Request {
	fv := url.Values{}
	switch tso.seq {
	case 1:
		fv.Set("first", "true")
	case 2:
		fv.Set("second", "true")
	}
	fv.Set("sequence", strconv.Itoa(tso.seq))

	//change the request to the real address
	fmt.Println("Changing the Request")
	req, _ := http.NewRequest("POST", serviceAddress, strings.NewReader(fv.Encode()))
	fmt.Printf("Changed address: %v \n", req.URL)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func (tso tsOperative) AfterCall(original *http.Request, res *http.Response) bool {
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	original = original.WithContext(context.WithValue(original.Context(), testingKey, string(body)))
	fmt.Println("Set the response in the context")

	if _, ok := original.Context().Value(testingKey).(string); !ok {
		fmt.Println("THe context was not modified")
	}

	return true
}

//Create the Test Service Version (x)
func createServiceWithVersion(version int, ad string) *Service {
	v := make([]*ServiceVersion, version)
	v[version-1] = &ServiceVersion{
		Address: ad,
		Version: version,
	}
	return &Service{Versions: v}
}

//Create the ServiceCall
func createServiceCall(flag string, c *http.Client, seq int) *ServiceCall {
	return &ServiceCall{
		Method:    "POST",
		Name:      SERVICE_NAME,
		Operative: tsOperative{flag: flag, seq: seq},
		client:    c,
	}
}

//Test the Single service call
func Test_ServiceCall(t *testing.T) {

	testingKey = 8

	st := httptest.NewServer(http.HandlerFunc(serviceTest))
	serviceAddress = st.URL
	defer st.Close()

	// set an address which is going to be overwritten by the Operative
	// Create the proper Service Version
	ts := createServiceWithVersion(2, st.URL+"/something/")
	r, _ := http.NewRequest("GET", "original", nil)
	err := createServiceCall("true", st.Client(), 1).Call(ts.Version(2), r)

	if err != nil {
		t.Fatalf("Service Call failed with : %s", err)
	}

	inContext, ok := r.Context().Value(testingKey).(string)

	if !ok {
		t.Fatalf("Context Value Failure")
	}
	if inContext != "Success" {
		t.Errorf("Token %v not expected", inContext)
	}
}
func checkLengthServiceCall(expected int, sc *ServiceChain, t *testing.T) {
	if len(sc.Services) != expected {
		t.Fatalf("SC not inserted: expected %v", expected)
	}
}

// Test the insertion of Services into Chain
func Test_ServiceInsertion(t *testing.T) {
	// Create a Service Chain
	sc := ServiceChain{}
	// Add a first service
	sc.AddServiceCall(ServiceCall{Name: "firstSC"})
	checkLengthServiceCall(1, &sc, t)

	// Add a Service Call before the first
	sc.AddServiceCallBefore(ServiceCall{Name: "secondSC"}, "firstSC")
	checkLengthServiceCall(2, &sc, t)
	// Check Where it has been inserted
	if sc.Services[0].Name != "secondSC" {
		t.Fatalf("At 0 found %v", sc.Services[0].Name)
	}

	// Add a Service in the middle (after the second)
	sc.AddServiceCallAfter(ServiceCall{Name: "thirdSC"}, "secondSC")
	checkLengthServiceCall(3, &sc, t)
	// Check Where it has been inserted
	if sc.Services[1].Name != "thirdSC" {
		t.Fatalf("At 1 found %v", sc.Services[1].Name)
	}

	//Add a service in given position: 1
	sc.AddServiceCallAt(ServiceCall{Name: "fourthSC"}, 1)
	checkLengthServiceCall(4, &sc, t)
	// Check Where it has been inserted
	if sc.Services[1].Name != "fourthSC" {
		t.Fatalf("At 1 (for at) found %v", sc.Services[1].Name)
	}
}

// Test a chain composed by the same service called twice

//the chain operative
type tcOperative struct {
}

func (tco tcOperative) Top(req *http.Request) {
}

func (tco tcOperative) Bottom(req *http.Request, w http.ResponseWriter) {
	inContext, ok := req.Context().Value(testingKey).(string)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if inContext != "Success second" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(inContext))

}

func HTest_Chain(t *testing.T) {
	//Test Server
	st := httptest.NewServer(http.HandlerFunc(serviceTest))
	// Create a Service Chain
	sc := ServiceChain{
		Operative: tcOperative{},
	}

	//Create the Operative for the Service
	tso1 := tsOperative{flag: "success", seq: 1}
	tso2 := tsOperative{flag: "success", seq: 2}

	sc.AddServiceCall(ServiceCall{
		Name:      "test",
		client:    st.Client(),
		Operative: tso1,
		Method:    "POST",
	})
	sc.AddServiceCall(ServiceCall{
		Name:      "test",
		client:    st.Client(),
		Operative: tso2,
		Method:    "POST",
	})

	//Create the Service Map, to be used by the Chain
	//first create the Service
	testS := Service{}
	testS.AddVersion(ServiceVersion{
		Version: 3,
		Address: serviceAddress,
	})

	//And now set the map
	ms := make(map[string]*Service)
	ms["test"] = &testS

	// Call the chain walking method
	resp := httptest.NewRecorder()
	sc.walkTheChain(ms, 3, nil, resp)
}
