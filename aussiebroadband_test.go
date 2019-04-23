package aussiebroadband

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var customer *Customer
var details *CustomerDetails

type loginDetails struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

func TestMain(m *testing.M) {
	dat, err := ioutil.ReadFile("logindetails.json")
	var login loginDetails
	json.Unmarshal(dat, &login)

	cust, err := NewCustomer(login.Username, login.Password)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	customer = cust
	details, err = customer.GetCustomerDetails()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	retCode := m.Run()

	os.Exit(retCode)
}

func TestNewCustomer(t *testing.T) {
	if t == nil {
		t.Error("Customer is null")
	}
}

func TestCustomerDetails(t *testing.T) {
	_, err := customer.GetCustomerDetails()
	if err != nil {
		t.Errorf("Error getting customer details: %s", err)
	}
}

func TestOutages(t *testing.T) {

	_, err := customer.GetOutagesNBN(details.Services.NBN[0].ServiceID)

	if err != nil {
		t.Errorf("Error: %s: %s", t.Name(), err)
	}
}

func TestTests(t *testing.T) {
	_, err := customer.GetTests(details.Services.NBN[0].ServiceID)

	if err != nil {
		t.Errorf("Error: %s: %s", t.Name(), err)
	}
}

func TestUsage(t *testing.T) {
	_, err := customer.GetUsage(details.Services.NBN[0].ServiceID)

	if err != nil {
		t.Errorf("Error: %s: %s", t.Name(), err)
	}

	_, err = customer.GetUsage(-1)
	if err == nil {
		t.Errorf("Error: %s should have thrown an error for service id -1", t.Name())
	}
}

func TestTransactions(t *testing.T) {
	_, err := customer.GetTransactions()

	if err != nil {
		t.Errorf("Error: %s: %s", t.Name(), err)
	}
}
