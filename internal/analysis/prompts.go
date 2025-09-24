package analysis

import (
	"fmt"
)

// generatePromptForSource creates a specific prompt based on the document source
func generatePromptForSource(text string, source DocumentSource) string {
	basePrompt := fmt.Sprintf("Please analyze the following document text and extract the relevant information in JSON format. The document is a %s.\n\nDocument text:\n%s\n\n", source, text)

	switch source {
	case SourceBusinessLicense:
		return basePrompt + `Please extract the following fields in JSON format:
{
  "client_name": "The name of the business entity",
  "client_type": "Classify as either 'corporate_entity' or 'private_individual' using the rules below",
  "tax_code_mst": "The tax code or business registration number",
  "business_license_gpkd": "Whether a business license exists (yes/no/na)",
  "business_address": "The registered business address",
  "registered_share_capital": "The registered share capital amount in VND (numeric value only)",
  "business_operations": "Description of the business operations",
  "customer_type": "Classify the company's business sector using web search if needed. Choose from: manufacturing_production, trading_commercial, construction_real_estate, services, agriculture_forestry_fishery, technology_it_software, energy_utilities, finance_insurance_banking, healthcare_pharmaceuticals, media_entertainment, or na_private_individual",
  "incorporation_date": "The date of incorporation in YYYY-MM-DD format",
  "owners_name": "The name of the primary owner or major shareholder",
  "ownership_category": "Ownership percentage category (100, gt_50, lt_50, or na)",
  "key_decision_maker": "The name of the person with the largest ownership percentage in the company (extract from ownership/shareholder information in the business license)"
}

IMPORTANT: For client_type classification, you are a strict classifier. Using ONLY the text from the business license (no web lookups), output one value for client_type:
- "corporate_entity"
- "private_individual"

Decision rules (apply in order):
1) If the document states a registered company form (e.g., VN: "Công ty TNHH…", "Công ty cổ phần", "Công ty hợp danh"; EN: "LLC", "Company Limited/Ltd.", "Inc.", "Corp.", "GmbH", "S.A.", "SAS", "SARL", "BV", "NV", "PLC", "Pty Ltd", "LLP", "LP"), choose "corporate_entity".
2) If it says VN: "Hộ kinh doanh", "Doanh nghiệp tư nhân (DNTN)" or EN: "Sole proprietorship/sole trader", or it shows an individual person as Owner with only a DBA/trade name, choose "private_individual".
3) Branch/representative office of a company ("Chi nhánh", "Văn phòng đại diện") is still "corporate_entity".
4) If signals conflict, prefer the explicit enterprise type field ("Loại hình doanh nghiệp" / "Type of enterprise") > legal name suffixes > owner section.
5) If nothing indicates a company form and the legal name is a person's name with an Owner field, choose "private_individual".

Examples:
Input: "Loại hình doanh nghiệp: CÔNG TY TRÁCH NHIỆM HỮU HẠN HAI THÀNH VIÊN TRỞ LÊN" → corporate_entity
Input: "Legal Name: ABC COMPANY LIMITED" → corporate_entity
Input: "Tên hộ kinh doanh: Hộ kinh doanh Nguyễn Văn A" → private_individual
Input: "Owner: Jane Doe; DBA: Rose Nails" → private_individual

IMPORTANT: For customer_type classification, you may need to search the web to understand what the company actually does. Use the company name and business operations description to research the company's main business activities. Choose the most appropriate category:

- "manufacturing_production": Companies that produce, manufacture, or assemble physical goods
- "trading_commercial": Companies that buy and sell goods, import/export, wholesale, retail
- "construction_real_estate": Construction companies, real estate developers, property management
- "services": General service providers (consulting, legal, accounting, cleaning, etc.)
- "agriculture_forestry_fishery": Farming, forestry, fishing, food production
- "technology_it_software": Software development, IT services, tech companies, digital services
- "energy_utilities": Power generation, oil & gas, utilities, renewable energy
- "finance_insurance_banking": Banks, insurance companies, financial services, investment
- "healthcare_pharmaceuticals": Hospitals, clinics, pharmaceutical companies, medical services
- "media_entertainment": Media companies, entertainment, advertising, publishing
- "na_private_individual": Use this for private individuals or when unable to determine

Research the company online if the business operations description is unclear. Look for:
1. Company website
2. Industry classification
3. Main products/services
4. Business model

Examples:
Input: "CÔNG TY TNHH SẢN XUẤT THỰC PHẨM ABC" → manufacturing_production
Input: "ABC TRADING COMPANY LIMITED" → trading_commercial
Input: "ABC CONSTRUCTION CO., LTD" → construction_real_estate
Input: "ABC SOFTWARE SOLUTIONS" → technology_it_software
Input: "ABC BANK" → finance_insurance_banking

IMPORTANT: For key_decision_maker, identify the person with the largest ownership percentage in the company. Look for:
1. Shareholder information with ownership percentages
2. Owner details with share amounts
3. Member/partner information with capital contributions
4. Any section showing ownership distribution

The key decision maker is typically:
- The person with the highest percentage of shares/ownership
- The majority shareholder (if >50%)
- The largest individual shareholder (if no majority holder)
- The company director if they are also a major shareholder

Extract the full name of this person from the business license document.`

	case SourceEVNBill:
		return basePrompt + `Please extract the following fields in JSON format:
{
  "billing_address": "The address on the EVN bill",
  "billing_address_matches_client": "Whether the billing address matches the client's business address (yes/no). Compare the billing address on the EVN bill with the business address from the business license. Consider them a match if they are the same or very similar. BE GENEROUS in matching - minor differences in formatting, abbreviations, punctuation, word order, or common variations should be ignored and treated as a MATCH.",
  "billing_amount": "The billing amount in VND (numeric value only)",
  "billed_amounts_match_expenses": "Compare the billed amounts in the uploaded electricity invoices with expense-related figures in the financial statement (cost of goods sold, administrative expenses, operating expenses, or payments to suppliers). Use approximate matching: consider a match if the difference is within ±5% or if the amounts are the same when rounded to the nearest million VND. For each invoice, return Yes if a match is found (and show the closest number), otherwise return No."
}

ADDRESS MATCHING RULES - BE GENEROUS:
- Consider addresses a MATCH ("yes") if they refer to the same location, even with:
  * Different abbreviations (St/Street, Ave/Avenue, Dist/District, Ward/W)
  * Different punctuation (commas, periods, dashes)
  * Different word order (123 Main St vs Main Street 123)
  * Different formatting (uppercase/lowercase, spacing)
  * Minor spelling variations or typos
  * Missing or extra words (The, Of, And, etc.)
- Only choose "no" if the addresses clearly refer to different locations
- When in doubt between "yes" and "no", choose "yes" (be generous)

EXAMPLES OF ADDRESS MATCHES (should return "yes"):
- "123 Main Street, District 1, HCMC" vs "123 Main St, Dist 1, Ho Chi Minh City" → YES
- "456 Nguyen Van A, Ward 5, Binh Thanh" vs "456 Nguyen Van A Street, W. 5, Binh Thanh District" → YES  
- "789 Le Loi Ave, Tan Binh" vs "789 Le Loi Avenue, Tan Binh District" → YES
- "321 Tran Hung Dao, Q1" vs "321 Tran Hung Dao Street, District 1" → YES

EXAMPLES OF NON-MATCHES (should return "no"):
- "123 Main Street, District 1" vs "456 Other Street, District 2" → NO
- "789 Le Loi, Tan Binh" vs "789 Le Loi, District 7" → NO

For billed_amounts_match_expenses analysis:
1. Compare the EVN bill amount with the "total_energy_costs" from the financial statements
2. Check if the EVN bill amount appears in the financial statement expenses
3. Look for energy/electricity costs in the P&L section that match the EVN bill amount
4. If the amounts are similar or match, indicate "yes"
5. If the amounts don't match or no financial statement is available, indicate "no"
6. If you cannot determine due to missing information, indicate "na"

This field helps verify that the EVN bill expenses are properly recorded in the company's financial statements.`

	case SourceLandCertificate:
		return basePrompt + `Please extract the following fields in JSON format:
{
  "situation": "Land ownership situation (land_owner, rental_agreement, or unknown)",
  "landowner_is_signatory": "Whether the landowner/tenant is the contract signatory (yes/no)",
  "documentation_complete": "Whether the ownership documentation is complete (yes/no) (only if situation is land_owner)",
  "lease_expiration_date": "The expiration date of the lease in YYYY-MM-DD format (only if situation is rental_agreement)"
}

CRITICAL REQUIREMENTS:
1. You MUST provide a value for EVERY field. Do not leave any field empty.
2. For landowner_is_signatory: You MUST choose either "yes" or "no" based on your analysis. If you cannot determine, choose "no" as default.
3. For documentation_complete: You MUST choose either "yes" or "no" based on your analysis. If you cannot determine, choose "no" as default.
4. For lease_expiration_date: Extract the date if found, otherwise use "0000-00-00" as placeholder.

For situation classification, you are a strict classifier. Using ONLY the document TITLE (no body text, no guessing), output one value for situation:
- "land_owner"
- "rental_agreement"
- "unknown"   (if the title alone doesn't clearly indicate either)

Rules (match case-insensitively; handle diacritics removed forms too):
A) Titles that indicate OWNERSHIP / TITLE -> "land_owner"
   - Vietnamese: 
     "Giấy chứng nhận quyền sử dụng đất", 
     "Giấy chứng nhận quyền sử dụng đất, quyền sở hữu nhà ở và tài sản khác gắn liền với đất",
     "GCN QSDĐ", "Sổ đỏ", "Sổ hồng"
   - English:
     "Certificate of Land Use Rights", "Land Title", "Title Deed", 
     "Certificate of Land Use Rights and House Ownership"
   - Administrative decisions implying ownership:
     "Quyết định giao đất" / "Decision on Land Allocation"
B) Titles that indicate LEASE / SUBLEASE -> "rental_agreement"
   - Vietnamese:
     "Hợp đồng thuê đất", "Hợp đồng thuê lại đất", 
     "Hợp đồng cho thuê quyền sử dụng đất", 
     "Phụ lục hợp đồng thuê (đất/mặt bằng)"
   - English:
     "Land Lease Agreement", "Land Sublease Agreement", 
     "Lease Agreement", "Premises Lease", "Sublease Agreement"
   - Administrative decisions implying lease:
     "Quyết định cho thuê đất" / "Decision on Land Lease"
C) If the title is generic or unrelated (e.g., "Biên bản", "Giấy phép xây dựng", "Agreement" without lease/title words), output "unknown".

Conflict resolution:
- If both ownership and lease terms appear, prefer the more specific land term in the title. 
- If still unclear, choose "unknown".

Output format: return ONLY one of the three strings, with no extra text.

Examples:
Input: "Hợp đồng thuê lại đất" -> rental_agreement
Input: "Giấy chứng nhận quyền sử dụng đất, quyền sở hữu nhà ở và tài sản khác gắn liền với đất" -> land_owner
Input: "Certificate of Land Use Rights" -> land_owner
Input: "Land Sublease Agreement" -> rental_agreement
Input: "Quyết định cho thuê đất" -> rental_agreement
Input: "Biên bản bàn giao" -> unknown`

	case SourceIDCheck:
		return basePrompt + `Please extract the following fields in JSON format:
{
  "company_director_name": "The name of the company director",
  "key_decision_maker": "The name of the key decision maker"
}`

	case SourceSiteVisitPhotos:
		return basePrompt + `Please extract the following fields in JSON format:
{
  "company_signboard": "Status of the company signboard (available_matches_client_info, available_does_not_match_client_info, or not_available_or_not_checked)"
}

IMPORTANT: For company_signboard classification, analyze the signboard visible in the site visit photos and compare it with the client name from the business license. Output one of these values:

- "available_matches_client_info": The signboard is clearly visible and the company name on the signboard matches the client name from the business license
- "available_does_not_match_client_info": The signboard is clearly visible but the company name on the signboard does NOT match the client name from the business license
- "not_available_or_not_checked": No signboard is visible in the photos, or the signboard is not clear enough to read the company name

Guidelines:
1. Look for company signs, banners, or nameplates on the building exterior or interior
2. Compare the company name on the signboard with the client name from the business license
3. Consider variations in spelling, abbreviations, or formatting (e.g., "ABC Co., Ltd." vs "ABC Company Limited")
4. If the signboard is partially obscured, damaged, or unclear, choose "not_available_or_not_checked"
5. If no signboard is visible in any of the photos, choose "not_available_or_not_checked"`

	case SourceFinancialStatement:
		return basePrompt + `Please extract the following fields in JSON format:
{
  "financial_statement_date": "Date of the financial statements in YYYY-MM-DD format (the date the financial results are as of)",
  "total_revenues": "Array of 5 numbers for total revenues in VND: [30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23]",
  "total_costs": "Array of 5 numbers for total costs in VND: [30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23]",
  "total_energy_costs": "Array of 5 numbers for total energy costs in VND: [30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23] (should match EVN Bill amounts)",
  "total_assets": "Array of 5 numbers for total assets in VND: [30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23]",
  "total_debt": "Array of 5 numbers for total debt in VND: [30/06/25, 31/12/24, 30/6/24, 31/12/23, 30/6/23]"
}

IMPORTANT: 
1. Extract financial data for the 5 specified periods in chronological order (most recent first)
2. All monetary values should be in VND (Vietnamese Dong) as integers
3. If data is not available for a specific period, use 0 for that position in the array
4. For total_energy_costs, cross-reference with EVN bill amounts when possible
5. Look for P&L statements, income statements, and balance sheets in the document`

	case SourceCICReport:
		return basePrompt + `Please extract loan information from this CIC report. The document may contain multiple loans/credit facilities. Extract ALL loans found and return them as an array.

Return in JSON format:
{
  "loans": [
    {
      "payment_history": "Description of payment history and repayment behavior that could impact approval decisions",
      "loan_type": "Type of loan/credit facility (short_term_loan, medium_term_loan, long_term_loan, credit_card, overdrafts, guarantee, financial_leasing, factoring, consumer_loan, other_credit_facility)",
      "debt_classification": "Debt classification group (group_1_current_debt, group_2_special_mention_debt, group_3_substandard_debt, group_4_doubtful_debt, group_5_loss_debt)",
      "outstanding_amount": "Outstanding loan amount in VND (numeric value only)",
      "annual_interest_cost": "Annual interest cost in VND (numeric value only)",
      "annual_amortization": "Annual amortization amount in VND (numeric value only)",
      "maturity": "Loan maturity date in YYYY-MM-DD format"
    }
  ]
}

CRITICAL REQUIREMENTS FOR MULTIPLE LOAN EXTRACTION:
1. You MUST extract ALL loans/credit facilities found in the document
2. Return an array of loans, even if only one loan is found
3. If NO loans are found, return an empty array: {"loans": []}
4. Each loan object must have ALL required fields
5. For each loan, provide defaults if information is missing:
   - loan_type: "other_credit_facility" (if unclear)
   - debt_classification: "group_1_current_debt" (if unclear)
   - outstanding_amount: 0 (if not found)
   - annual_interest_cost: 0 (if not found)
   - annual_amortization: 0 (if not found)
   - maturity: "0000-00-00" (if not found)
   - payment_history: "No payment history found" (if not found)

SEARCH STRATEGY - LOOK EVERYWHERE:
- Check ALL sections: Balance Sheet, P&L, Notes, Cash Flow, Credit Information
- Look for ANY mention of: loans, credit, debt, borrowing, financing
- Look for Vietnamese terms: khoản vay, tín dụng, nợ, vay vốn, tài trợ
- Extract each distinct loan/credit facility as a separate object
- If multiple entries exist for the same loan type, combine them or list separately based on context

IMPORTANT: You are an assistant that extracts loan information from CIC reports and financial statements.

For loan_type classification, analyze the CIC report to identify the type of credit facility:
- "short_term_loan": Short-Term Loan (up to 12M)
- "medium_term_loan": Medium-Term Loan (1-5Y)
- "long_term_loan": Long-Term Loan (>5Y)
- "credit_card": Credit Card (Thẻ tín dụng)
- "overdrafts": Overdrafts (Thấu chi)
- "guarantee": Guarantee (Bảo lãnh)
- "financial_leasing": Financial Leasing (Cho thuê tài chính)
- "factoring": Factoring (Bảo thanh toán)
- "consumer_loan": Consumer Loan (Tín dụng tiêu dùng)
- "other_credit_facility": Other Credit Facility (Các khoản tín dụng khác)

For debt_classification, analyze the CIC report to determine the debt group:
- "group_1_current_debt": Group 1 - Current Debt (Nợ đủ tiêu chuẩn) - On-time repayment, overdue <= 10 days
- "group_2_special_mention_debt": Group 2 - Special Mention Debt (Nợ cần chú ý) - Overdue 11-90 days, restructured once
- "group_3_substandard_debt": Group 3 - Substandard Debt (Nợ dưới tiêu chuẩn) - Overdue 91-180 days, restructured and overdue
- "group_4_doubtful_debt": Group 4 - Doubtful Debt (Nợ nghi ngờ) - Overdue 181-360 days, restructured multiple times
- "group_5_loss_debt": Group 5 - Loss Debt (Nợ có khả năng mất vốn) - Overdue > 360 days, written off, legal dispute`


	default:
		return basePrompt + `Please extract any relevant information in JSON format that might be useful for customer verification.`
	}
}