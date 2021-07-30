package serverResponseStruct

import (
	"leapsy.com/packages/serverDataStruct"
)

// Response-取得我的裝置
type MyDeviceResponse struct {
	Command       int                     `json:"command"`
	CommandType   int                     `json:"commandType"`
	ResultCode    int                     `json:"resultCode"`
	Results       string                  `json:"results"`
	TransactionID string                  `json:"transactionID"`
	Device        serverDataStruct.Device `json:"device"`
}
