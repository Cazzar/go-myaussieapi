package aussiebroadband

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	httpclient "github.com/ddliu/go-httpclient"
)

//Customer object for the API
type Customer struct {
	http     *httpclient.HttpClient
	username string
	password string
}

//NBNService part of the Customer details.
type NBNService struct {
	ServiceID   int    `json:"service_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Plan        string `json:"plan"`
	Description string `json:"description"`
	NbnDetails  struct {
		Product string `json:"product"`
		PoiName string `json:"poiName"`
	} `json:"nbnDetails"`
	NextBillDate     time.Time `json:"nextBillDate"`
	OpenDate         string    `json:"openDate"`
	UsageAnniversary int       `json:"usageAnniversary"`
	IPAddresses      []string  `json:"ipAddresses"`
	Address          struct {
		Subaddresstype   string `json:"subaddresstype"`
		Subaddressnumber string `json:"subaddressnumber"`
		Streetnumber     string `json:"streetnumber"`
		Streetname       string `json:"streetname"`
		Streettype       string `json:"streettype"`
		Locality         string `json:"locality"`
		Postcode         string `json:"postcode"`
		State            string `json:"state"`
	} `json:"address"`
	Contract struct {
		ServiceID       int    `json:"service_id"`
		ContractStart   string `json:"contract_start"`
		ContractLength  int    `json:"contract_length"`
		ContractVersion string `json:"contract_version"`
	} `json:"contract"`
}

//CustomerDetails - https://myaussie-api.aussiebroadband.com.au/customer
type CustomerDetails struct {
	CustomerNumber int    `json:"customer_number"`
	BillingName    string `json:"billing_name"`
	Billformat     int    `json:"billformat"`
	Brand          string `json:"brand"`
	PostalAddress  struct {
		Address  string `json:"address"`
		Town     string `json:"town"`
		State    string `json:"state"`
		Postcode string `json:"postcode"`
	} `json:"postalAddress"`
	CommunicationPreferences struct {
		Outages struct {
			Sms    bool `json:"sms"`
			Sms247 bool `json:"sms247"`
			Email  bool `json:"email"`
		} `json:"outages"`
	} `json:"communicationPreferences"`
	Phone               string   `json:"phone"`
	Email               []string `json:"email"`
	PaymentMethod       string   `json:"payment_method"`
	IsSuspended         bool     `json:"isSuspended"`
	AccountBalanceCents int      `json:"accountBalanceCents"`
	Services            struct {
		NBN []NBNService `json:"NBN"`
	} `json:"services"`
	Permissions struct {
		CreatePaymentPlan          bool `json:"createPaymentPlan"`
		UpdatePaymentDetails       bool `json:"updatePaymentDetails"`
		CreateContact              bool `json:"createContact"`
		UpdateContacts             bool `json:"updateContacts"`
		UpdateCustomer             bool `json:"updateCustomer"`
		ChangePassword             bool `json:"changePassword"`
		CreateTickets              bool `json:"createTickets"`
		MakePayment                bool `json:"makePayment"`
		PurchaseDatablocksNextBill bool `json:"purchaseDatablocksNextBill"`
		CreateOrder                bool `json:"createOrder"`
		ViewOrders                 bool `json:"viewOrders"`
	} `json:"permissions"`
	CreditCard struct {
		NameOnCard string `json:"nameOnCard"`
		Number     string `json:"number"`
		Expiry     string `json:"expiry"`
	} `json:"creditCard"`
}

//UsageInformation - https://myaussie-api.aussiebroadband.com.au/broadband/<sid>/usage
type UsageInformation struct {
	UsedMb        int    `json:"usedMb"`
	DownloadedMb  int    `json:"downloadedMb"`
	UploadedMb    int    `json:"uploadedMb"`
	RemainingMb   *int   `json:"remainingMb"`
	DaysTotal     int    `json:"daysTotal"`
	DaysRemaining int    `json:"daysRemaining"`
	LastUpdated   string `json:"lastUpdated"`
}

//NewCustomer - Create a new instance of the customer struct, therefore allowing usage of the API
func NewCustomer(username string, password string) (*Customer, error) {
	const VERSION = "0.0.1"
	httpclient := httpclient.NewHttpClient().WithOption(httpclient.OPT_USERAGENT, "Cazzar's AussieBB API Client "+VERSION)

	customer := &Customer{
		http:     httpclient,
		username: username,
		password: password,
	}

	url := "https://myaussie-auth.aussiebroadband.com.au/login"

	resp, err := customer.http.Post(url, map[string]string{
		"username": customer.username,
		"password": customer.password,
	})

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	return customer, nil
}

//GetCustomerDetails - Pull the customer details from the MyAussie customer endpoint.
func (cust *Customer) GetCustomerDetails() (*CustomerDetails, error) {
	resp, err := cust.http.Get("https://myaussie-api.aussiebroadband.com.au/customer")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	data := &CustomerDetails{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//GetUsage Get the current usage for a given service ID.
func (cust *Customer) GetUsage(serviceID int) (*UsageInformation, error) {
	resp, err := cust.http.Get(fmt.Sprintf("https://myaussie-api.aussiebroadband.com.au/broadband/%d/usage", serviceID))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := UsageInformation{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
