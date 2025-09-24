package export

import (
	"fmt"
	"time"

	"extraction/internal/models"
	"github.com/xuri/excelize/v2"
)

// WriteCustomerCheck writes a CustomerCheck to an Excel file
func WriteCustomerCheck(check *models.CustomerCheck, outPath string) error {
	f := excelize.NewFile()
	defaultSheet := f.GetSheetName(0)

	corporateSheet := "Corporate"
	landSheet := "Land"
	additionalSheet := "Additional"

	f.NewSheet(corporateSheet)
	f.NewSheet(landSheet)
	f.NewSheet(additionalSheet)
	f.DeleteSheet(defaultSheet)
	sheetIndex, _ := f.GetSheetIndex(corporateSheet)
	f.SetActiveSheet(sheetIndex)

	writeCorporateInfo(f, corporateSheet, check)
	writeLandInfo(f, landSheet, check)
	writeAdditionalInfo(f, additionalSheet, check)

	if err := f.SaveAs(outPath); err != nil {
		return fmt.Errorf("save xlsx: %w", err)
	}
	return nil
}

func writeCorporateInfo(f *excelize.File, sheet string, check *models.CustomerCheck) {
	headers := []string{"Field", "Value", "Source Document"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDEBF7"}, Pattern: 1}})
	for i := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	sectionStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#E2EFDA"}, Pattern: 1}})

	row := 2
	cell, _ := excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "General Corporate Information")
	_ = f.MergeCell(sheet, cell, "C2")
	_ = f.SetCellStyle(sheet, cell, "C2", sectionStyle)
	row++
	writeField(f, sheet, row, "Client Name", check.Corporate.General.ClientName, "Business License")
	row++
	writeField(f, sheet, row, "Client Type", string(check.Corporate.General.ClientType), "Business License")
	row++
	writeField(f, sheet, row, "Tax Code (MST)", check.Corporate.General.TaxCodeMST, "Business License")
	row++
	writeField(f, sheet, row, "Business License GPKD", string(check.Corporate.General.BusinessLicenseGPKD), "Business License")
	row++
	writeField(f, sheet, row, "Business Address", check.Corporate.General.BusinessAddress, "Business License")
	row++
	var capitalStr string
	if check.Corporate.General.RegisteredShareCapital != nil {
		capitalStr = fmt.Sprintf("%d VND", *check.Corporate.General.RegisteredShareCapital)
	}
	writeField(f, sheet, row, "Registered Share Capital", capitalStr, "Business License")
	row++
	writeField(f, sheet, row, "Customer Type", string(check.Corporate.General.CustomerType), "Business License")
	row++
	writeField(f, sheet, row, "Business Operations", check.Corporate.General.BusinessOperations, "Business License")

	row += 2
	cell, _ = excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "Corporate History")
	_ = f.MergeCell(sheet, cell, fmt.Sprintf("%s%d", "C", row))
	_ = f.SetCellStyle(sheet, cell, fmt.Sprintf("%s%d", "C", row), sectionStyle)
	row++
	var dateStr string
	if check.Corporate.History.IncorporationDate != nil {
		dateStr = check.Corporate.History.IncorporationDate.Format("2006-01-02")
	}
	writeField(f, sheet, row, "Incorporation Date", dateStr, "Business License")
	row++
	writeField(f, sheet, row, "History Description", check.Corporate.History.HistoryDescription, "CIC Report")

	row += 2
	cell, _ = excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "Ownership Information")
	_ = f.MergeCell(sheet, cell, fmt.Sprintf("%s%d", "C", row))
	_ = f.SetCellStyle(sheet, cell, fmt.Sprintf("%s%d", "C", row), sectionStyle)
	row++
	writeField(f, sheet, row, "Owner's Name", check.Corporate.Ownership.OwnersName, "Business License")
	row++
	writeField(f, sheet, row, "Ownership Category", string(check.Corporate.Ownership.OwnershipCategory), "Business License")
	row++
	writeField(f, sheet, row, "Company Director Name", check.Corporate.Ownership.CompanyDirectorName, "ID Check")
	row++
	writeField(f, sheet, row, "Key Decision Maker", check.Corporate.Ownership.KeyDecisionMaker, "ID Check")
}

