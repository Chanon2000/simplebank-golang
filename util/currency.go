package util
// Implement logic check currency ว่า supported หรือไม่ใน file นี้
// Constants for all supported currencies
const ( // ปัจจบัน support แค่ 3 currency นี้ไปก่อน
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
