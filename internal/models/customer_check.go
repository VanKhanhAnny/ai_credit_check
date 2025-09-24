package models

import (
	"time"
)

// MoneyVND stores VND as an integer (₫ has no minor unit).
type MoneyVND int64

// ==================== Enums (match your dropdowns) ====================

type TriState string

const (
	TriStateEmpty TriState = ""    // Default/zero value  
	TriNA         TriState = "na"
	TriYes        TriState = "yes"
	TriNo         TriState = "no"
)

type YesNo string

const (
	YesNoNA YesNo = ""     // Default/zero value
	Yes     YesNo = "yes"
	No      YesNo = "no"
)

type ClientType string

const (
	ClientTypeCorporateEntity   ClientType = "corporate_entity"
	ClientTypePrivateIndividual ClientType = "private_individual"
)

type CustomerType string

const (
	CustomerTypeNA            CustomerType = "na_private_individual"
	CustomerTypeManufacturing CustomerType = "manufacturing_production"
	CustomerTypeTrading       CustomerType = "trading_commercial"
	CustomerTypeConstruction  CustomerType = "construction_real_estate"
	CustomerTypeServices      CustomerType = "services"
	CustomerTypeAgriculture   CustomerType = "agriculture_forestry_fishery"
	CustomerTypeTechnology    CustomerType = "technology_it_software"
	CustomerTypeEnergy        CustomerType = "energy_utilities"
	CustomerTypeFinance       CustomerType = "finance_insurance_banking"
	CustomerTypeHealthcare    CustomerType = "healthcare_pharmaceuticals"
	CustomerTypeMedia         CustomerType = "media_entertainment"
)

type SourceOfClient string

const (
	SourceEPC           SourceOfClient = "epc"
	SourceDirectNetwork SourceOfClient = "direct_own_network"
	SourceClient        SourceOfClient = "client"
)

type OwnershipBracket string

const (
	Ownership100 OwnershipBracket = "100"
	OwnershipGT50 OwnershipBracket = "gt_50"
	OwnershipLT50 OwnershipBracket = "lt_50"
	OwnershipNA   OwnershipBracket = "na"
)

type LandOwnershipSituation string

const (
	LandOwner       LandOwnershipSituation = "land_owner"
	RentalAgreement LandOwnershipSituation = "rental_agreement"
	Unknown         LandOwnershipSituation = "unknown"
)

type CompanySignboardStatus string

const (
	SignboardMatches    CompanySignboardStatus = "available_matches_client_info"
	SignboardMismatched CompanySignboardStatus = "available_does_not_match_client_info"
	SignboardNotAvail   CompanySignboardStatus = "not_available_or_not_checked"
)

// ==================== Root aggregate ====================

type CustomerCheck struct {
	CheckCompletedAt *time.Time     `json:"check_completed_at,omitempty"`
	Corporate        CorporateInfo  `json:"corporate"`
	Land             LandInfo       `json:"land"`
	Financial        FinancialInfo  `json:"financial"`
	Additional       AdditionalInfo `json:"additional"`
}

// ==================== Corporate ====================

type CorporateInfo struct {
	General      GeneralCorporateInfo   `json:"general"`
	History      CorporateHistory       `json:"history"`
	Relationship RelationshipBackground `json:"relationship"`
	Ownership    OwnershipInfo          `json:"ownership"`
}

type GeneralCorporateInfo struct {
	ClientName             string       `json:"client_name,omitempty"`
	ClientType             ClientType   `json:"client_type,omitempty"`
	TaxCodeMST             string       `json:"tax_code_mst,omitempty"`
	BusinessLicenseGPKD    TriState     `json:"business_license_gpkd,omitempty"`
	BusinessAddress        string       `json:"business_address,omitempty"`
	RegisteredShareCapital *MoneyVND    `json:"registered_share_capital,omitempty"`
	CustomerType           CustomerType `json:"customer_type,omitempty"`
	BusinessOperations     string       `json:"business_operations,omitempty"`
}

type CorporateHistory struct {
	IncorporationDate  *time.Time `json:"incorporation_date,omitempty"`
	HistoryDescription string     `json:"history_description,omitempty"`
}

type RelationshipBackground struct {
	Source SourceOfClient `json:"source,omitempty"`
}

