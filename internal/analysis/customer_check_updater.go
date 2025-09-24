package analysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"extraction/internal/models"
)

// UpdateCustomerCheck updates a CustomerCheck object with information extracted from a document
func UpdateCustomerCheck(check *models.CustomerCheck, extractedData map[string]interface{}, source DocumentSource) {
	switch source {
	case SourceBusinessLicense:
		updateFromBusinessLicense(&check.Corporate, extractedData)
	case SourceEVNBill:
		updateFromEVNBill(&check.Land.EVN, extractedData)
	case SourceLandCertificate:
		updateFromLandCertificate(&check.Land.Ownership, extractedData)
	case SourceIDCheck:
		updateFromIDCheck(&check.Corporate.Ownership, extractedData)
	case SourceSiteVisitPhotos:
		updateFromSiteVisit(&check.Additional.SiteVisit, extractedData)
	case SourceFinancialStatement:
		updateFromFinancialStatement(&check.Financial, extractedData)
	case SourceCICReport:
		updateFromCICReport(&check.Corporate.History, &check.Financial.Loans, extractedData)
	}
}

func updateFromBusinessLicense(info *models.CorporateInfo, data map[string]interface{}) {
	if clientName, ok := data["client_name"].(string); ok {
		info.General.ClientName = clientName
	}
	if clientType, ok := data["client_type"].(string); ok {
		switch strings.ToLower(clientType) {
		case "corporate_entity":
			info.General.ClientType = models.ClientTypeCorporateEntity
		case "private_individual":
			info.General.ClientType = models.ClientTypePrivateIndividual
		}
	}
	if taxCode, ok := data["tax_code_mst"].(string); ok {
		info.General.TaxCodeMST = taxCode
	}
	if license, ok := data["business_license_gpkd"].(string); ok {
		switch strings.ToLower(license) {
		case "yes":
			info.General.BusinessLicenseGPKD = models.TriYes
		case "no":
			info.General.BusinessLicenseGPKD = models.TriNo
		case "na", "n/a":
			info.General.BusinessLicenseGPKD = models.TriNA
		}
	}
	if address, ok := data["business_address"].(string); ok {
		info.General.BusinessAddress = address
	}
	if capital, ok := data["registered_share_capital"].(float64); ok {
		v := models.MoneyVND(capital)
		info.General.RegisteredShareCapital = &v
	}
	if operations, ok := data["business_operations"].(string); ok {
		info.General.BusinessOperations = operations
	}
	if customerType, ok := data["customer_type"].(string); ok {
		switch customerType {
		case "manufacturing_production":
			info.General.CustomerType = models.CustomerTypeManufacturing
		case "trading_commercial":
			info.General.CustomerType = models.CustomerTypeTrading
		case "construction_real_estate":
			info.General.CustomerType = models.CustomerTypeConstruction
		case "services":
			info.General.CustomerType = models.CustomerTypeServices
		case "agriculture_forestry_fishery":
			info.General.CustomerType = models.CustomerTypeAgriculture
		case "technology_it_software":
			info.General.CustomerType = models.CustomerTypeTechnology
		case "energy_utilities":
			info.General.CustomerType = models.CustomerTypeEnergy
		case "finance_insurance_banking":
			info.General.CustomerType = models.CustomerTypeFinance
		case "healthcare_pharmaceuticals":
			info.General.CustomerType = models.CustomerTypeHealthcare
		case "media_entertainment":
			info.General.CustomerType = models.CustomerTypeMedia
		case "na_private_individual":
			info.General.CustomerType = models.CustomerTypeNA
		}
	}
	if date, ok := data["incorporation_date"].(string); ok {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			info.History.IncorporationDate = &t
		}
	}
	if owner, ok := data["owners_name"].(string); ok {
		info.Ownership.OwnersName = owner
	}
	if ownership, ok := data["ownership_category"].(string); ok {
		switch ownership {
		case "100":
			info.Ownership.OwnershipCategory = models.Ownership100
		case "gt_50", ">50%":
			info.Ownership.OwnershipCategory = models.OwnershipGT50
		case "lt_50", "<50%":
			info.Ownership.OwnershipCategory = models.OwnershipLT50
		case "na", "n/a":
			info.Ownership.OwnershipCategory = models.OwnershipNA
		}
	}
	if decisionMaker, ok := data["key_decision_maker"].(string); ok {
		info.Ownership.KeyDecisionMaker = decisionMaker
	}
}

