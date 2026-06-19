package info

type Amount struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type Address struct {
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
	CityCode     string `json:"cityCode,omitempty"`
	CityName     string `json:"cityName,omitempty"`
	State        string `json:"state,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
}

type Party struct {
	Role                  string   `json:"role,omitempty"`
	Name                  string   `json:"name,omitempty"`
	Document              string   `json:"document,omitempty"`
	StateRegistration     string   `json:"stateRegistration,omitempty"`
	MunicipalRegistration string   `json:"municipalRegistration,omitempty"`
	Address               *Address `json:"address,omitempty"`
	Phone                 string   `json:"phone,omitempty"`
	Email                 string   `json:"email,omitempty"`
	SimpleNationalOption  string   `json:"simpleNationalOption,omitempty"`
	SimpleNationalRegime  string   `json:"simpleNationalRegime,omitempty"`
	SpecialTaxRegime      string   `json:"specialTaxRegime,omitempty"`
}

type RelatedDocument struct {
	Type      string `json:"type,omitempty"`
	AccessKey string `json:"accessKey,omitempty"`
	Number    string `json:"number,omitempty"`
	Series    string `json:"series,omitempty"`
}

type Location struct {
	CountryCode string `json:"countryCode,omitempty"`
	State       string `json:"state,omitempty"`
	CityCode    string `json:"cityCode,omitempty"`
	CityName    string `json:"cityName,omitempty"`
}

type Payment struct {
	Method           string `json:"method,omitempty"`
	Amount           string `json:"amount,omitempty"`
	Date             string `json:"date,omitempty"`
	PayerDocument    string `json:"payerDocument,omitempty"`
	ReceiverDocument string `json:"receiverDocument,omitempty"`
}

type Invoice struct {
	Number     string `json:"number,omitempty"`
	OrigAmount string `json:"origAmount,omitempty"`
	Discount   string `json:"discount,omitempty"`
	NetAmount  string `json:"netAmount,omitempty"`
}

type Duplicata struct {
	Number  string `json:"number,omitempty"`
	DueDate string `json:"dueDate,omitempty"`
	Amount  string `json:"amount,omitempty"`
}

type Billing struct {
	Invoice    *Invoice    `json:"invoice,omitempty"`
	Duplicates []Duplicata `json:"duplicates,omitempty"`
}

const (
	LifecycleEventRegistrationStateRequest    = "request"
	LifecycleEventRegistrationStateRegistered = "registered"
)

type LifecycleEventFacts struct {
	RegistrationState string `json:"registrationState,omitempty"`
	Type              string `json:"type,omitempty"`
	Sequence          string `json:"sequence,omitempty"`
	RequestNumber     string `json:"requestNumber,omitempty"`
	IssueDate         string `json:"issueDate,omitempty"`
	ProcessingTime    string `json:"processingTime,omitempty"`
}

type AmountsInfo interface {
	GetAmounts() []Amount
}

type PartiesInfo interface {
	GetParties() []Party
}

type RelatedDocumentsInfo interface {
	GetRelatedDocuments() []RelatedDocument
}

type RouteInfo interface {
	GetModal() string
	GetOrigin() Location
	GetDestination() Location
}

type PaymentsInfo interface {
	GetPayments() []Payment
}

type BillingInfo interface {
	GetBilling() *Billing
	GetDuplicatas() []Duplicata
}

type AdditionalInformation interface {
	GetAdditionalInfo() string
}

type LifecycleEventInfo interface {
	GetEventType() string
	GetEventSequence() string
}

type LifecycleEventFactsInfo interface {
	GetLifecycleEventFacts() *LifecycleEventFacts
}
