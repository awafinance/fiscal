package info

type Amount struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type Party struct {
	Role     string `json:"role,omitempty"`
	Name     string `json:"name,omitempty"`
	Document string `json:"document,omitempty"`
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

type Address struct {
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	CityCode     string `json:"cityCode,omitempty"`
	CityName     string `json:"cityName,omitempty"`
	State        string `json:"state,omitempty"`
	ZipCode      string `json:"zipCode,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
	CountryName  string `json:"countryName,omitempty"`
}

type EmitterDetail struct {
	TradeName string   `json:"tradeName,omitempty"`
	IE        string   `json:"ie,omitempty"`
	IEST      string   `json:"iest,omitempty"`
	IM        string   `json:"im,omitempty"`
	CNAE      string   `json:"cnae,omitempty"`
	CRT       string   `json:"crt,omitempty"`
	Phone     string   `json:"phone,omitempty"`
	Email     string   `json:"email,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

type EmitterDetailInfo interface {
	GetEmitterDetail() *EmitterDetail
}