func updateFromEVNBill(info *models.EVNInformation, data map[string]interface{}) {
	if address, ok := data["billing_address"].(string); ok {
		info.BillingAddress = address
	}
	if matches, ok := data["billing_address_matches_client"].(string); ok {
		switch strings.ToLower(strings.TrimSpace(matches)) {
		case "yes", "true", "1":
			info.BillingAddressMatchesClient = models.Yes
		case "no", "false", "0":
			info.BillingAddressMatchesClient = models.No
		default:
			info.BillingAddressMatchesClient = models.No // Default to No for unclear responses
		}
	}
	if amount, ok := data["billing_amount"].(float64); ok {
		v := models.MoneyVND(amount)
		info.BillingAmount = &v
	}
	if matches, ok := data["billed_amounts_match_expenses"].(string); ok {
		switch strings.ToLower(strings.TrimSpace(matches)) {
		case "yes", "true", "1", "match", "matches":
			info.BilledAmountsMatchExpenses = models.TriYes
		case "no", "false", "0", "does not match", "doesn't match":
			info.BilledAmountsMatchExpenses = models.TriNo
		default:
			info.BilledAmountsMatchExpenses = models.TriNo // Default to No for unclear responses
		}
	}
}


func updateFromLandCertificate(info *models.LandOwnershipInformation, data map[string]interface{}) {
	// Set situation based on AI classification
	if situation, ok := data["situation"].(string); ok {
		switch strings.ToLower(situation) {
		case "land_owner":
			info.Situation = models.LandOwner
		case "rental_agreement":
			info.Situation = models.RentalAgreement
		case "unknown":
			info.Situation = models.Unknown
		default:
			info.Situation = models.Unknown // default to unknown if unclear
		}
	} else {
		info.Situation = models.Unknown // default to unknown if not specified
	}
	
	if signatory, ok := data["landowner_is_signatory"].(string); ok {
		switch strings.ToLower(strings.TrimSpace(signatory)) {
		case "yes", "true", "1":
			info.LandownerIsSignatory = models.Yes
		case "no", "false", "0":
			info.LandownerIsSignatory = models.No
		default:
			info.LandownerIsSignatory = models.YesNoNA // Default to NA for unclear responses
		}
	}
	
	if complete, ok := data["documentation_complete"].(string); ok {
		switch strings.ToLower(strings.TrimSpace(complete)) {
		case "yes", "true", "1", "complete":
			info.OwnedDocsComplete = models.Yes
		case "no", "false", "0", "incomplete":
			info.OwnedDocsComplete = models.No
		default:
			info.OwnedDocsComplete = models.YesNoNA // Default to NA for unclear responses
		}
	}
	
	if date, ok := data["lease_expiration_date"].(string); ok && date != "0000-00-00" && date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			info.LeaseExpirationDate = &t
		}
	}
}

func updateFromIDCheck(info *models.OwnershipInfo, data map[string]interface{}) {
	if director, ok := data["company_director_name"].(string); ok {
		info.CompanyDirectorName = director
	}
	if decisionMaker, ok := data["key_decision_maker"].(string); ok {
		info.KeyDecisionMaker = decisionMaker
	}
}

func updateFromSiteVisit(info *models.SiteVisit, data map[string]interface{}) {
	if signboard, ok := data["company_signboard"].(string); ok {
		switch signboard {
		case "available_matches_client_info":
			info.CompanySignboard = models.SignboardMatches
		case "available_does_not_match_client_info":
			info.CompanySignboard = models.SignboardMismatched
		case "not_available_or_not_checked":
			info.CompanySignboard = models.SignboardNotAvail
		}
	}
}

