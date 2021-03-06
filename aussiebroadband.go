package aussiebroadband

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
	"golang.org/x/net/publicsuffix"

	httpclient "github.com/ddliu/go-httpclient"
)

//Customer object for the API
type Customer struct {
	http         *httpclient.HttpClient
	Username     string
	password     string
	RefreshToken string
	ExpiresAt    time.Time
	Cookie       string
}

//NBNService part of the Customer details.
type NBNService struct {
	ServiceID   int    `json:"service_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Plan        string `json:"plan"`
	Description string `json:"description"`
	NbnDetails  struct {
		Product        string `json:"product"`
		PoiName        string `json:"poiName"`
		CVCGraph       string `json:"cvcGraph"`
		SpeedPotential *struct {
			DownloadMbps int    `json:"downloadMbps"`
			UploadMbps   int    `json:"uploadMbps"`
			LastTested   string `json:"lastTested"`
		} `json:"speedPotential"`
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

//https://discordapp.com/channels/417239092267319297/559903882310844423/560045101967998977

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

//OutagesNBN - https://myaussie-api.aussiebroadband.com.au/nbn/<sid>/outages
type OutagesNBN struct {
	CurrentNBNOutages []struct {
		Created   string `json:"created"`
		Status    string `json:"status"`
		UpdatedAt string `json:"updated_at"`
	} `json:"currentNbnOutages"`
	ScheduledNBNOutages []struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Duration  string `json:"duration"`
	}
	//`json:"networkEvents"`
	//`json:"aussieOutages"`
}

//Test - https://myaussie-api.aussiebroadband.com.au/tests/<sid>
type Test struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CompletedAt string `json:"completed_at"`
}

//Payment - https://myaussie-api.aussiebroadband.com.au/billing/transactions
type Payment struct {
	ID                  int    `json:"id"`
	Type                string `json:"type"`
	Time                string `json:"time"`
	Description         string `json:"description"`
	AmountCents         int    `json:"amountCents"`
	BalanceCents        int    `json:"balanceCents"`
	RunningBalanceCents int    `json:"runningBalanceCents"`
}

//AuthResponse https://myaussie-auth.aussiebroadband.com.au/login
type AuthResponse struct {
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

const apiVersion = "0.1.0"

//NewCustomer - Create a new instance of the customer struct, therefore allowing usage of the API
func NewCustomer(username string, password string) (*Customer, error) {
	httpclient := httpclient.NewHttpClient().Defaults(httpclient.Map{ 
		httpclient.OPT_USERAGENT: "Cazzar's AussieBB API Client " + apiVersion,
	})

	customer := &Customer{
		http:     httpclient,
		Username: username,
		password: password,
	}

	url := "https://myaussie-auth.aussiebroadband.com.au/login"

	resp, err := customer.http.Post(url, map[string]string{
		"username": customer.Username,
		"password": customer.password,
	})

	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	data := &AuthResponse{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}

	customer.RefreshToken = data.RefreshToken
	customer.ExpiresAt = time.Now().Add(time.Second * time.Duration(data.ExpiresIn))
	customer.Cookie = httpclient.CookieValue("https://my.aussiebroadband.com.au", "myaussie_cookie")

	return customer, nil
}

//FromToken - Create a customer object from a token/refrsh token
func FromToken(username string, password string, token string, refreshToken string, expires time.Time) (*Customer, error) {
	fmt.Println(expires.Format("2006-01-02T15:04:05"))
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	url, err := url.ParseRequestURI("https://my.aussiebroadband.com.au")
	jar.SetCookies(url, []*http.Cookie{&http.Cookie{
		Name:     "myaussie_cookie",
		Value:    token,
		Domain:   ".aussiebroadband.com.au",
		HttpOnly: true,
		Secure:   true,
		//Expires:  expires,
	}})

	httpcl := httpclient.NewHttpClient().Defaults(httpclient.Map {
		httpclient.OPT_USERAGENT: "Cazzar's AussieBB API Client " + apiVersion,
		httpclient.OPT_COOKIEJAR: jar,
	})

	customer := &Customer{
		http:         httpcl,
		Username:     username,
		password:     password,
		RefreshToken: refreshToken,
		Cookie:       token,
		ExpiresAt:    expires,
	}

	return customer, nil
}

//RefreshIfNeeded - Check if the Cookies close to expiring, if so, use the RefreshToken to reload
func (cust *Customer) RefreshIfNeeded() (bool, error) {
	if cust.RefreshToken == "" {
		return false, errors.New("Missing Refresh Token")
	}

	if cust.ExpiresAt.Before(time.Now().Add(time.Hour * time.Duration(24))) {
		resp, err := cust.http.PutJson("https://myaussie-auth.aussiebroadband.com.au/login", map[string]string{
			"refreshToken": cust.RefreshToken,
		})

		if err != nil {
			return false, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return false, fmt.Errorf(resp.Status)
		}

		data := &AuthResponse{}
		err = json.Unmarshal(body, data)
		if err != nil {
			return false, err
		}

		cust.RefreshToken = data.RefreshToken
		cust.ExpiresAt = time.Now().Add(time.Second * time.Duration(data.ExpiresIn))
		cust.Cookie = cust.http.CookieValue("https://my.aussiebroadband.com.au", "myaussie_cookie")
		return true, nil
	}

	return false, nil
}

//GetCustomerDetails - Pull the customer details from the MyAussie customer endpoint.
func (cust *Customer) GetCustomerDetails() (*CustomerDetails, error) {
	resp, err := cust.http.Get("https://myaussie-api.aussiebroadband.com.au/customer")
	if err != nil {
		return nil, err
	}
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

//GetTransactions Get all the payments
func (cust *Customer) GetTransactions() (*[]Payment, error) {
	resp, err := cust.http.Get("https://myaussie-api.aussiebroadband.com.au/billing/transactions")
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

	var txns []Payment
	err = json.Unmarshal(body, &txns)
	if err != nil {
		return nil, err
	}

	return &txns, nil
}

//GetOutagesNBN Get outages for a NBN service
func (cust *Customer) GetOutagesNBN(serviceID int) (*OutagesNBN, error) {
	resp, err := cust.http.Get(fmt.Sprintf("https://myaussie-api.aussiebroadband.com.au/nbn/%d/outages", serviceID))
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

	var outages OutagesNBN
	err = json.Unmarshal(body, &outages)
	if err != nil {
		return nil, err
	}

	return &outages, nil
}

//GetTests - Get the tests associated with the service id
func (cust *Customer) GetTests(serviceID int) (*[]Test, error) {
	resp, err := cust.http.Get(fmt.Sprintf("https://myaussie-api.aussiebroadband.com.au/tests/%d", serviceID))
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

	var tests []Test
	err = json.Unmarshal(body, &tests)
	if err != nil {
		return nil, err
	}

	return &tests, nil
}
