package api

// เพื่อ test account api function
import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/chanon2000/simplebank/db/mock"
	db "github.com/chanon2000/simplebank/db/sqlc"
	"github.com/chanon2000/simplebank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)



func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	// คนสอนเรียกว่า "สร้าง table driven test set"
	testCases := []struct { // คือ anonymous class เพื่อทำการเก็บ test data
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore) // คือ function ที่เราใส่ mock store เป็น input ซึ่งเราสามารถใช้มันเพื่อ build stub ได้ตาม case ที่แตกต่างกัน
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder) // เพื่อ check output ของ api
	}{
		// ในนี้คือกำหนดแต่ละ scenario
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	// ทำการ refactor code เพื่อให้สามารถ test ได้ในหลายๆ scenarios
	for i := range testCases {
		tc := testCases[i] // tc = test case (หรือก็คือ scenarios ต่างๆ)

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store) // สร้าง stub ใน buildStubs function 

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder) // check response ด้วย checkResponse
		})
	}

	// setup store และ controller โดยใช้ gomock // ซึ่งจะทำให้คุณเห็นว่าการใช้ gomock เขียน unit test นั้นมันทำให้การเขียน unit test ง่ายขึ้นเยอะเลย 
	// 1. สร้าง controller
	// ctrl := gomock.NewController(t) // ctrl = controller // สร้าง controller ก่อน โดยใส่ t object เป็น input
	// defer ctrl.Finish() // ต้อง Finish controller หลังจากทำงานเสร็จด้วย

	// 2. สร้าง new mock store
	// store := mockdb.NewMockStore(ctrl) // สร้าง new mock store โดยต้องใส่ controller เป็น input

	// 3. build stubs ให้กับ mock store // ซึ่งเนื่องจาก getAccount function ที่เราจะ test นั้นมีแค่ GetAccount method ที่ถูกเรียกใช้ เราเลยจะ build stub ในกับแค่ function
	// build stubs โดยเรียก store.EXPECT().GetAccount()
	// store.EXPECT().
	// 		GetAccount(gomock.Any(), gomock.Eq(account.ID)). // GetAccount รับ 2 arg // โดย arg เราใส่ gomock.Any() ไปเลย เพื่อบอกว่า arg แรกเป็นอะไรก็ได้ (เราคงไม่จำเป็นต้อง test การใส่ context หรอก)
	// 		// arg ใส่ id โดยใส่ gomock.Eq(account.ID) เพื่อบอกว่า id ที่ใส่เข้ามาต้องมีค่าเท่ากับ account.ID
	// 		Times(1). // กำหนดว่า GetAccount จะต้องถูกเรียกแค่ครั้งเดียว
	// 		Return(account, nil) // กำหนดว่า GetAccount จะต้อง return เป็น account, nil
	
	// start test server and send request
	// server := NewServer(store) // start server ด้วย mock store
	// เพื่อ testing HTTP API in Go เราไม่จำเป็นต้อง start real HTTP server แต่เราแค่ใช้ recorder feature ของ httptest package เพื่อจด response ของ API request
	// recorder := httptest.NewRecorder() // โดยการเรียก httptest.NewRecorder() เพื่อสร้าง ResponseRecorder

	// url := fmt.Sprintf("/accounts/%d", account.ID) // สร้าง url path ของ pi ที่เราต้องการจะเรียก
	// request, err := http.NewRequest(http.MethodGet, url, nil) // สร้าง new http request ด้วย NewRequest ตามด้วยใส่ method, url, body เป็น input โดยในที่นี้ใส่ nil เป็น body
	// require.NoError(t, err) 

	// server.router.ServeHTTP(recorder, request) // ทำการส่ง request ที่สร้างนี้โดย ServerHTTP โดยใส่ recorder และ request // ซึ่ง response ที่ได้จะอยู่ใน recorder ที่เราใส่เป็น input นี้

}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) { // เนื่องจาก recorder.Body มี type เป็น *bytes.Buffer
	data, err := io.ReadAll(body) // io.ReadAll เพื่ออ่าน body ของ response หรือก็คือใน recorder
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount) // Unmarshal data ไปให้กับ gotAccount
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}