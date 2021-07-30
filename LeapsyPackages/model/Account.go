package model

import (
	"time"
	// "leapsy.com/packages/networkHub"
)

// "leapsy.com/times"

// Device - 警報紀錄
type Account struct {
	UserID               string    // 使用者登入帳號
	UserPassword         string    // 使用者登入密碼
	UserName             string    // 使用者名稱
	IsExpert             int       // 是否為專家帳號
	IsFrontline          int       // 是否為一線人員帳號
	Area                 []int     // 專家場域
	Pic                  string    // 帳號頭像
	VerificationCodeTime time.Time // 最後取得驗證碼之時間
}

// PrimitiveM - 轉成primitive.M
/*
 * @return primitive.M returnPrimitiveM 回傳結果
 */
// func (modelAccount Account) Account() (networkHubAccount networkHub.Account) {

// 	networkHubAccount.UserID = modelAccount.UserID

// 	return
// }

// modelAccount := mongoDB.Find....
// modelAccount.Account()