func writeLandInfo(f *excelize.File, sheet string, check *models.CustomerCheck) {
	headers := []string{"Field", "Value", "Source Document"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDEBF7"}, Pattern: 1}})
	for i := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	sectionStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#E2EFDA"}, Pattern: 1}})

	row := 2
	cell, _ := excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "EVN Information")
	_ = f.MergeCell(sheet, cell, "C2")
	_ = f.SetCellStyle(sheet, cell, "C2", sectionStyle)
	row++
	writeField(f, sheet, row, "Billing Address", check.Land.EVN.BillingAddress, "EVN Bill")
	row++
	writeField(f, sheet, row, "Billing Address Matches Client", string(check.Land.EVN.BillingAddressMatchesClient), "EVN Bill")
	row++
	var amountStr string
	if check.Land.EVN.BillingAmount != nil {
		amountStr = fmt.Sprintf("%d VND", *check.Land.EVN.BillingAmount)
	}
	writeField(f, sheet, row, "Billing Amount", amountStr, "EVN Bill")
	row++
	writeField(f, sheet, row, "Billed Amounts Match Expenses", string(check.Land.EVN.BilledAmountsMatchExpenses), "Financial Statement")

	row += 2
	cell, _ = excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "Land Ownership Information")
	_ = f.MergeCell(sheet, cell, fmt.Sprintf("%s%d", "C", row))
	_ = f.SetCellStyle(sheet, cell, fmt.Sprintf("%s%d", "C", row), sectionStyle)
	row++
	var sourceDoc string
	if check.Land.Ownership.Situation == models.LandOwner {
		sourceDoc = "Land Certificate"
	} else {
		sourceDoc = "Rental Agreement"
	}
	writeField(f, sheet, row, "Situation", string(check.Land.Ownership.Situation), sourceDoc)
	row++
	if check.Land.Ownership.Situation == models.RentalAgreement {
		writeField(f, sheet, row, "Landowner Is Signatory", string(check.Land.Ownership.LandownerIsSignatory), "Rental Agreement")
		row++
		var expirationStr string
		if check.Land.Ownership.LeaseExpirationDate != nil {
			expirationStr = check.Land.Ownership.LeaseExpirationDate.Format("2006-01-02")
		}
		writeField(f, sheet, row, "Lease Expiration Date", expirationStr, "Rental Agreement")
	} else if check.Land.Ownership.Situation == models.LandOwner {
		writeField(f, sheet, row, "Owned Docs Complete", string(check.Land.Ownership.OwnedDocsComplete), "Land Certificate")
	}
}

