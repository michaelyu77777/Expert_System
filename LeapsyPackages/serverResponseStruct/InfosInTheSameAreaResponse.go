package serverResponseStruct

import (
	"leapsy.com/packages/serverDataStruct"
)

// Response-取得所有線上Info
type InfosInTheSameAreaResponse struct {
	Command       int                      `json:"command"`
	CommandType   int                      `json:"commandType"`
	ResultCode    int                      `json:"resultCode"`
	Results       string                   `json:"results"`
	TransactionID string                   `json:"transactionID"`
	Info          []*serverDataStruct.Info `json:"info"`
	//Device        []*Device `json:"device"`
}
