package api

import (
	"database/sql"
	"net/http"

	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"` // oneof เพื่อกำหนด value ที่ field นี้สามารถเก็บได้
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { // ถ้า ctx.ShouldBindJSON(&req) แล้ว err ไม่เท่ากับ nil ก็จะ res error ไป นั้นเอง
		// ShouldBindJSON ก็คือเอา json body ของ request ใส่ลง req variable
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // res error
		return
	}

	arg := db.CreateAccountParams{
		Owner: req.Owner,
		Currency: req.Currency,
		Balance: 0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok { // ถ้าเป็น error ที่มาจาก postgres จะเข้า if นี้ // ซึ่งหลักๆตรงนี้เราดัก error เมื่อเราสร้าง account ที่ไม่มี user ซึ่งจะทำให้เกิด error ที่ foreign key นั้นเอง
			// log.Println(pqErr.Code.Name())
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation": // foreign_key_violation เมื่อ error เกี่ยวกับ foreign key ใน postgres, unique_violation เมื่อ error เกี่ยวกับ
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"` // ใส่ uri:"id" เพื่อบอก Gin ว่าให้เก็บ uri parameter ชื่อ id ลงที่ field นี้ (ซึ่งมันจะ bind ค่าลงเมื่อคุณเรียก ShouldBindUrl แล้วใส่ pointer ของ getAccountRequest struct)
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows { // ดัก err เมื่อ GetAccount หา account ที่มี id ตามที่กำหนดไม่เจอ
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID int32 `form:"page_id" binding:"required,min=1"` // ใช้ form เนื่องจากเราจะส่ง pageID นี้ผ่าน query string ไม่ใช่ uri parameter // int32 ก็พอ
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { // ใช้ ShouldBindQuery เพื่อเอา value จาก query string ใส่ลง struct var
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAccountsParams{
		Limit: req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}