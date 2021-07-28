package model

// "leapsy.com/times"

// Device - 警報紀錄
type Account struct {
	UserID       string // 使用者登入帳號
	UserPassword string // 使用者登入密碼
	UserName     string // 使用者名稱
	IsExpert     int    // 是否為專家帳號
	IsFrontline  int    // 是否為一線人員帳號
	Area         []int  // 專家場域
	Pic          string // 帳號頭像
}

// var (
// 	alertRecordToECSDeviceMap = map[string]string{ // alertRecordToECSDeviceMap - 警報紀錄與環控資料庫警報紀錄欄位對照
// 		`AlertEventID`:   `ALERTEVENTID`,   // 警報編號 int
// 		`AlertEventTime`: `ALERTEVENTTIME`, // 日期時間	datetime
// 		`VarTag`:         `VARTAG`,         // 點名稱	nvarchar(50)
// 		`Comment`:        `COMMENT`,        // 說明	nvarchar(max)
// 		`AlertType`:      `ALERTTYPE`,      // 警報群組	int
// 		`LineText`:       `LINETEXT`,       // 行文字	nvarchar(max)
// 	}
// )

// getMappedToECSDeviceFieldName - 取得警報紀錄對應的環控警報紀錄欄位名
/**
 * @param  string alertRecordFieldName 警報紀錄欄位名
 * @return string 環控警報紀錄欄位名
 */
// func getMappedToECSDeviceFieldName(alertRecordFieldName string) string {
// 	return alertRecordToECSDeviceMap[alertRecordFieldName] // 回傳警報紀錄對應的環控警報紀錄欄位名
// }

// Device - 將ECSDevice轉成Device
/**
 * @return Device 警報紀錄
 */
// func (ecsDevice ESDevice) Device() (alertRecord Device) {

// 	valueOfECSDevice := reflect.ValueOf(ecsDevice)        // 環控警報紀錄的值
// 	typeOfDevice := reflect.TypeOf(alertRecord)           // 警報紀錄的資料型別
// 	valueOfDevice := reflect.ValueOf(&alertRecord).Elem() // 警報紀錄的值

// 	for index := 0; index < typeOfDevice.NumField(); index++ { // 針對警報紀錄每一個欄位

// 		alertRecordFieldName := typeOfDevice.Field(index).Name                    // 警報紀錄欄位名
// 		alertRecordFieldValue := valueOfDevice.Field(index)                       // 警報紀錄欄位值
// 		ecsDeviceFieldName := getMappedToECSDeviceFieldName(alertRecordFieldName) // 環控警報紀錄欄位名

// 		if `` != ecsDeviceFieldName { // 若有對應的環控警報紀錄欄位名

// 			ecsDeviceFieldValue := valueOfECSDevice.FieldByName(ecsDeviceFieldName) // 環控警報紀錄欄位值

// 			switch typeOfDevice.Field(index).Type.String() { // 若警報紀錄欄位型別為

// 			case `int`: // 整數

// 				integer, strconvAtoiError := strconv.Atoi(ecsDeviceFieldValue.String()) // 將環控警報紀錄欄位值字串轉為整數

// 				logings.SendLog(
// 					[]string{`環控警報紀錄欄位 %s 值轉成整數`},
// 					[]interface{}{ecsDeviceFieldName},
// 					strconvAtoiError,
// 					logrus.InfoLevel,
// 				)

// 				alertRecordFieldValue.SetInt(int64(integer)) // 設定警報紀錄欄位值為環控警報紀錄欄位轉化後的整數值

// 			case `time.Time`: // 時間
// 				alertRecordFieldValue.Set(reflect.ValueOf(times.ALERTEVENTTIMEStringToTime(ecsDeviceFieldValue.String()))) // 設定警報紀錄欄位值為環控警報紀錄欄位的時間值

// 			case `bool`: // 布林值
// 				alertRecordFieldValue.SetBool(ecsDeviceFieldValue.Bool()) // 設定警報紀錄欄位值為環控警報紀錄欄位的布林值

// 			default: // 預設
// 				alertRecordFieldValue.SetString(ecsDeviceFieldValue.String()) // 設定警報紀錄欄位值為環控警報紀錄欄位的字串值

// 			}

// 		}

// 	}

// 	return // 回傳
// }

// PrimitiveM - 轉成primitive.M
/*
 * @return primitive.M returnPrimitiveM 回傳結果
 */
// func (alertRecord Device) PrimitiveM() (returnPrimitiveM primitive.M) {

// 	returnPrimitiveM = bson.M{
// 		`alerteventid`:   alertRecord.AlertEventID,   // 警報編號
// 		`alerttype`:      alertRecord.AlertType,      // 警報群組
// 		`alerteventtime`: alertRecord.AlertEventTime, // 日期時間
// 		`vartag`:         alertRecord.VarTag,         // 點名稱
// 		`comment`:        alertRecord.Comment,        // 說明
// 		`linetext`:       alertRecord.LineText,       // 行文字
// 		`isread`:         alertRecord.IsRead,         // 是否已讀
// 		`ishidden`:       alertRecord.IsHidden,       // 是否隱藏
// 	}

// 	return
// }
