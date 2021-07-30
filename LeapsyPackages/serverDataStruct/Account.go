package serverDataStruct

import "time"

// 帳戶
type Account struct {
	UserID       string   `json:"userID"`       // 使用者登入帳號
	UserPassword string   `json:"userPassword"` // 使用者登入密碼
	UserName     string   `json:"userName"`     // 使用者名稱
	IsExpert     int      `json:"isExpert"`     // 是否為專家帳號:1是,2否
	IsFrontline  int      `json:"isFrontline"`  // 是否為一線人員帳號:1是,2否
	Area         []int    `json:"area"`         // 專家所屬場域代號
	AreaName     []string `json:"areaName"`     // 專家所屬場域名稱
	Pic          string   `json:"pic"`          // 帳號頭像

	// (不回傳給client)
	// verificationCodeTime time.Time // 最後取得驗證碼之時間
	VerificationCodeTime time.Time `json:"-"` // 最後取得驗證碼之時間
}
