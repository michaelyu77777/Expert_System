package serverResponseStruct

import (
	"leapsy.com/packages/serverDataStruct"
)

// Response-取得我的帳戶
type MyAccountResponse struct {
	Command       int                      `json:"command"`
	CommandType   int                      `json:"commandType"`
	ResultCode    int                      `json:"resultCode"`
	Results       string                   `json:"results"`
	TransactionID string                   `json:"transactionID"`
	Account       serverDataStruct.Account `json:"account"`
}
