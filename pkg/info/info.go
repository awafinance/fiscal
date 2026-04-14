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