func updateFromFinancialStatement(info *models.FinancialInfo, data map[string]interface{}) {
	if date, ok := data["financial_statement_date"].(string); ok {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			info.FinancialStatementDate = &t
		}
	}
	
	// Update P&L data
	if revenues, ok := data["total_revenues"].([]interface{}); ok && len(revenues) == 5 {
		for i, v := range revenues {
			if amount, ok := v.(float64); ok {
				info.PL.TotalRevenues[i] = models.MoneyVND(amount)
			}
		}
	}
	if costs, ok := data["total_costs"].([]interface{}); ok && len(costs) == 5 {
		for i, v := range costs {
			if amount, ok := v.(float64); ok {
				info.PL.TotalCosts[i] = models.MoneyVND(amount)
			}
		}
	}
	if energyCosts, ok := data["total_energy_costs"].([]interface{}); ok && len(energyCosts) == 5 {
		for i, v := range energyCosts {
			if amount, ok := v.(float64); ok {
				info.PL.TotalEnergyCosts[i] = models.MoneyVND(amount)
			}
		}
	}
	
	// Update Balance Sheet data
	if assets, ok := data["total_assets"].([]interface{}); ok && len(assets) == 5 {
		for i, v := range assets {
			if amount, ok := v.(float64); ok {
				info.BalanceSheet.TotalAssets[i] = models.MoneyVND(amount)
			}
		}
	}
	if debt, ok := data["total_debt"].([]interface{}); ok && len(debt) == 5 {
		for i, v := range debt {
			if amount, ok := v.(float64); ok {
				info.BalanceSheet.TotalDebt[i] = models.MoneyVND(amount)
			}
		}
	}
}

func updateFromCICReport(info *models.CorporateHistory, loans *[]models.LoanInfo, data map[string]interface{}) {
	// Extract loans array from the data
	if loansData, ok := data["loans"].([]interface{}); ok {
		for _, loanData := range loansData {
			if loanMap, ok := loanData.(map[string]interface{}); ok {
				var loanInfo models.LoanInfo
				
				// Set payment history
				if description, ok := loanMap["payment_history"].(string); ok {
					loanInfo.PaymentHistory = description
				} else {
					loanInfo.PaymentHistory = "No payment history found"
				}
				
				// Update loan type
				loanInfo.LoanType = models.LoanTypeOtherCredit // Default loan type
				if loanType, ok := loanMap["loan_type"].(string); ok {
					switch strings.ToLower(loanType) {
					case "short_term_loan":
						loanInfo.LoanType = models.LoanTypeShortTerm
					case "medium_term_loan":
						loanInfo.LoanType = models.LoanTypeMediumTerm
					case "long_term_loan":
						loanInfo.LoanType = models.LoanTypeLongTerm
					case "credit_card":
						loanInfo.LoanType = models.LoanTypeCreditCard
					case "overdrafts":
						loanInfo.LoanType = models.LoanTypeOverdrafts
					case "guarantee":
						loanInfo.LoanType = models.LoanTypeGuarantee
					case "financial_leasing":
						loanInfo.LoanType = models.LoanTypeFinancialLeasing
					case "factoring":
						loanInfo.LoanType = models.LoanTypeFactoring
					case "consumer_loan":
						loanInfo.LoanType = models.LoanTypeConsumerLoan
					case "other_credit_facility":
						loanInfo.LoanType = models.LoanTypeOtherCredit
					}
				}
				
				// Update debt classification
				loanInfo.DebtClassification = models.DebtClassificationGroup1 // Default debt classification
				if debtClass, ok := loanMap["debt_classification"].(string); ok {
					switch strings.ToLower(debtClass) {
					case "group_1_current_debt":
						loanInfo.DebtClassification = models.DebtClassificationGroup1
					case "group_2_special_mention_debt":
						loanInfo.DebtClassification = models.DebtClassificationGroup2
					case "group_3_substandard_debt":
						loanInfo.DebtClassification = models.DebtClassificationGroup3
					case "group_4_doubtful_debt":
						loanInfo.DebtClassification = models.DebtClassificationGroup4
					case "group_5_loss_debt":
						loanInfo.DebtClassification = models.DebtClassificationGroup5
					}
				}
				
				// Set default amounts to 0 if not provided
				defaultAmount := models.MoneyVND(0)
				loanInfo.OutstandingAmount = &defaultAmount
				if amount, ok := loanMap["outstanding_amount"].(float64); ok && amount > 0 {
					v := models.MoneyVND(amount)
					loanInfo.OutstandingAmount = &v
				}
				
				defaultInterest := models.MoneyVND(0)
				loanInfo.AnnualInterestCost = &defaultInterest
				if interest, ok := loanMap["annual_interest_cost"].(float64); ok && interest > 0 {
					v := models.MoneyVND(interest)
					loanInfo.AnnualInterestCost = &v
				}
				
				defaultAmortization := models.MoneyVND(0)
				loanInfo.AnnualAmortization = &defaultAmortization
				if amortization, ok := loanMap["annual_amortization"].(float64); ok && amortization > 0 {
					v := models.MoneyVND(amortization)
					loanInfo.AnnualAmortization = &v
				}
				
				if maturity, ok := loanMap["maturity"].(string); ok && maturity != "0000-00-00" && maturity != "" {
					if t, err := time.Parse("2006-01-02", maturity); err == nil {
						loanInfo.Maturity = &t
					}
				}
				
				// Add the loan to the loans array
				*loans = append(*loans, loanInfo)
			}
		}
	}
}

