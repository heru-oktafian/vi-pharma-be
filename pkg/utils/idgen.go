package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// generateIDWithPrefix generates a 15-character unique identifier with a 3-character prefix
func generateIDWithPrefix(prefix string) string {
	const randomPartLength = 12 // 3 + 12 = 15 characters total

	result := make([]byte, randomPartLength)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < randomPartLength; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			result[i] = charset[i%len(charset)]
		} else {
			result[i] = charset[num.Int64()]
		}
	}

	return prefix + string(result)
}

// GenerateUserID generates a 15-character unique identifier for users (prefix "USR")
func GenerateUserID() string {
	return generateIDWithPrefix("USR")
}

// GenerateBranchID generates a 15-character unique identifier for branches (prefix "CBG")
func GenerateBranchID() string {
	return generateIDWithPrefix("CBG")
}

// GenerateUserBranchID generates a 15-character unique identifier for user_branches (prefix "UBR")
func GenerateUserBranchID() string {
	return generateIDWithPrefix("UBR")
}

// GenerateUnitID generates a 15-character unique identifier for units (prefix "UNT")
func GenerateUnitID() string {
	return generateIDWithPrefix("UNT")
}

// GenerateCategoryID generates a 15-character unique identifier for categories (prefix "CAT")
func GenerateCategoryID() string {
	return generateIDWithPrefix("CAT")
}

// GenerateSupplierID generates a 15-character unique identifier for suppliers/PBF (prefix "SUP")
func GenerateSupplierID() string {
	return generateIDWithPrefix("SUP")
}

// GenerateDoctorID generates a 15-character unique identifier for doctors (prefix "DOC")
func GenerateDoctorID() string {
	return generateIDWithPrefix("DOC")
}

// GenerateMedicineID generates a 15-character unique identifier for medicines (prefix "MED")
func GenerateMedicineID() string {
	return generateIDWithPrefix("MED")
}

// GenerateMedicineUnitID generates a 15-character unique identifier for medicine units (prefix "MNU")
func GenerateMedicineUnitID() string {
	return generateIDWithPrefix("MNU")
}

// GenerateMedicinePriceID generates a 15-character unique identifier for medicine prices (prefix "MPR")
func GenerateMedicinePriceID() string {
	return generateIDWithPrefix("MPR")
}

// GenerateBatchID generates a 15-character unique identifier for medicine batches (prefix "BAT")
func GenerateBatchID() string {
	return generateIDWithPrefix("BAT")
}

// GeneratePurchaseInvoiceID generates a 15-character unique identifier for purchase invoices (prefix "PIV")
func GeneratePurchaseInvoiceID() string {
	return generateIDWithPrefix("PIV")
}

// GeneratePurchaseInvoiceItemID generates a 15-character unique identifier for invoice items (prefix "PII")
func GeneratePurchaseInvoiceItemID() string {
	return generateIDWithPrefix("PII")
}

// GenerateStockMutationID generates a 15-character unique identifier for stock mutations (prefix "MUT")
func GenerateStockMutationID() string {
	return generateIDWithPrefix("MUT")
}

// GenerateStockOpnameID generates a 15-character unique identifier for stock opname (prefix "SOP")
func GenerateStockOpnameID() string {
	return generateIDWithPrefix("SOP")
}

// GenerateStockOpnameItemID generates a 15-character unique identifier for opname items (prefix "SOI")
func GenerateStockOpnameItemID() string {
	return generateIDWithPrefix("SOI")
}

// GeneratePurchaseOrderID generates a 15-character unique identifier for purchase orders / SP (prefix "POR")
func GeneratePurchaseOrderID() string {
	return generateIDWithPrefix("POR")
}

// GeneratePurchaseOrderItemID generates a 15-character unique identifier for purchase order items (prefix "POI")
func GeneratePurchaseOrderItemID() string {
	return generateIDWithPrefix("POI")
}

// GenerateSaleID generates a 15-character unique identifier for POS sales transactions (prefix "SAL")
func GenerateSaleID() string {
	return generateIDWithPrefix("SAL")
}

// GenerateSaleItemID generates a 15-character unique identifier for sale items (prefix "SIT")
func GenerateSaleItemID() string {
	return generateIDWithPrefix("SIT")
}