type OwnershipInfo struct {
	OwnersName          string           `json:"owners_name,omitempty"`
	OwnershipCategory   OwnershipBracket `json:"ownership_category,omitempty"`
	CompanyDirectorName string           `json:"company_director_name,omitempty"`
	KeyDecisionMaker    string           `json:"key_decision_maker,omitempty"`
}

// ==================== Land ====================

type LandInfo struct {
	EVN       EVNInformation           `json:"evn"`
	Ownership LandOwnershipInformation `json:"ownership"`
}

type EVNInformation struct {
	BillingAddress              string    `json:"billing_address,omitempty"`
	BillingAddressMatchesClient YesNo     `json:"billing_address_matches_client,omitempty"`
	BillingAmount               *MoneyVND `json:"billing_amount,omitempty"`
	BilledAmountsMatchExpenses  TriState  `json:"billed_amounts_match_expenses,omitempty"`
}

type LandOwnershipInformation struct {
	Situation            LandOwnershipSituation `json:"situation,omitempty"`
	LandownerIsSignatory YesNo                  `json:"landowner_is_signatory,omitempty"`
	LeaseExpirationDate  *time.Time             `json:"lease_expiration_date,omitempty"`
	OwnedDocsComplete    YesNo                  `json:"owned_docs_complete,omitempty"`
}

// ==================== Financial ====================

type FinancialInfo struct {
	FinancialStatementDate *time.Time `json:"financial_statement_date,omitempty"`
	PL                    PLInfo     `json:"pl"`
	BalanceSheet          BalanceSheetInfo `json:"balance_sheet"`
	Loans                 []LoanInfo `json:"loans"`
}

type PLInfo struct {
	TotalRevenues     [5]MoneyVND `json:"total_revenues"`     // 30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23
	TotalCosts        [5]MoneyVND `json:"total_costs"`        // 30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23
	TotalEnergyCosts  [5]MoneyVND `json:"total_energy_costs"` // 30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23
}

type BalanceSheetInfo struct {
	TotalAssets [5]MoneyVND `json:"total_assets"` // 30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23
	TotalDebt   [5]MoneyVND `json:"total_debt"`   // 30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23
}

type LoanInfo struct {
	LoanType           LoanType       `json:"loan_type,omitempty"`
	DebtClassification DebtClassification `json:"debt_classification,omitempty"`
	OutstandingAmount  *MoneyVND      `json:"outstanding_amount,omitempty"`
	AnnualInterestCost *MoneyVND      `json:"annual_interest_cost,omitempty"`
	AnnualAmortization *MoneyVND      `json:"annual_amortization,omitempty"`
	Maturity           *time.Time     `json:"maturity,omitempty"`
	PaymentHistory     string         `json:"payment_history,omitempty"`
}

type LoanType string

const (
	LoanTypeEmpty            LoanType = ""                     // Default/zero value
	LoanTypeShortTerm        LoanType = "short_term_loan"
	LoanTypeMediumTerm       LoanType = "medium_term_loan"
	LoanTypeLongTerm         LoanType = "long_term_loan"
	LoanTypeCreditCard       LoanType = "credit_card"
	LoanTypeOverdrafts       LoanType = "overdrafts"
	LoanTypeGuarantee        LoanType = "guarantee"
	LoanTypeFinancialLeasing LoanType = "financial_leasing"
	LoanTypeFactoring        LoanType = "factoring"
	LoanTypeConsumerLoan     LoanType = "consumer_loan"
	LoanTypeOtherCredit      LoanType = "other_credit_facility"
)

type DebtClassification string

const (
	DebtClassificationEmpty   DebtClassification = ""                         // Default/zero value
	DebtClassificationGroup1  DebtClassification = "group_1_current_debt"
	DebtClassificationGroup2  DebtClassification = "group_2_special_mention_debt"
	DebtClassificationGroup3  DebtClassification = "group_3_substandard_debt"
	DebtClassificationGroup4  DebtClassification = "group_4_doubtful_debt"
	DebtClassificationGroup5  DebtClassification = "group_5_loss_debt"
)

// ==================== Additional / Site Visit ====================

type AdditionalInfo struct {
	SiteVisit SiteVisit `json:"site_visit"`
}

type SiteVisit struct {
	CompanySignboard CompanySignboardStatus `json:"company_signboard,omitempty"`
}
