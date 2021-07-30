package serverDataStruct

// 客戶端Info
type Info struct {
	AccountPointer *Account `json:"account"` //使用者帳戶資料
	DevicePointer  *Device  `json:"device"`  //使用者登入密碼
}
