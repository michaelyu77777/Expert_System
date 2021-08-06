package model

// "leapsy.com/times"

// Device - 警報紀錄
type Device struct {
	DeviceID    string // 裝置ID
	DeviceBrand string // 裝置品牌(ID可能重複)
	DeviceType  int    // 裝置類型
	Area        []int  // 裝置場域
	DeviceName  string // 裝置名稱
}
