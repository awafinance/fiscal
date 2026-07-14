package fiscal

import "github.com/awafinance/fiscal/pkg/info"

type Amount = info.Amount
type Address = info.Address
type Party = info.Party
type RelatedDocument = info.RelatedDocument
type Location = info.Location
type LifecycleEventFacts = info.LifecycleEventFacts

const (
	LifecycleEventRegistrationStateRequest    = info.LifecycleEventRegistrationStateRequest
	LifecycleEventRegistrationStateRegistered = info.LifecycleEventRegistrationStateRegistered
)

type AmountsInfo = info.AmountsInfo
type PartiesInfo = info.PartiesInfo
type RelatedDocumentsInfo = info.RelatedDocumentsInfo
type RouteInfo = info.RouteInfo
type LifecycleEventInfo = info.LifecycleEventInfo
type LifecycleEventFactsInfo = info.LifecycleEventFactsInfo

type DocumentInfo interface {
	GetAccessKey() string
	GetVersion() string
	GetEnvironment() string
	GetNumber() string
	GetSeries() string
	GetModel() string
	GetIssueDate() string
	GetAmount() string
	GetIssuer() string
	GetIssuerDocument() string
	GetRecipient() string
	GetRecipientDocument() string
	GetProtocolNumber() string
	GetStatusCode() string
	GetStatusReason() string
	IsAuthorized() bool
}

func (d *Document) Info() DocumentInfo {
	if d == nil {
		return nil
	}
	return d.info
}
