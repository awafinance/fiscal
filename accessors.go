package fiscal

import "github.com/awafinance/fiscal/pkg/info"

type Amount = info.Amount
type Address = info.Address
type Party = info.Party
type RelatedDocument = info.RelatedDocument
type Location = info.Location
type LifecycleEventFacts = info.LifecycleEventFacts
type MonetaryInterpretation = info.MonetaryInterpretation
type MonetaryProvenance = info.MonetaryProvenance
type MonetarySchemaSource = info.MonetarySchemaSource
type MonetaryFact = info.MonetaryFact
type MonetaryFactKind = info.MonetaryFactKind
type MonetaryFactScope = info.MonetaryFactScope
type MonetaryRetention = info.MonetaryRetention
type MonetaryQualifier = info.MonetaryQualifier
type MonetaryCheck = info.MonetaryCheck
type MonetaryCheckStatus = info.MonetaryCheckStatus
type MonetaryCalculation = info.MonetaryCalculation
type MonetaryCheckTerm = info.MonetaryCheckTerm

const MonetaryContractVersion = info.MonetaryContractVersion

const (
	MonetaryInterpretationComplete    = info.MonetaryInterpretationComplete
	MonetaryInterpretationUnsupported = info.MonetaryInterpretationUnsupported

	FiscalArtifactFullDocument = info.FiscalArtifactFullDocument
	FiscalArtifactDeclaration  = info.FiscalArtifactDeclaration
	FiscalArtifactSummary      = info.FiscalArtifactSummary
	FiscalArtifactLifecycle    = info.FiscalArtifactLifecycle

	MonetaryFactDeclaredAmount     = info.MonetaryFactDeclaredAmount
	MonetaryFactDocumentTotal      = info.MonetaryFactDocumentTotal
	MonetaryFactServiceTotal       = info.MonetaryFactServiceTotal
	MonetaryFactSettlementAmount   = info.MonetaryFactSettlementAmount
	MonetaryFactServiceAmount      = info.MonetaryFactServiceAmount
	MonetaryFactComponent          = info.MonetaryFactComponent
	MonetaryFactDiscount           = info.MonetaryFactDiscount
	MonetaryFactTaxBase            = info.MonetaryFactTaxBase
	MonetaryFactActualTax          = info.MonetaryFactActualTax
	MonetaryFactRetention          = info.MonetaryFactRetention
	MonetaryFactApproximateTax     = info.MonetaryFactApproximateTax
	MonetaryFactApproximateTaxRate = info.MonetaryFactApproximateTaxRate
	MonetaryFactBilling            = info.MonetaryFactBilling
	MonetaryFactInstallment        = info.MonetaryFactInstallment
	MonetaryFactPayment            = info.MonetaryFactPayment
	MonetaryFactUnitPrice          = info.MonetaryFactUnitPrice

	MonetaryScopeDocument    = info.MonetaryScopeDocument
	MonetaryScopeItem        = info.MonetaryScopeItem
	MonetaryScopeComponent   = info.MonetaryScopeComponent
	MonetaryScopeBilling     = info.MonetaryScopeBilling
	MonetaryScopeInstallment = info.MonetaryScopeInstallment
	MonetaryScopePayment     = info.MonetaryScopePayment

	MonetaryCheckPassed           = info.MonetaryCheckPassed
	MonetaryCheckFailed           = info.MonetaryCheckFailed
	MonetaryCheckNotApplicable    = info.MonetaryCheckNotApplicable
	MonetaryCheckInsufficientData = info.MonetaryCheckInsufficientData
)

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

func (d *Document) MonetaryInterpretation() *MonetaryInterpretation {
	if d == nil {
		return nil
	}
	return d.monetary
}