// compareAddressesWithGemini uses Gemini to compare two addresses and determine if they refer to the same location
func compareAddressesWithGemini(businessAddress, billingAddress string) (bool, error) {
	// Create a Gemini client
	client, err := NewGeminiClient()
	if err != nil {
		return false, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Create a simpler, more direct prompt for address comparison
	prompt := fmt.Sprintf(`Compare these two addresses and determine if they refer to the same location:

Address 1: "%s"
Address 2: "%s"

Rules:
- BE GENEROUS in matching
- Ignore minor differences in formatting, abbreviations, punctuation, word order
- Consider these as MATCHES: "Street" vs "St", "District" vs "Dist", "Ward" vs "W", "Ho Chi Minh City" vs "HCMC"
- Only return "no" if addresses clearly refer to different locations

Return ONLY: "yes" or "no"`, businessAddress, billingAddress)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Call Gemini to analyze the addresses
	result, err := client.AnalyzeDocument(ctx, prompt, SourceBusinessLicense)
	if err != nil {
		return false, fmt.Errorf("failed to analyze addresses with Gemini: %w", err)
	}

	// Parse the result - look for various possible response formats
	if result == nil {
		return false, fmt.Errorf("empty result from Gemini")
	}

	// Try to extract the result from various possible fields
	var response string
	if text, ok := result["addresses_match"].(string); ok {
		response = text
	} else if text, ok := result["result"].(string); ok {
		response = text
	} else if text, ok := result["match"].(string); ok {
		response = text
	} else {
		// If no structured response, look for the raw text response
		for _, v := range result {
			if str, ok := v.(string); ok && (strings.ToLower(str) == "yes" || strings.ToLower(str) == "no") {
				response = str
				break
			}
		}
	}

	// Parse the response
	response = strings.ToLower(strings.TrimSpace(response))
	if response == "yes" || response == "true" || response == "1" {
		return true, nil
	} else if response == "no" || response == "false" || response == "0" {
		return false, nil
	}

	return false, fmt.Errorf("could not parse Gemini response: %v", result)
}

// CompareAddresses compares billing address with business address and updates the match status
func CompareAddresses(check *models.CustomerCheck) {
	if check.Corporate.General.BusinessAddress != "" && check.Land.EVN.BillingAddress != "" {
		fmt.Printf("Comparing addresses:\nBusiness: %s\nBilling: %s\n", check.Corporate.General.BusinessAddress, check.Land.EVN.BillingAddress)
		
		// Use Gemini for more accurate address comparison
		matches, err := compareAddressesWithGemini(check.Corporate.General.BusinessAddress, check.Land.EVN.BillingAddress)
		if err != nil {
			// Fallback to simple comparison if Gemini fails
			fmt.Printf("Gemini address comparison failed, using fallback: %v\n", err)
			if addressesMatch(check.Corporate.General.BusinessAddress, check.Land.EVN.BillingAddress) {
				check.Land.EVN.BillingAddressMatchesClient = models.Yes
				fmt.Printf("Fallback comparison: YES (addresses match)\n")
			} else {
				check.Land.EVN.BillingAddressMatchesClient = models.No
				fmt.Printf("Fallback comparison: NO (addresses don't match)\n")
			}
		} else {
			if matches {
				check.Land.EVN.BillingAddressMatchesClient = models.Yes
				fmt.Printf("Gemini comparison: YES (addresses match)\n")
			} else {
				check.Land.EVN.BillingAddressMatchesClient = models.No
				fmt.Printf("Gemini comparison: NO (addresses don't match)\n")
			}
		}
	} else {
		fmt.Printf("Cannot compare addresses - missing data:\nBusiness Address: '%s'\nBilling Address: '%s'\n", 
			check.Corporate.General.BusinessAddress, check.Land.EVN.BillingAddress)
	}
}

// addressesMatch compares two addresses and returns true if they refer to the same location
func addressesMatch(address1, address2 string) bool {
	// Normalize addresses by converting to lowercase and removing extra spaces
	addr1 := strings.ToLower(strings.TrimSpace(address1))
	addr2 := strings.ToLower(strings.TrimSpace(address2))
	
	// If they're exactly the same, they match
	if addr1 == addr2 {
		return true
	}
	
	// Remove common punctuation and normalize
	addr1 = normalizeAddress(addr1)
	addr2 = normalizeAddress(addr2)
	
	// Check if normalized addresses match
	if addr1 == addr2 {
		return true
	}
	
	// Check if one address contains the other (for cases where one is more detailed)
	if strings.Contains(addr1, addr2) || strings.Contains(addr2, addr1) {
		return true
	}
	
	// Additional fuzzy matching - check if they share significant parts
	// Split addresses into words and check for overlap
	words1 := strings.Fields(addr1)
	words2 := strings.Fields(addr2)
	
	// If more than 70% of words match, consider it a match
	if len(words1) > 0 && len(words2) > 0 {
		matches := 0
		for _, w1 := range words1 {
			for _, w2 := range words2 {
				if w1 == w2 && len(w1) > 2 { // Ignore short words like "st", "rd"
					matches++
					break
				}
			}
		}
		
		// If more than 70% of words match, consider it a match
		maxWords := len(words1)
		if len(words2) > maxWords {
			maxWords = len(words2)
		}
		
		if float64(matches)/float64(maxWords) > 0.5 {
			return true
		}
	}
	
	return false
}

// normalizeAddress normalizes an address for comparison
func normalizeAddress(address string) string {
	// Replace common abbreviations
	replacements := map[string]string{
		"ho chi minh city": "hcmc",
		"ho chi minh": "hcmc",
		"tp. ho chi minh": "hcmc",
		"tp ho chi minh": "hcmc",
		"thanh pho ho chi minh": "hcmc",
		"quan": "q",
		"phuong": "p",
		"to": "t",
		"ap": "a",
	}
	
	normalized := address
	for full, abbrev := range replacements {
		normalized = strings.ReplaceAll(normalized, full, abbrev)
	}
	
	// Remove punctuation and extra spaces
	normalized = strings.ReplaceAll(normalized, ",", "")
	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	
	// Replace multiple spaces with single space
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}
	
	return strings.TrimSpace(normalized)
}