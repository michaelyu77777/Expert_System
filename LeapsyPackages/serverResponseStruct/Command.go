package serverResponseStruct

// 客戶端送來的 Command
type Command struct {
	// 指令
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	TransactionID string `json:"transactionID"` //分辨多執行緒順序不同的封包

	// 登入Info
	UserID       string `json:"userID"`       //使用者登入帳號
	UserPassword string `json:"userPassword"` //使用者登入密碼

	// 裝置Info
	DeviceID    string `json:"deviceID"`    //裝置ID
	DeviceBrand string `json:"deviceBrand"` //裝置品牌(怕平板裝置的ID會重複)
	DeviceType  int    `json:"deviceType"`  //裝置類型

	Area          []int    `json:"area"`           //場域代號
	AreaName      []string `json:"areaName"`       //場域名稱
	DeviceName    []string `json:"deviceName"`     //裝置名稱
	Pic           string   `json:"pic"`            //裝置截圖(求助截圖)
	OnlineStatus  int      `json:"onlineStatus"`   //在線狀態
	DeviceStatus  int      `json:"deviceStatus"`   //設備狀態
	CameraStatus  int      `json:"cameraStatus"`   //相機狀態
	ThermalStatus int      `jason:"thermalStatus"` //熱呈像狀態
	MicStatus     int      `json:"micStatus"`      //麥克風狀態
	RoomID        int      `json:"roomID"`         //房號

	// 加密後字串
	AreaEncryptionString string `json:"areaEncryptionString"` //場域代號加密字串
	StringToEncryption   string `json:"stringToEncryption"`   //要加密的字串
	StringToDecryption   string `json:"stringToDecryption"`   //要加密的字串
}
