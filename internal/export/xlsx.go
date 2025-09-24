package export

import (
	"fmt"

	"extraction/internal/types"
	"github.com/xuri/excelize/v2"
)

func WriteResults(results []types.FileResult, outPath string) error {
	f := excelize.NewFile()
	sheet := f.GetSheetName(f.GetActiveSheetIndex())
	headers := []string{"SourceURL", "LocalPath", "FileName", "FileType", "Error", "ExtractedText"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for rowIdx, r := range results {
		row := rowIdx + 2
		cells := []any{r.SourceURL, r.LocalPath, r.FileName, r.FileType, r.Error, r.ExtractedText}
		for colIdx, v := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	if err := f.SaveAs(outPath); err != nil {
		return fmt.Errorf("save xlsx: %w", err)
	}
	return nil
}


