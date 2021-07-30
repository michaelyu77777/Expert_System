package serverDataStruct

// 裝置
type Device struct {
	DeviceID    string   `json:"deviceID"`    //裝置ID
	DeviceBrand string   `json:"deviceBrand"` //裝置品牌(怕平板裝置的ID會重複)
	DeviceType  int      `json:"deviceType"`  //裝置類型
	Area        []int    `json:"area"`        //場域
	AreaName    []string `json:"areaName"`    //場域名稱
	DeviceName  string   `json:"deviceName"`  //裝置名稱
	// 以下為可重設值
	Pic          string `json:"pic"`          //裝置截圖
	OnlineStatus int    `json:"onlineStatus"` //在線狀態
	DeviceStatus int    `json:"deviceStatus"` //設備狀態
	CameraStatus int    `json:"cameraStatus"` //相機狀態
	MicStatus    int    `json:"micStatus"`    //麥克風狀態
	RoomID       int    `json:"roomID"`       //房號
}