func writeAdditionalInfo(f *excelize.File, sheet string, check *models.CustomerCheck) {
	headers := []string{"Field", "Value", "Source Document"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDEBF7"}, Pattern: 1}})
	for i := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	sectionStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#E2EFDA"}, Pattern: 1}})

	row := 2
	cell, _ := excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "Site Visit Information")
	_ = f.MergeCell(sheet, cell, "C2")
	_ = f.SetCellStyle(sheet, cell, "C2", sectionStyle)
	row++
	writeField(f, sheet, row, "Company Signboard", string(check.Additional.SiteVisit.CompanySignboard), "Site Visit Photos")

	row += 2
	cell, _ = excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "Finance Information")
	_ = f.MergeCell(sheet, cell, fmt.Sprintf("%s%d", "C", row))
	_ = f.SetCellStyle(sheet, cell, fmt.Sprintf("%s%d", "C", row), sectionStyle)
	row++
	
	// Financial Statement Date
	var financialDateStr string
	if check.Financial.FinancialStatementDate != nil {
		financialDateStr = check.Financial.FinancialStatementDate.Format("2006-01-02")
	}
	writeField(f, sheet, row, "Date of Financial Statements", financialDateStr, "Financial Statement")
	row++
	
	// P&L Section
	writeField(f, sheet, row, "P&L - Total Revenues (30/06/25)", formatMoneyVND(check.Financial.PL.TotalRevenues[0]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Revenues (31/12/24)", formatMoneyVND(check.Financial.PL.TotalRevenues[1]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Revenues (30/6/24)", formatMoneyVND(check.Financial.PL.TotalRevenues[2]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Revenues (31/12/23)", formatMoneyVND(check.Financial.PL.TotalRevenues[3]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Revenues (30/6/23)", formatMoneyVND(check.Financial.PL.TotalRevenues[4]), "Financial Statement")
	row++
	
	writeField(f, sheet, row, "P&L - Total Costs (30/06/25)", formatMoneyVND(check.Financial.PL.TotalCosts[0]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Costs (31/12/24)", formatMoneyVND(check.Financial.PL.TotalCosts[1]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Costs (30/6/24)", formatMoneyVND(check.Financial.PL.TotalCosts[2]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Costs (31/12/23)", formatMoneyVND(check.Financial.PL.TotalCosts[3]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Costs (30/6/23)", formatMoneyVND(check.Financial.PL.TotalCosts[4]), "Financial Statement")
	row++
	
	writeField(f, sheet, row, "P&L - Total Energy Costs (30/06/25)", formatMoneyVND(check.Financial.PL.TotalEnergyCosts[0]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Energy Costs (31/12/24)", formatMoneyVND(check.Financial.PL.TotalEnergyCosts[1]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Energy Costs (30/6/24)", formatMoneyVND(check.Financial.PL.TotalEnergyCosts[2]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Energy Costs (31/12/23)", formatMoneyVND(check.Financial.PL.TotalEnergyCosts[3]), "Financial Statement")
	row++
	writeField(f, sheet, row, "P&L - Total Energy Costs (30/6/23)", formatMoneyVND(check.Financial.PL.TotalEnergyCosts[4]), "Financial Statement")
	row++
	
	// Balance Sheet Section
	writeField(f, sheet, row, "Balance Sheet - Total Assets (30/06/25)", formatMoneyVND(check.Financial.BalanceSheet.TotalAssets[0]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Assets (31/12/24)", formatMoneyVND(check.Financial.BalanceSheet.TotalAssets[1]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Assets (30/6/24)", formatMoneyVND(check.Financial.BalanceSheet.TotalAssets[2]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Assets (31/12/23)", formatMoneyVND(check.Financial.BalanceSheet.TotalAssets[3]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Assets (30/6/23)", formatMoneyVND(check.Financial.BalanceSheet.TotalAssets[4]), "Financial Statement")
	row++
	
	writeField(f, sheet, row, "Balance Sheet - Total Debt (30/06/25)", formatMoneyVND(check.Financial.BalanceSheet.TotalDebt[0]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Debt (31/12/24)", formatMoneyVND(check.Financial.BalanceSheet.TotalDebt[1]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Debt (30/6/24)", formatMoneyVND(check.Financial.BalanceSheet.TotalDebt[2]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Debt (31/12/23)", formatMoneyVND(check.Financial.BalanceSheet.TotalDebt[3]), "Financial Statement")
	row++
	writeField(f, sheet, row, "Balance Sheet - Total Debt (30/6/23)", formatMoneyVND(check.Financial.BalanceSheet.TotalDebt[4]), "Financial Statement")
	row++
	
	// Dynamic Loans Section
	if len(check.Financial.Loans) == 0 {
		writeField(f, sheet, row, "Number of Loans", "0", "CIC Report")
		row++
		writeField(f, sheet, row, "No loans found in CIC report", "", "CIC Report")
	} else {
		writeField(f, sheet, row, "Number of Loans", fmt.Sprintf("%d", len(check.Financial.Loans)), "CIC Report")
		row++
		
		// Export each loan with dynamic numbering
		for i, loan := range check.Financial.Loans {
			loanNum := i + 1
			loanPrefix := fmt.Sprintf("Loan %d", loanNum)
			
			writeField(f, sheet, row, loanPrefix+" - Loan Type", string(loan.LoanType), "CIC Report")
			row++
			writeField(f, sheet, row, loanPrefix+" - Debt Classification", string(loan.DebtClassification), "CIC Report")
			row++
			writeField(f, sheet, row, loanPrefix+" - Outstanding Amount", formatMoneyVNDPtr(loan.OutstandingAmount), "CIC Report")
			row++
			writeField(f, sheet, row, loanPrefix+" - Annual Interest Cost", formatMoneyVNDPtr(loan.AnnualInterestCost), "CIC Report")
			row++
			writeField(f, sheet, row, loanPrefix+" - Annual Amortization", formatMoneyVNDPtr(loan.AnnualAmortization), "CIC Report")
			row++
			var maturityStr string
			if loan.Maturity != nil {
				maturityStr = loan.Maturity.Format("01/02/2006")
			} else {
				maturityStr = "Not available"
			}
			writeField(f, sheet, row, loanPrefix+" - Maturity", maturityStr, "CIC Report")
			row++
			writeField(f, sheet, row, loanPrefix+" - Payment History", loan.PaymentHistory, "CIC Report")
			row++
		}
	}

	row += 2
	cell, _ = excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell, "Check Information")
	_ = f.MergeCell(sheet, cell, fmt.Sprintf("%s%d", "C", row))
	_ = f.SetCellStyle(sheet, cell, fmt.Sprintf("%s%d", "C", row), sectionStyle)
	row++
	var completedAtStr string
	if check.CheckCompletedAt != nil {
		completedAtStr = check.CheckCompletedAt.Format(time.RFC3339)
	}
	writeField(f, sheet, row, "Check Completed At", completedAtStr, "System")
}

func writeField(f *excelize.File, sheet string, row int, fieldName, value, source string) {
	cell1, _ := excelize.CoordinatesToCellName(1, row)
	_ = f.SetCellValue(sheet, cell1, fieldName)
	cell2, _ := excelize.CoordinatesToCellName(2, row)
	_ = f.SetCellValue(sheet, cell2, value)
	cell3, _ := excelize.CoordinatesToCellName(3, row)
	_ = f.SetCellValue(sheet, cell3, source)
}

func formatMoneyVND(amount models.MoneyVND) string {
	return fmt.Sprintf("%.0f", float64(amount))
}

func formatMoneyVNDPtr(amount *models.MoneyVND) string {
	if amount == nil {
		return ""
	}
	return formatMoneyVND(*amount)
}