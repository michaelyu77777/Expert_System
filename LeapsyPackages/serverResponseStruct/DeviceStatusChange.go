package serverResponseStruct

import (
	"leapsy.com/packages/serverDataStruct"
)

// Broadcast(廣播)-裝置狀態改變
type DeviceStatusChange struct {
	//指令
	Command     int                       `json:"command"`
	CommandType int                       `json:"commandType"`
	Device      []serverDataStruct.Device `json:"device"`
}
