package analysis

// DocumentSource represents the source document type
type DocumentSource string

const (
	SourceBusinessLicense   DocumentSource = "business_license"
	SourceEVNBill           DocumentSource = "evn_bill"
	SourceLandCertificate   DocumentSource = "land_certificate"
	SourceIDCheck           DocumentSource = "id_check"
	SourceFinancialStatement DocumentSource = "financial_statement"
	SourceSiteVisitPhotos   DocumentSource = "site_visit_photos"
	SourceCICReport         DocumentSource = "cic_report"
	SourceCICReport2        DocumentSource = "cic_report_2"
	SourceUnknown           DocumentSource = "unknown"
)
