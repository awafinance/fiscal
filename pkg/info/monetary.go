package info

const MonetaryContractVersion = "1"

type MonetaryInterpretationStatus string

const (
	MonetaryInterpretationComplete    MonetaryInterpretationStatus = "complete"
	MonetaryInterpretationUnsupported MonetaryInterpretationStatus = "unsupported"
)

type FiscalArtifactScope string

const (
	FiscalArtifactFullDocument FiscalArtifactScope = "full_document"
	FiscalArtifactDeclaration  FiscalArtifactScope = "declaration"
	FiscalArtifactSummary      FiscalArtifactScope = "summary"
	FiscalArtifactLifecycle    FiscalArtifactScope = "lifecycle_event"
)

type MonetaryProvenance struct {
	Family      string                 `json:"family"`
	Root        string                 `json:"root"`
	Model       string                 `json:"model,omitempty"`
	Layout      string                 `json:"layout,omitempty"`
	Namespace   string                 `json:"namespace"`
	Scope       FiscalArtifactScope    `json:"scope"`
	Schema      string                 `json:"schema,omitempty"`
	SchemaFiles []MonetarySchemaSource `json:"schemaFiles,omitempty"`
}

type MonetarySchemaSource struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type MonetaryFactKind string

const (
	MonetaryFactDeclaredAmount     MonetaryFactKind = "declared_amount"
	MonetaryFactDocumentTotal      MonetaryFactKind = "document_total"
	MonetaryFactServiceTotal       MonetaryFactKind = "service_total"
	MonetaryFactSettlementAmount   MonetaryFactKind = "settlement_amount"
	MonetaryFactServiceAmount      MonetaryFactKind = "service_amount"
	MonetaryFactComponent          MonetaryFactKind = "component"
	MonetaryFactDiscount           MonetaryFactKind = "discount"
	MonetaryFactTaxBase            MonetaryFactKind = "tax_base"
	MonetaryFactActualTax          MonetaryFactKind = "actual_tax"
	MonetaryFactRetention          MonetaryFactKind = "retention"
	MonetaryFactApproximateTax     MonetaryFactKind = "approximate_tax_burden"
	MonetaryFactApproximateTaxRate MonetaryFactKind = "approximate_tax_burden_percentage"
	MonetaryFactBilling            MonetaryFactKind = "billing"
	MonetaryFactInstallment        MonetaryFactKind = "installment"
	MonetaryFactPayment            MonetaryFactKind = "payment"
	MonetaryFactUnitPrice          MonetaryFactKind = "unit_price"
)

type MonetaryFactScope string

const (
	MonetaryScopeDocument    MonetaryFactScope = "document"
	MonetaryScopeItem        MonetaryFactScope = "item"
	MonetaryScopeComponent   MonetaryFactScope = "component"
	MonetaryScopeBilling     MonetaryFactScope = "billing"
	MonetaryScopeInstallment MonetaryFactScope = "installment"
	MonetaryScopePayment     MonetaryFactScope = "payment"
)

type MonetaryQualifier struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MonetaryRetention struct {
	IndicatorCode string   `json:"indicatorCode,omitempty"`
	Actor         string   `json:"actor"`
	TaxTypes      []string `json:"taxTypes"`
	Allocation    string   `json:"allocation"`
}

type MonetaryFact struct {
	ID         string              `json:"id"`
	Code       string              `json:"code"`
	Path       string              `json:"path"`
	Element    string              `json:"element"`
	Value      string              `json:"value"`
	Unit       string              `json:"unit"`
	Currency   string              `json:"currency,omitempty"`
	Kind       MonetaryFactKind    `json:"kind"`
	Scope      MonetaryFactScope   `json:"scope"`
	Occurrence int                 `json:"occurrence"`
	Label      string              `json:"label,omitempty"`
	TaxTypes   []string            `json:"taxTypes,omitempty"`
	Qualifiers []MonetaryQualifier `json:"qualifiers,omitempty"`
	Retention  *MonetaryRetention  `json:"retention,omitempty"`
}

type MonetaryCheckStatus string

const (
	MonetaryCheckPassed           MonetaryCheckStatus = "passed"
	MonetaryCheckFailed           MonetaryCheckStatus = "failed"
	MonetaryCheckNotApplicable    MonetaryCheckStatus = "not_applicable"
	MonetaryCheckInsufficientData MonetaryCheckStatus = "insufficient_data"
)

type MonetaryCheckTerm struct {
	FactID    string `json:"factId"`
	Operation string `json:"operation"`
	Value     string `json:"value"`
	Reason    string `json:"reason,omitempty"`
}

type MonetaryCalculation struct {
	Variant    string              `json:"variant"`
	Status     MonetaryCheckStatus `json:"status"`
	Calculated string              `json:"calculated,omitempty"`
	Difference string              `json:"difference,omitempty"`
	Terms      []MonetaryCheckTerm `json:"terms,omitempty"`
}

type MonetaryCheck struct {
	Rule          string                `json:"rule"`
	Description   string                `json:"description"`
	Status        MonetaryCheckStatus   `json:"status"`
	DeclaredFact  string                `json:"declaredFact,omitempty"`
	DeclaredValue string                `json:"declaredValue,omitempty"`
	Tolerance     string                `json:"tolerance,omitempty"`
	RejectionCode string                `json:"rejectionCode,omitempty"`
	Reason        string                `json:"reason,omitempty"`
	Calculations  []MonetaryCalculation `json:"calculations,omitempty"`
}

type MonetaryInterpretation struct {
	ContractVersion string                       `json:"contractVersion"`
	Status          MonetaryInterpretationStatus `json:"status"`
	Reason          string                       `json:"reason,omitempty"`
	Provenance      MonetaryProvenance           `json:"provenance"`
	Facts           []MonetaryFact               `json:"facts"`
	Checks          []MonetaryCheck              `json:"checks"`
}
