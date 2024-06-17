package api

import (
	"github.com/go-playground/validator/v10" // ใช้ version 10
	"github.com/chanon2000/simplebank/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool { // validator.Func คือ type
	// validator.FieldLevel เป็น interface ที่จะมี helper functions มากมาย
	if currency, ok := fieldLevel.Field().Interface().(string); ok { // .Field() เพื่อ get value จาก field // .Interface() เพื่อทำให้ value มี type เป็น empty interface{} // .(string) จากนั้นก็ convert เป็น string ซึ่งก็จะ return currency value กับ ok value
		return util.IsSupportedCurrency(currency)
	}
	return false
}
