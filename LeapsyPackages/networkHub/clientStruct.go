package networkHub

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"../configurations"
	"../logings"
	"../network"
	"../paths"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/juliangruber/go-intersect"
)

// client - 客戶端
type client struct {
	connectionPointerMutexPointer *sync.RWMutex // 讀寫鎖
	connectionPointer             *net.Conn     // 連線指標

	outputChannel chan websocketData // 輸出通道

	inputChannel chan websocketData // 輸入通道

	fileExtensionMutexPointer *sync.RWMutex // 讀寫鎖
	fileExtension             string        // 副檔名
}

// initialize - 初始化
func (clientPointer *client) initialize() {

	if nil != clientPointer { // 若指標不為空

		if nil == clientPointer.connectionPointerMutexPointer { // 若沒讀寫鎖
			var connectionPointerMutex sync.RWMutex                               // 讀寫鎖
			clientPointer.connectionPointerMutexPointer = &connectionPointerMutex // 儲存
		}

		if nil == clientPointer.fileExtensionMutexPointer { // 若沒讀寫鎖
			var fileExtensionMutex sync.RWMutex                           // 讀寫鎖
			clientPointer.fileExtensionMutexPointer = &fileExtensionMutex // 儲存
		}

	}

}

// setConnectionPointer - 設定連線指標
func (clientPointer *client) setConnectionPointer(connectionPointer *net.Conn) {

	if nil != clientPointer { // 若指標不為空
		clientPointer.initialize()                           // 初始化
		clientPointer.connectionPointerMutexPointer.Lock()   // 鎖寫
		clientPointer.connectionPointer = connectionPointer  // 客戶端的連線
		clientPointer.connectionPointerMutexPointer.Unlock() // 解鎖寫
	}

}

// getConnectionPointer - 取得連線指標
func (clientPointer *client) getConnectionPointer() (returnConnectionPointer *net.Conn) {

	if nil != clientPointer { // 若指標不為空
		clientPointer.initialize()                                // 初始化
		clientPointer.connectionPointerMutexPointer.RLock()       // 鎖讀
		returnConnectionPointer = clientPointer.connectionPointer // 客戶端的連線指標
		clientPointer.connectionPointerMutexPointer.RUnlock()     // 解鎖讀
	}

	return // 回傳
}

// setFileExtension - 設定副檔名
func (clientPointer *client) setFileExtension(fileExtension string) {

	if nil != clientPointer { // 若指標不為空
		clientPointer.initialize()                       // 初始化
		clientPointer.fileExtensionMutexPointer.Lock()   // 鎖寫
		clientPointer.fileExtension = fileExtension      // 儲存附檔名
		clientPointer.fileExtensionMutexPointer.Unlock() // 解鎖寫
	}

}

// getFileExtension - 取得副檔名
func (clientPointer *client) getFileExtension() (returnFileExtension string) {

	if nil != clientPointer { // 若指標不為空
		clientPointer.initialize()                        // 初始化
		clientPointer.fileExtensionMutexPointer.RLock()   // 鎖讀
		returnFileExtension = clientPointer.fileExtension // 取得附檔名
		clientPointer.fileExtensionMutexPointer.RUnlock() // 解鎖讀
	}

	return // 回傳
}

// getInputWebsocketDataFromConnection - 取得連線輸入的websocket資料
/**
 * @param  *net.Conn connectionPointer  連線指標
 * @param  websocketData websocketData        websocket資料
 * @return websocketData returnWebsocketData 回傳websocket資料
 * @return bool returnIsSuccess 回傳是否成功
 */
func getInputWebsocketDataFromConnection(connectionPointer *net.Conn) (returnWebsocketData websocketData,
	returnIsSuccess bool) {

	if nil != connectionPointer { // 若指標不為空

		connection := *connectionPointer // 客戶端的連線

		dataBytes, wsOpCode, wsutilReadClientDataError := wsutil.ReadClientData(connection) // 讀取連線傳來的資料到緩存

		formatSlice := `%s %s 接收 %s `
		defaultArgs := append(network.GetAliasAddressPair(connection.LocalAddr().String()), connection.RemoteAddr().String())

		// 取得記錄器格式字串與參數
		formatString, args := logings.GetLogFuncFormatAndArguments(
			[]string{formatSlice + `的資料為 %v `},
			append(defaultArgs, dataBytes),
			wsutilReadClientDataError,
		)

		// formatString, args := logings.GetLogFuncFormatAndArguments(
		// 	[]string{`取得資料 %v %v`},
		// 	[]interface{}{dataBytes,dataBytes},
		// 	wsutilReadClientDataError,
		// )

		if nil != wsutilReadClientDataError { // 若讀取連線傳來的資料到緩存錯誤
			logger.Warnf(formatString, args...) // 記錄警告
			return                              // 回傳
		}

		if ws.OpText == wsOpCode {

			dataString := string(dataBytes) // 資料字串

			// 取得記錄器格式字串與參數
			formatString, args = logings.GetLogFuncFormatAndArguments(
				[]string{formatSlice + `的字串為 %s `},
				append(defaultArgs, dataString),
				nil,
			)

		}

		logger.Infof(formatString, args...) // 記錄資訊

		returnWebsocketData = websocketData{
			wsOpCode:  wsOpCode,
			dataBytes: dataBytes,
		} // 回傳收到的資料
		returnIsSuccess = true // 回傳成功

	}

	return // 回傳
}

// giveOutputWebsocketDataToConnection - 對連線輸出websocket資料
/**
 * @param  *net.Conn connectionPointer  連線指標
 * @param  websocketData websocketData        websocket資料
 * @return bool returnIsSuccess 回傳是否成功
 */
func giveOutputWebsocketDataToConnection(connectionPointer *net.Conn, outputWebsocketData websocketData) (returnIsSuccess bool) {

	if nil != connectionPointer { // 若指標不為空

		connection := *connectionPointer // 客戶端的連線

		wsOpCode := outputWebsocketData.wsOpCode
		dataBytes := outputWebsocketData.dataBytes

		wsutilWriteServerMessageError := wsutil.WriteServerMessage(connection, wsOpCode, dataBytes) // 輸出資料到連線

		formatSlice := `%s %s 傳給 %s `
		defaultArgs := append(
			network.GetAliasAddressPair(connection.LocalAddr().String()),
			connection.RemoteAddr().String(),
		)

		// 取得記錄器格式字串與參數
		formatString, args := logings.GetLogFuncFormatAndArguments(
			[]string{formatSlice + `的資料為 %v`},
			append(defaultArgs, dataBytes),
			wsutilWriteServerMessageError,
		)

		if nil != wsutilWriteServerMessageError { // 若輸出資料到連線錯誤，則中斷循環處理通道接收
			logger.Warnf(formatString, args...) // 記錄警告
			return                              // 回傳
		}

		if ws.OpText == wsOpCode {
			// 取得記錄器格式字串與參數
			formatString, args = logings.GetLogFuncFormatAndArguments(
				[]string{formatSlice + `的字串為 %s`},
				append(defaultArgs, string(dataBytes)),
				nil,
			)
		}

		go logger.Infof(formatString, args...) // 記錄資訊

		returnIsSuccess = true // 回傳成功
	}

	return // 回傳

}

// 客戶端 Command

type Command struct {
	// 指令
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	TransactionID string `json:"transactionID"` //分辨多執行緒順序不同的封包

	// 登入Info
	UserID       string `json:"userID"`       //使用者登入帳號
	UserPassword string `json:"userPassword"` //使用者登入密碼
	// FunctionNumber int    `json:"functionNumber"` //是否為帳號驗證功能: 只驗證帳號是否存在＋寄驗證信功能
	// IDPWIsRequired int    `json:"IDPWIsRequired"` //是否為必須登入模式: 裝置需要登入才能使用其他功能

	// 裝置Info
	DeviceID    string `json:"deviceID"`    //裝置ID
	DeviceBrand string `json:"deviceBrand"` //裝置品牌(怕平板裝置的ID會重複)
	DeviceType  int    `json:"deviceType"`  //裝置類型

	Area         []int    `json:"area"`         //場域代號
	AreaName     []string `json:"areaName"`     //場域名稱
	DeviceName   string   `json:"deviceName"`   //裝置名稱
	Pic          string   `json:"pic"`          //裝置截圖(求助截圖)
	OnlineStatus int      `json:"onlineStatus"` //在線狀態
	DeviceStatus int      `json:"deviceStatus"` //設備狀態
	CameraStatus int      `json:"cameraStatus"` //相機狀態
	MicStatus    int      `json:"micStatus"`    //麥克風狀態
	RoomID       int      `json:"roomID"`       //房號

	//測試用
	//Argument1 string `json:"argument1"`
	//Argument2 string `json:"argument2"`
}

// 心跳包 Response
type Heartbeat struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 客戶端資訊
type Info struct {
	// UserID         string  `json:"userID"`         //使用者登入帳號
	// UserPassword   string  `json:"userPassword"`   //使用者登入密碼
	// IDPWIsRequired bool    `json:"IDPWIsRequired"` //是否需要登入才能操作
	Account *Account `json:"account"` //使用者帳戶資料
	Device  *Device  `json:"device"`  //使用者登入密碼
}

// 帳戶資訊
type Account struct {
	UserID       string   `json:"userID"`       // 使用者登入帳號
	UserPassword string   `json:"userPassword"` // 使用者登入密碼
	UserName     string   `json:"userName"`     // 使用者名稱
	IsExpert     int      `json:"isExpert"`     // 是否為專家帳號:1是,2否
	IsFrontline  int      `json:"isFrontline"`  // 是否為一線人員帳號:1是,2否
	Area         []int    `json:"area"`         // 專家所屬場域代號
	AreaName     []string `json:"areaName"`     // 專家所屬場域名稱
	Pic          string   `json:"pic"`          // 帳號頭像
}

// 裝置資訊
type Device struct {
	DeviceID     string   `json:"deviceID"`     //裝置ID
	DeviceBrand  string   `json:"deviceBrand"`  //裝置品牌(怕平板裝置的ID會重複)
	DeviceType   int      `json:"deviceType"`   //裝置類型
	Area         []int    `json:"area"`         //場域
	AreaName     []string `json:"areaName"`     //場域名稱
	DeviceName   string   `json:"deviceName"`   //裝置名稱
	Pic          string   `json:"pic"`          //裝置截圖
	OnlineStatus int      `json:"onlineStatus"` //在線狀態
	DeviceStatus int      `json:"deviceStatus"` //設備狀態
	CameraStatus int      `json:"cameraStatus"` //相機狀態
	MicStatus    int      `json:"micStatus"`    //麥克風狀態
	RoomID       int      `json:"roomID"`       //房號
}

// 客戶端資訊:(為了Logger不印出密碼)
type InfoForLogger struct {
	UserID *string `json:"userID"` //使用者登入帳號
	Device *Device `json:"device"` //使用者登入密碼
}

// 登入 - Response -
type LoginResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// Response: 取得我的Account清單
type MyAccountResponse struct {
	Command       int     `json:"command"`
	CommandType   int     `json:"commandType"`
	ResultCode    int     `json:"resultCode"`
	Results       string  `json:"results"`
	TransactionID string  `json:"transactionID"`
	Account       Account `json:"account"`
}

// 取得我的Device清單 -Response-
type MyDeviceResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
	Device        Device `json:"device"`
}

// 取得所有裝置清單 - Response -
type DevicesResponse struct {
	Command       int     `json:"command"`
	CommandType   int     `json:"commandType"`
	ResultCode    int     `json:"resultCode"`
	Results       string  `json:"results"`
	TransactionID string  `json:"transactionID"`
	Info          []*Info `json:"info"`
	//Device        []*Device `json:"device"`
}

// 求助 - Response -
type HelpResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 裝置狀態改變 - Broadcast(廣播) -
type DeviceStatusChange struct {
	//指令
	Command     int      `json:"command"`
	CommandType int      `json:"commandType"`
	Device      []Device `json:"device"`
}

type DeviceStatusChangeByPointer struct {
	//指令
	Command     int       `json:"command"`
	CommandType int       `json:"commandType"`
	Device      []*Device `json:"device"`
}

// Map-連線/登入資訊
var clientInfoMap = make(map[*client]*Info)

// 所有線上裝置清單(棄用！？)
var onlineDeviceList []*Device

// 所有裝置清單
var allDeviceList = []*Device{}

// 所有帳號清單
var allAccountList = []*Account{}

// 所有 area number 對應到 area name 名稱
var areaNumberNameMap = make(map[int]string)

// 加密解密KEY(AES加密)（key 必須是 16、24 或者 32 位的[]byte）
const key_AES = "RB7Wfa$WHssV4LZce6HCyNYpdPPlYnDn" //32位數

// 指令代碼
const CommandNumberOfLogout = 8
const CommandNumberOfBroadcastingInArea = 10
const CommandNumberOfBroadcastingInRoom = 11

// 指令類型代碼
const CommandTypeNumberOfAPI = 1         // 客戶端-->Server
const CommandTypeNumberOfAPIResponse = 2 // Server-->客戶端
const CommandTypeNumberOfBroadcast = 3   // Server廣播
const CommandTypeNumberOfHeartbeat = 4   // 心跳包

// 結果代碼
const (
	ResultCodeSuccess = 0 // 成功結果代碼
	ResultCodeFail    = 1 // 失敗結果代碼
)

// 連線逾時時間:
//const timeout = 30
var timeout = time.Duration(configurations.GetConfigPositiveIntValueOrPanic(`local`, `timeout`)) // 轉成time.Duration型態，方便做時間乘法

// 房間號(總計)
var roomID = 0

// 基底: Response Json
var baseResponseJsonString = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"}`
var baseResponseJsonStringExtend = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"` // 可延展的

// 基底: Broadcasting Json
// var baseBroadCastingJsonString = `{"command":%d,"commandType":%d,"device":%+v}`

// 基底: 收到Json格式
var baseLoggerServerReceiveJsonString = `伺服器收到Json:%s。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`
var baseLoggerMissFieldsString = `<檢查Json欄位>:失敗-Json欄位不齊全,以下欄位不齊全:%s。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`
var baseLoggerReceiveJsonRetrunErrorString = `伺服器內部jason轉譯失敗，當轉譯(Json欄位不齊全,以下欄位不齊全:%s)時。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d` // Server轉譯json出錯

// 基底: 登入時發生的狀況Response(尚未建立Account/Device資料時)
var baseLoggerWhenLoginString = `<指令:%s-%s>。客戶端Command:%+v、此連線Pointer:%p、連線清單:%+v、所有裝置:%+v、線上裝置:%+v、房號已取到:%d`

// 基底: 共用(指令成功、指令失敗、失敗原因、廣播、指令結束)
var baseLoggerInfoCommonMessage = `指令<%s>:%s。Command:%+v、帳號:%+v、裝置:%+v、連線:%p、連線清單:%+v、裝置清單:%+v、,房號已取到:%d` // 普通紀錄

var baseLoggerServerReceiveCommnad = `收到<%s>指令。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`                          // 收到指令
var baseLoggerNotLoggedInWarnString = `指令<%s>失敗:連線尚未登入。客戶端Command:%+v、此連線帳號:%+s、此連線裝置ID:%+s、此連線裝置Brand:%+s、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d` // 連線尚未登入          // 失敗:連線尚未登入
var baseLoggerNotCompletedFieldsWarnString = `指令<%s>失敗:以下欄位不齊全:%s。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`       // 失敗:欄位不齊全
var baseLoggerSuccessString = `指令<%s>成功。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`                                 // 成功
var baseLoggerInfoBroadcastInArea = `指令<%s>-(場域)廣播-狀態變更。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、,房號已取到:%d`                // 場域廣播
var baseLoggerWarnReasonString = `指令<%s>失敗:%s。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`                                               // 失敗:原因
var baseLoggerErrorJsonString = `指令<%s>jason轉譯出錯。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、房號已取到:%d`               // Server轉譯json出錯

// 基底: 連線逾時專用
var baseLoggerInfoForTimeout = `<偵測連線逾時>%s，timeout=%d。此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、,房號已取到:%d` // 場域廣播（逾時 timeout)
var baseLoggerWarnForTimeout = `<偵測連線逾時>%s，timeout=%d。此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、,房號已取到:%d` // 主動告知client（逾時 timeout)
var baseLoggerErrorForTimeout = `<偵測連線逾時>%s，timeout=%d。此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、房號已取到:%d` // Server轉譯json出錯
var baseLoggerInfoForTimeoutWithoutNilDevice = `連線已登入，並已從清單移除裝置，判定為<登出>，不再繼續偵測逾時，timeout=%d。此連線帳號:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`

// 待補:(定時)更新DB裝置清單，好讓後台增加裝置時，也可以再依定時間內同步補上，但要再確認，是否有資料更新不一致問題
func UpdateAllDevicesList() {
	// 待補:固定時間內
	// for {
	importAllDevicesList()
	// }
}

// 待補:(定時)更新DB裝置清單，好讓後台增加帳號時，也可以再依定時間內同步補上，但要再確認，是否有資料更新不一致問題
func UpdateAllAccountList() {
	// 待補:固定時間內
	// for {
	importAllAccountList()
	// }
}

// 待補：定時更新場域內容
func UpdateAllAreaMap() {
	importAllAreasNameToMap()
}

// 匯入所有帳號
func importAllAccountList() {

	picExpertA := getAccountPicString("pic/picExpertA.txt")
	picExpertB := getAccountPicString("pic/picExpertB.txt")
	picFrontline := getAccountPicString("pic/picFrontline.txt")
	picDefault := getAccountPicString("pic/picDefault.txt")

	//專家帳號 場域A
	accountExpertA := Account{
		UserID:       "expertA@leapsyworld.com",
		UserPassword: "expertA@leapsyworld.com",
		UserName:     "專家-Adora",
		IsExpert:     1,
		IsFrontline:  2,
		Area:         []int{1},
		AreaName:     []string{"場域A"},
		Pic:          picExpertA,
	}
	//專家帳號 場域B
	accountExpertB := Account{
		UserID:       "expertB@leapsyworld.com",
		UserPassword: "expertB@leapsyworld.com",
		UserName:     "專家-Belle",
		IsExpert:     1,
		IsFrontline:  2,
		Area:         []int{2},
		AreaName:     []string{"場域B"},
		Pic:          picExpertB,
	}

	//一線人員帳號 一般帳號
	accountFrontLine := Account{
		UserID:       "frontLine@leapsyworld.com",
		UserPassword: "frontLine@leapsyworld.com",
		UserName:     "一線人員帳號",
		IsExpert:     2,
		IsFrontline:  1,
		Area:         []int{},
		AreaName:     []string{},
		Pic:          picFrontline,
	}

	//一線人員帳號 匿名帳號
	defaultAccount := Account{
		UserID:       "default",
		UserPassword: "default",
		UserName:     "預設帳號",
		IsExpert:     2,
		IsFrontline:  1,
		Area:         []int{},
		AreaName:     []string{},
		Pic:          picDefault,
	}

	allAccountList = append(allAccountList, &accountExpertA)
	allAccountList = append(allAccountList, &accountExpertB)
	allAccountList = append(allAccountList, &accountFrontLine)
	allAccountList = append(allAccountList, &defaultAccount)

}

// 匯入所有裝置到AllDeviceList中
func importAllDevicesList() {

	// 待補:真的匯入資料庫所有裝置清單

	// 新增假資料：眼鏡假資料-場域A 眼鏡Model
	modelGlassesA := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   1,        //眼鏡
		Area:         []int{1}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     []string{"場域A"},
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 離線
		DeviceStatus: 0,            // 未設定
		MicStatus:    0,            // 未設定
		CameraStatus: 0,            // 未設定
		RoomID:       0,            // 無房間
	}

	// 新增假資料：場域B 眼鏡Model
	modelGlassesB := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   1,        //眼鏡
		Area:         []int{2}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     []string{"場域B"},
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 離線
		DeviceStatus: 0,            // 未設定
		MicStatus:    0,            // 未設定
		CameraStatus: 0,            // 未設定
		RoomID:       0,            // 無房間
	}

	// 新增假資料：場域A 平版Model
	modelTabA := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   2,        // 平版
		Area:         []int{1}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     []string{"場域A"},
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 離線
		DeviceStatus: 0,            // 未設定
		MicStatus:    0,            // 未設定
		CameraStatus: 0,            // 未設定
		RoomID:       0,            // 無房間
	}

	// 新增假資料：場域B 平版Model
	modelTabB := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   2,        // 平版
		Area:         []int{2}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     []string{"場域B"},
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 離線
		DeviceStatus: 0,            // 未設定
		MicStatus:    0,            // 未設定
		CameraStatus: 0,            // 未設定
		RoomID:       0,            // 無房間
	}

	// 新增假資料：場域A 眼鏡
	var glassesA [5]*Device
	for i, e := range glassesA {
		device := modelGlassesA
		e = &device
		e.DeviceID = "00" + strconv.Itoa(i+1)
		e.DeviceBrand = "00" + strconv.Itoa(i+1)
		glassesA[i] = e
	}
	fmt.Printf("假資料眼鏡A=%+v\n", glassesA)

	// 場域B 眼鏡
	var glassesB [5]*Device
	for i, e := range glassesB {
		device := modelGlassesB
		e = &device
		e.DeviceID = "00" + strconv.Itoa(i+6)
		e.DeviceBrand = "00" + strconv.Itoa(i+6)
		glassesB[i] = e
	}
	fmt.Printf("假資料眼鏡B=%+v\n", glassesB)

	// 場域A 平版
	var tabsA [1]*Device
	for i, e := range tabsA {
		device := modelTabA
		e = &device
		e.DeviceID = "00" + strconv.Itoa(i+11)
		e.DeviceBrand = "00" + strconv.Itoa(i+11)
		tabsA[i] = e
	}
	fmt.Printf("假資料平版A=%+v\n", tabsA)

	// 場域B 平版
	var tabsB [1]*Device
	for i, e := range tabsB {
		device := modelTabB
		e = &device
		e.DeviceID = "00" + strconv.Itoa(i+12)
		e.DeviceBrand = "00" + strconv.Itoa(i+12)
		tabsB[i] = e
	}
	fmt.Printf("假資料平版B=%+v\n", tabsB)

	// 加入眼鏡A
	for _, e := range glassesA {
		allDeviceList = append(allDeviceList, e)
	}
	// 加入眼鏡B
	for _, e := range glassesB {
		allDeviceList = append(allDeviceList, e)
	}
	// 加入平版A
	for _, e := range tabsA {
		allDeviceList = append(allDeviceList, e)
	}
	// 加入平版B
	for _, e := range tabsB {
		allDeviceList = append(allDeviceList, e)
	}

	// fmt.Printf("\n取得所有裝置%+v\n", AllDeviceList)
	for _, e := range allDeviceList {
		fmt.Printf("資料庫所有裝置清單:%+v\n", e)
	}

}

// 匯入所有場域對應名稱
func importAllAreasNameToMap() {
	areaNumberNameMap[1] = "場域A"
	areaNumberNameMap[2] = "場域B"
}

// 取得某些場域的AllDeviceList
func getAllDevicesListByAreas(area []int) []*Device {

	result := []*Device{}

	for _, e := range allDeviceList {
		intersection := intersect.Hash(e.Area, area) //取交集array
		// 若場域有交集則加入
		if len(intersection) > 0 {
			result = append(result, e)
		}
	}
	return result
}

// 登入邏輯重寫
func processLoginWithDuplicate(clientPointer *client, command Command, device *Device, account *Account) {

	// 建立帳號
	// account := Account{
	// 	UserID:         command.UserID,
	// 	UserPassword:   command.UserPassword,
	// 	IDPWIsRequired: command.IDPWIsRequired,
	// }

	// 建立Info
	info := Info{
		Account: account,
		Device:  device,
	}

	// 登入步驟:
	// 四種況狀判斷：相同連線、相異連線、不同裝置、相同裝置 重複登入之處理
	if _, ok := clientInfoMap[clientPointer]; ok {
		//相同連線

		if clientInfoMap[clientPointer].Device.DeviceID == command.DeviceID && clientInfoMap[clientPointer].Device.DeviceBrand == command.DeviceBrand {
			// 裝置相同：同裝置重複登入

			// 設定info
			clientInfoMap[clientPointer] = &info

			// 狀態為上線
			clientInfoMap[clientPointer].Device.OnlineStatus = 1

			// 裝置變閒置
			clientInfoMap[clientPointer].Device.DeviceStatus = 1

		} else {
			//裝置不同（現實中不會出現，只有測試才會出現）

			//舊的裝置＝離線
			clientInfoMap[clientPointer].Device.OnlineStatus = 2

			//設定新的info
			clientInfoMap[clientPointer] = &info

			//新的裝置＝上線
			clientInfoMap[clientPointer].Device.OnlineStatus = 1

			// 裝置變閒置
			clientInfoMap[clientPointer].Device.DeviceStatus = 1

			//不需要斷線

		}

	} else {
		// 不同連線

		// 去找Map中有沒有已經存在相同裝置
		isExist, oldClient := isDeviceExistInClientInfoMap(device)

		if isExist {
			// 裝置相同：（現實中，只有實體裝置重複ID才會實現）

			// 舊的連線：斷線＋從MAP移除＋告知舊的連線即將斷線
			processDisconnect(oldClient)

			// 新的連線，加入到Map，並且對應到新的裝置與帳號
			clientInfoMap[clientPointer] = &info
			fmt.Printf("找到重複的連線，從Map中刪除，將此Socket斷線。\n")

			// 狀態為上線
			clientInfoMap[clientPointer].Device.OnlineStatus = 1

			// 裝置變閒置
			clientInfoMap[clientPointer].Device.DeviceStatus = 1

		} else {
			//裝置不同：正常新增一的新裝置
			clientInfoMap[clientPointer] = &info

			// 裝置狀態＝線上
			clientInfoMap[clientPointer].Device.OnlineStatus = 1

			// 裝置變閒置
			clientInfoMap[clientPointer].Device.DeviceStatus = 1
		}

	}

}

// func getAccount(userID string, userPassword string) *Account {

// 	var result *Account

// 	for _, e := range allAccountList {
// 		if userID == e.UserID && userPassword == userPassword {
// 			result = e
// 			break
// 		}
// 	}
// 	return result
// }

// 斷線處理並通知
func processDisconnect(clientPointer *client) {

	// 告知舊的連線，即將斷線
	// 暫存即將斷線的資料
	// (讓logger可以進行平行處理，怕尚未執行到，就先刪掉了連線與裝置，就無法印出了)

	// 通知舊連線:有裝置重複登入，已斷線
	if jsonBytes, err := json.Marshal(HelpResponse{Command: CommandNumberOfLogout, CommandType: 2, ResultCode: 1, Results: `已斷線，有其他相同裝置ID登入伺服器。`, TransactionID: ""}); err == nil {
		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
		fmt.Println(`通知舊連線:有裝置重複登入，切斷連線`)
		logger.Infof(`通知舊連線:有裝置重複登入，切斷連線`)
	} else {
		fmt.Println(`json出錯`)
		logger.Errorf(`json出錯`)
	}

	// 舊的連線，從Map移除
	delete(clientInfoMap, clientPointer) // 此連線從Map刪除

	// 舊的連線，進行斷線
	disconnectHub(clientPointer) // 此連線斷線
}

// 去ClientInfoMap找，是否已經有相同裝置存在
func isDeviceExistInClientInfoMap(device *Device) (bool, *client) {

	for c, e := range clientInfoMap {
		// 若找到相同裝置，回傳連線c
		if e.Device.DeviceID == device.DeviceID && e.Device.DeviceBrand == device.DeviceBrand {
			return true, c
		}
	}
	return false, nil
}

// 根據裝置找到連線，關閉連線(排除自己這條連線)
func findClientByDeviceAndCloseSocket(device *Device, excluder *client) {

	for client, e := range clientInfoMap {
		// 若找到相同裝置，關閉此client連線
		if e.Device.DeviceID == device.DeviceID && e.Device.DeviceBrand == device.DeviceBrand && client != excluder {
			delete(clientInfoMap, client) // 此連線從Map刪除
			disconnectHub(client)         // 此連線斷線
			fmt.Printf("找到重複的連線，從Map中刪除，將此Socket斷線。\n")
		}
	}

}

// 進行帳號登入+裝置登入(維護:onlineList+處理重複登入)
// func processLogin(whatKindCommandString string, clientPointer *client, command Command, device *Device) bool {

// 	fmt.Printf("確認點GG")
// 	success := true            // 登入結果
// 	CopyDeviceNotFound := true // 沒有發現重複裝置

// 	// 看有沒有跟舊裝置重複

// 	// 2021.05.13: 改邏輯:此處邏輯錯了?

// 	for oldClientPointer, oldInfo := range clientInfoMap {

// 		if oldInfo.Device.DeviceID == device.DeviceID && oldInfo.Device.DeviceBrand == device.DeviceBrand {

// 			// 若發現重複
// 			CopyDeviceNotFound = false // 發現了重複裝置

// 			// 若是相同連線(不換client+不斷線)
// 			if clientPointer == oldClientPointer {

// 				allDevices := getAllDeviceByList() // 取得裝置清單-實體
// 				go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "發現裝置重複，且是相同連線", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "發現裝置重複，且是相同連線", command, clientPointer, clientInfoMap, allDevices, roomID)

// 				// 暫存舊device
// 				oldDevicePointer := clientInfoMap[oldClientPointer].Device

// 				// 更新MAP前
// 				allDevices = getAllDeviceByList() // 取得裝置清單-實體
// 				go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "更新clientInfoMAP前", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "更新clientInfoMAP前", command, clientPointer, clientInfoMap, allDevices, roomID)

// 				// go fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, "發現裝置重複，且是相同連線", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
// 				// go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, "發現裝置重複，且是相同連線", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
// 				// fmt.Printf("_______發現裝置重複，是相同連線\n")
// 				// logger.Infof("_______發現裝置重複，是相同連線\n")

// 				// logger:發現裝置重複，且是相同連線
// 				// go fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, "更新clientInfoMAP前", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
// 				// go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, "更新clientInfoMAP前", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

// 				// fmt.Printf("_______更新前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 				// logger.Infof("_______更新前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 				//fmt.Printf("_______更新Map前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 				//logger.Infof("_______更新Map前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 				// 重設帳號
// 				account := Account{
// 					UserID:         command.UserID,
// 					UserPassword:   command.UserPassword,
// 					IDPWIsRequired: command.IDPWIsRequired,
// 				}

// 				// oldInfo.Account.UserID = command.UserID
// 				// oldInfo.Account.UserPassword = command.UserPassword
// 				// oldInfo.Account.IDPWIsRequired = command.IDPWIsRequired
// 				oldInfo.Account = &account
// 				oldInfo.Device = device
// 				clientInfoMap[oldClientPointer] = oldInfo

// 				//fmt.Println("_______更新Map後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 				//logger.Infof("_______更新Map後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 				// 更新List
// 				//fmt.Println("_______更新List前,clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList)
// 				updateDeviceListByOldAndNewDevicePointers(oldDevicePointer, oldInfo.Device)
// 				//fmt.Println("_______更新List後,clientInfoMap=", clientInfoMap, ",onlineDeviceList", getLoginBasicInfoString(clientPointer))

// 				// 更新MAP
// 				allDevices = getAllDeviceByList() // 取得裝置清單-實體
// 				go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "更新clientInfoMAP後", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "更新clientInfoMAP後", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				// go fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, "更新clientInfoMAP後", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
// 				// go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, "更新clientInfoMAP後", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

// 				// fmt.Printf("_______更新後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 				// logger.Infof("_______更新後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 			} else {

// 				// 2021.05.13: 改邏輯:此處邏輯錯了?

// 				// 若是不同連線(換client+斷舊連線)

// 				// fmt.Printf("_______發現裝置重複，是不同連線\n")

// 				// logger:
// 				allDevices := getAllDeviceByList() // 取得裝置清單-實體
// 				fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "發現裝置重複，是不同連線", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "發現裝置重複，是不同連線", command, clientPointer, clientInfoMap, allDevices, roomID)

// 				// 暫存舊device
// 				infoOld := clientInfoMap[oldClientPointer]
// 				//accountOld := infoOld.Account
// 				deviceOld := infoOld.Device
// 				//oldDevicePointer := clientInfoMap[oldClientPointer].Device

// 				// 更新MAP前
// 				allDevices = getAllDeviceByList() // 取得裝置清單-實體
// 				fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "更新clientInfoMAP前", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "更新clientInfoMAP前", command, clientPointer, clientInfoMap, allDevices, roomID)

// 				// 重設帳號
// 				accountNew := Account{
// 					UserID:         command.UserID,
// 					UserPassword:   command.UserPassword,
// 					IDPWIsRequired: command.IDPWIsRequired,
// 				}

// 				// 更新Map Client連線(並斷掉舊連線)
// 				updateClientInfoMapAndDisconnectOldClient(oldClientPointer, clientPointer, &accountNew, device)

// 				// 更新List
// 				//updateDeviceListByOldAndNewDevicePointers(deviceOld, clientInfoMap[clientPointer].Device) //換成新的
// 				updateDeviceListByOldAndNewDevicePointers(deviceOld, device) //換成新的

// 				// 更新MAP
// 				allDevices = getAllDeviceByList() // 取得裝置清單-實體
// 				fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "更新clientInfoMAP後", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "更新clientInfoMAP後", command, clientPointer, clientInfoMap, allDevices, roomID)
// 				// go fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, "更新clientInfoMAP後", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
// 				// go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, "更新clientInfoMAP後", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

// 				// fmt.Printf("_______更新後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 				// logger.Infof("_______更新後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 			}

// 			//不進行廣播:因為其他人只需要知道這個裝置有沒有在線上。取代前後都是在線上，因此不廣播

// 			return success
// 		}

// 	}

// 	// 若沒發現重複裝置(此狀況現實狀況不會有，唯有測試時才會出現的狀況)
// 	if CopyDeviceNotFound {

// 		// 相同連線卻有不同裝置: 現實狀況不會發生，目前不支援此狀況
// 		if _, ok := clientInfoMap[clientPointer]; ok {

// 			// Response：失敗
// 			jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "失敗:相同連線卻有不同裝置:目前不支援此狀況(實際狀況不會發生)", command.TransactionID))
// 			clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

// 			// logger:失敗
// 			allDevices := getAllDeviceByList() // 取得裝置清單-實體
// 			go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "失敗:相同連線卻有不同裝置:目前不支援此狀況(實際狀況不會發生)", command, clientPointer, clientInfoMap, allDevices, roomID)
// 			go logger.Warnf(baseLoggerWhenLoginString, whatKindCommandString, "失敗:相同連線卻有不同裝置:目前不支援此狀況(實際狀況不會發生)", command, clientPointer, clientInfoMap, allDevices, roomID)

// 			// // logger:發現裝置重複，且是相同連線
// 			// allDevices := getAllDeviceByList() // 取得裝置清單-實體
// 			// go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "發現裝置不同，但是相同連線", command, clientPointer, clientInfoMap, allDevices, roomID)
// 			// go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "發現裝置不同，但是相同連線", command, clientPointer, clientInfoMap, allDevices, roomID)
// 			// // go fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, "發現裝置重複，且是相同連線", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
// 			// // go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, "發現裝置重複，且是相同連線", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

// 			// // 暫存舊的Info
// 			// oldDevicePointer := clientInfoMap[clientPointer].Device
// 			// //oldAccountPointer := clientInfoMap[clientPointer].Account

// 			// // 更新MAP前
// 			// allDevices = getAllDeviceByList() // 取得裝置清單-實體
// 			// go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "更新clientInfoMAP前", command, clientPointer, clientInfoMap, allDevices, roomID)
// 			// go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "更新clientInfoMAP前", command, clientPointer, clientInfoMap, allDevices, roomID)
// 			// // go fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, "更新clientInfoMAP前", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
// 			// // go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, "更新clientInfoMAP前", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

// 			// // fmt.Printf("_______更新前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 			// // logger.Infof("_______更新前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s\n", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 			// //fmt.Printf("_______更新Map前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 			// //logger.Infof("_______更新Map前,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 			// // 建立帳號
// 			// account := Account{
// 			// 	UserID:         command.UserID,
// 			// 	UserPassword:   command.UserPassword,
// 			// 	IDPWIsRequired: command.IDPWIsRequired,
// 			// }
// 			// device := getDevice(command.DeviceID, command.DeviceBrand) //取device
// 			// info := clientInfoMap[clientPointer]                       //取出
// 			// info.Account = &account                                    //更新值
// 			// info.Device = device                                       //更新值
// 			// clientInfoMap[clientPointer] = info                        //存回
// 			// newDevicePointer := clientInfoMap[clientPointer].Device
// 			// // oldInfo.Account.UserID = command.UserID
// 			// // oldInfo.Account.UserPassword = command.UserPassword
// 			// // oldInfo.Account.IDPWIsRequired = command.IDPWIsRequired
// 			// // oldInfo.Account = &account
// 			// // oldInfo.Device = device

// 			// //fmt.Println("_______更新Map後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))
// 			// //logger.Infof("_______更新Map後,clientInfoMap= %s, onlineDeviceList= %s, basicInfo= %s", clientInfoMap, onlineDeviceList, getLoginBasicInfoString(clientPointer))

// 			// // 更新List
// 			// //fmt.Println("_______更新List前,clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList)
// 			// updateDeviceListByOldAndNewDevicePointers(oldDevicePointer, newDevicePointer)
// 			// //fmt.Println("_______更新List後,clientInfoMap=", clientInfoMap, ",onlineDeviceList", getLoginBasicInfoString(clientPointer))

// 			// // 更新MAP
// 			// allDevices = getAllDeviceByList() // 取得裝置清單-實體
// 			// go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "更新clientInfoMAP後", command, clientPointer, clientInfoMap, allDevices, roomID)
// 			// go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "更新clientInfoMAP後", command, clientPointer, clientInfoMap, allDevices, roomID)
// 			success = false
// 			return success

// 		} else {
// 			// 若為全新的連線:正常增加

// 			// 新增帳號
// 			accountNew := Account{
// 				UserID:         command.UserID,
// 				UserPassword:   command.UserPassword,
// 				IDPWIsRequired: command.IDPWIsRequired,
// 			}

// 			// 無重複裝置: 加入新的連線到<Map>中
// 			clientInfoMap[clientPointer] = Info{Account: &accountNew, Device: device}
// 			fmt.Println("加入新連線後到Map後 clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList)

// 			// 新增裝置: 到<清單>中
// 			onlineDeviceList = append(onlineDeviceList, device)
// 			fmt.Println("新增裝置到清單後 clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList)
// 			return success
// 		}
// 	}
// 	return false
// }

// 待補:之後再拿掉 main直接呼叫 getAccount(即可)
// func checkLoginPassword(id string, pw string) *Account {

// 	// 找到帳號相關資料
// 	account := getAccount(id, pw)
// 	return account

// 	// if id == pw {
// 	// 	return true
// 	// } else {
// 	// 	return false
// 	// }
// }

func checkPassword(id string, pw string) (bool, *Account) {
	for _, e := range allAccountList {
		if id == e.UserID {
			if pw == e.UserPassword {
				return true, e
			} else {
				return false, nil
			}
		}
	}
	return false, nil
}

// 判斷某連線是否已經做完command:1指令，並加入到Map中(判定：透過 clientInfoMap[client] 看Info是否有值)
func checkLogedIn(client *client, command Command, whatKindCommandString string) bool {

	logedIn := false

	if _, ok := clientInfoMap[client]; ok {
		logedIn = true
	}

	if !logedIn {

		// 尚未登入
		// 失敗:Response
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, `連線尚未登入`, command.TransactionID))
		client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

		// logger
		details := `執行失敗，連線尚未登入，指令結束`
		myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(client) //所有值複製一份做logger
		processLoggerInfofBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		// allDevices := getAllDeviceByList() // 取得裝置清單-實體
		// go fmt.Printf(baseLoggerNotLoggedInWarnString+"\n", whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)
		// go logger.Warnf(baseLoggerNotLoggedInWarnString, whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)

		return logedIn

	} else {
		// 已登入
		return logedIn
	}

}

// 判斷某連線裝置是否為閒置
func checkDeviceStatusIsIdle(client *client, command Command, whatKindCommandString string) bool {

	// 若連線存在
	if e, ok := clientInfoMap[client]; ok {

		// 若為閒置
		if e.Device.DeviceStatus == 1 {
			return true
		} else {
			// 非閒置狀態
			// 失敗:Response
			jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, `裝置並非閒置狀態`, command.TransactionID))
			client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

			// logger
			details := `執行失敗，裝置並非閒置狀態，指令結束`
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(client) //所有值複製一份做logger
			processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			// allDevices := getAllDeviceByList() // 取得裝置清單-實體
			// go fmt.Printf(baseLoggerNotLoggedInWarnString+"\n", whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)
			// go logger.Warnf(baseLoggerNotLoggedInWarnString, whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)

			return false
		}
	} else {
		// 此連線不存在
		// 失敗:Response
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, `此連線不存在`, command.TransactionID))
		client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

		// logger
		details := `執行失敗，此連線不存在，指令結束`
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(client) //所有值複製一份做logger
		processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		// allDevices := getAllDeviceByList() // 取得裝置清單-實體
		// go fmt.Printf(baseLoggerNotLoggedInWarnString+"\n", whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)
		// go logger.Warnf(baseLoggerNotLoggedInWarnString, whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)

		return false
	}
}

func checkLogedInByClient(client *client) bool {

	logedIn := false

	if _, ok := clientInfoMap[client]; ok {
		logedIn = true
	}

	return logedIn

}

// 從清單移除某裝置
func removeDeviceFromList(slice []*Device, s int) []*Device {
	return append(slice[:s], slice[s+1:]...) //回傳移除後的array
}

// 更新 onlineDeviceList
func updateDeviceListByOldAndNewDevicePointers(deviceOld *Device, deviceNew *Device) {

	onlineDeviceList = removeDeviceFromListByDevice(onlineDeviceList, deviceOld) //移除舊的
	onlineDeviceList = append(onlineDeviceList, deviceNew)                       //增加新的

}

// 更新 ClientInfoMap 連線 + 斷掉舊的Client連線
//func updateClientInfoMapAndDisconnectOldClient(oldClientPointer *client, newClientPointer *client, userID string, userPassword string, device *Device) {
func updateClientInfoMapAndDisconnectOldClient(oldClientPointer *client, newClientPointer *client, accountNew *Account, device *Device) {
	// 通知舊連線:有裝置重複登入，已斷線
	if jsonBytes, err := json.Marshal(HelpResponse{Command: CommandNumberOfLogout, CommandType: 2, ResultCode: 1, Results: `已斷線，有相同裝置登入伺服器。`, TransactionID: ""}); err == nil {
		oldClientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
		fmt.Println(`通知舊連線:有裝置重複登入，切斷連線`)
		logger.Infof(`通知舊連線:有裝置重複登入，切斷連線`)
	} else {
		fmt.Println(`json出錯`)
		logger.Errorf(`json出錯`)
	}

	// 刪除Map舊連線
	//fmt.Println("_______刪除舊連線前,clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList, "device=", getLoginBasicInfoString(oldClientPointer))
	delete(clientInfoMap, oldClientPointer)
	// fmt.Println("_______刪除舊連線後,clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList)

	// 斷線舊連線
	disconnectHub(oldClientPointer) //斷線
	// fmt.Println("_______已斷線舊的連線,clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList)

	// 加入Map新連線
	// fmt.Println("_______加入新連線前,clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList)
	clientInfoMap[newClientPointer] = &Info{Account: accountNew, Device: device}
	// fmt.Println("_______加入新連線後,clientInfoMap=", clientInfoMap, ",onlineDeviceList", onlineDeviceList, ",device=", getLoginBasicInfoString(newClientPointer))
}

// 從Device List中 移除某 devicePointer
func removeDeviceFromListByDevice(list []*Device, device *Device) []*Device {

	// 尋找清單相同裝置
	for i, d := range onlineDeviceList {

		//if d != nil {
		if d.DeviceID == device.DeviceID && d.DeviceBrand == device.DeviceBrand {
			return append(list[:i], list[i+1:]...) //回傳移除後的array
		}
		//}
	}

	return []*Device{} //回傳空的
}

// 取得所有裝置清單 By clientInfoMap（For Logger）
func getOnlineDevicesByClientInfoMap() []Device {

	deviceArray := []Device{}

	for _, info := range clientInfoMap {
		device := info.Device
		deviceArray = append(deviceArray, *device)
	}

	return deviceArray
}

//取得所有裝置清單(For Logger)
func getAllDeviceByList() []Device {
	deviceArray := []Device{}
	for _, d := range allDeviceList {
		deviceArray = append(deviceArray, *d)
	}
	return deviceArray
}

//取得線上裝置清單(實體內容)(For Logger)
func getPhisicalDeviceArrayFromOnlineDeviceList() []Device {
	deviceArray := []Device{}
	for _, d := range onlineDeviceList {
		//for _, d := range deviceListByArea {

		// 若非空
		if d != nil {
			deviceArray = append(deviceArray, *d)
		}
	}
	return deviceArray
}

// 回傳沒有password的Account
// func getAccountWithoutPassword(account *Account) *AccountWithoutPassword {
// 	accountNew := &AccountWithoutPassword{
// 		UserID: account.UserID,
// 	}
// 	return accountNew
// }

// func getAccountNoPassword(id string, pw string) Account {
// 	account := getAccount(id, pw)
// 	accountNoPw := *account
// 	accountNoPw.UserPassword = ""
// 	return accountNoPw
// }

// 取得裝置
func getDevice(deviceID string, deviceBrand string) *Device {

	var device *Device
	// 若找到則返回
	for i, _ := range allDeviceList {
		if allDeviceList[i].DeviceID == deviceID {
			if allDeviceList[i].DeviceBrand == deviceBrand {

				device = allDeviceList[i]
				fmt.Println("找到裝置", device)
				fmt.Println("所有裝置清單", allDeviceList)
			}
		}
	}

	return device // 回傳
}

// 取得裝置:同區域＋同類型＋去掉某一裝置（自己）
func getDevicesByAreaAndDeviceTypeExeptOneDevice(area []int, deviceType int, device *Device) []*Device {

	result := []*Device{}

	// 若找到則返回
	for _, e := range allDeviceList {
		intersection := intersect.Hash(e.Area, area) //取交集array

		// 同區域＋同類型＋去掉某一裝置（自己）
		if len(intersection) > 0 && e.DeviceType == deviceType && e != device {

			result = append(result, e)

		}
	}

	fmt.Printf("找到指定場域＋指定類型＋排除自己的所有裝置:%+v \n", result)
	return result // 回傳
}

// 專家＋平版端 要取得所有Devic+Account = Info: 同區域＋某類型＋去掉某一裝置（自己）
func getDevicesWithInfoByAreaAndDeviceTypeExeptOneDevice(area []int, deviceType int, device *Device) []*Info {

	result := []*Info{}

	// 若找到則返回
	for _, e := range allDeviceList {
		intersection := intersect.Hash(e.Area, area) //取交集array

		// 同區域＋同類型＋去掉某一裝置（自己）
		if len(intersection) > 0 && e.DeviceType == deviceType && e != device {

			// 若裝置在線，則額外多加入Account資訊
			if 1 == e.OnlineStatus {
				info := getInfoByOnlineDevice(e)
				result = append(result, info)
			} else {
				//若裝置離線，包空的Account
				account := &Account{}
				info := &Info{Account: account, Device: e}
				result = append(result, info)
			}

		}
	}

	fmt.Printf("找到指定場域＋指定裝置類型＋排除自己的所有裝置:%+v \n", result)
	return result // 回傳
}

// 取得在線裝置所對應的Account
func getInfoByOnlineDevice(device *Device) *Info {

	for _, e := range clientInfoMap {
		if device == e.Device {
			return e
		}
	}
	//找不到就回傳空的
	return &Info{}
}

// 取得裝置
// func getDevice(deviceID string, deviceBrand string) *Device {

// 	var device *Device
// 	// 若找到則返回
// 	for i, _ := range onlineDeviceList {
// 		if onlineDeviceList[i].DeviceID == deviceID {
// 			if onlineDeviceList[i].DeviceBrand == deviceBrand {

// 				device = onlineDeviceList[i]
// 				fmt.Println("找到裝置", device)
// 				fmt.Println("裝置清單", onlineDeviceList)
// 			}
// 		}
// 	}

// 	return device // 回傳
// }

// 排除某連線進行廣播 (excluder 被排除的client)
func broadcastExceptOne(excluder *client, websocketData websocketData) {

	//Response to all
	for client, _ := range clientInfoMap {

		// 僅排除一個連線
		if client != excluder {

			client.outputChannel <- websocketData //Socket Response

		}
	}
}

// 針對指定群組進行廣播，排除某連線(自己)
func broadcastByGroup(group []*client, websocketData websocketData, excluder *client) {

	for i, _ := range group {

		//排除自己
		if group[i] != excluder {
			group[i].outputChannel <- websocketData //Socket Respone
		}
	}
}

// 針對某場域(Area)進行廣播，排除某連線(自己)
func broadcastByArea(area []int, websocketData websocketData, excluder *client) {

	for client, _ := range clientInfoMap {

		// 找相同的場域
		intersection := intersect.Hash(clientInfoMap[client].Device.Area, area) //取交集array

		// 若有找到
		if len(intersection) > 0 {

			//if clientInfoMap[client].Device.Area == area {
			if client != excluder { //排除自己

				client.outputChannel <- websocketData //Socket Response
			}
		}
	}
}

// 針對某房間(RoomID)進行廣播，排除某連線(自己)
func broadcastByRoomID(roomID int, websocketData websocketData, excluder *client) {

	for client, _ := range clientInfoMap {

		// 找到相同房間的連線
		if clientInfoMap[client].Device.RoomID == roomID {
			if client != excluder { //排除自己
				client.outputChannel <- websocketData //Socket Response
			}
		}
	}
}

// 包裝成 array / slice
func getArray(device *Device) []Device {
	var array = []Device{}
	array = append(array, *device)
	return array
}

func getArrayPointer(device *Device) []*Device {
	var array = []*Device{}
	array = append(array, device)
	return array
}

// 檢查欄位是否齊全
func checkCommandFields(command Command, fields []string) (bool, []string) {

	missFields := []string{} // 遺失的欄位
	ok := true               // 是否齊全

	for _, e := range fields {

		switch field := e; field {

		case "command":
			if command.Command == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "commandType":
			if command.CommandType == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "transactionID":
			if command.TransactionID == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "userID":
			if command.UserID == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "userPassword":
			if command.UserPassword == "" {
				missFields = append(missFields, field)
				ok = false
			}

		// case "functionNumber":
		// 	if command.FunctionNumber == 0 {
		// 		missFields = append(missFields, field)
		// 		ok = false
		// 	}

		// case "IDPWIsRequired":
		// 	if command.IDPWIsRequired == 0 {
		// 		missFields = append(missFields, field)
		// 		ok = false
		// 	}

		case "deviceID":
			if command.DeviceID == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceBrand":
			if command.DeviceBrand == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceType":
			if command.DeviceType == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "pic":
			if command.Pic == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "roomID":
			if command.RoomID == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		}

	}

	return ok, missFields
}

// 印出回傳裝置清單
func printDeviceList() string {

	s := ""

	for i, e := range onlineDeviceList {
		fmt.Println(` DeviceID:` + e.DeviceID + ",DeviceBrand:" + e.DeviceBrand)
		s += `[` + (string)(i) + `],DeviceID:` + e.DeviceID
		s += `[` + (string)(i) + `],DeviceBrand:` + e.DeviceBrand
	}

	return s
}

// 判斷欄位(個別指令專屬欄位)是否齊全
func checkFieldsCompleted(fields []string, client *client, command Command, whatKindCommandString string) bool {

	//fields := []string{"roomID"}
	ok, missFields := checkCommandFields(command, fields)

	if !ok {

		m := strings.Join(missFields, ",")

		// Response: 失敗
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 1, `以下欄位不齊全:`+m, command.TransactionID))
		client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

		// allDevices := getAllDeviceByList() // 取得裝置清單-實體
		// go fmt.Printf(baseLoggerNotCompletedFieldsWarnString+"\n", whatKindCommandString, m, command, clientInfoMap[client].Account.UserID, clientInfoMap[client].Device, client, clientInfoMap, allDevices, roomID)
		// go logger.Warnf(baseLoggerNotCompletedFieldsWarnString, whatKindCommandString, m, command, clientInfoMap[client].Account.UserID, clientInfoMap[client].Device, client, clientInfoMap, allDevices, roomID)
		return false

	} else {

		return true
	}

}

// 取得連線資訊物件
// 取得登入基本資訊字串
func getLoginBasicInfoString(c *client) string {

	s := " UserID:" + clientInfoMap[c].Account.UserID

	if clientInfoMap[c].Device != nil {
		s += ",DeviceID:" + clientInfoMap[c].Device.DeviceID
		s += ",DeviceBrand:" + clientInfoMap[c].Device.DeviceBrand
		s += ",DeviceName:" + clientInfoMap[c].Device.DeviceName
		//s += ",DeviceType:" + (string)(clientInfoMap[c].Device.DeviceType)
		s += ",DeviceType:" + strconv.Itoa(clientInfoMap[c].Device.DeviceType)                                      // int轉string
		s += ",Area:" + strings.Replace(strings.Trim(fmt.Sprint(clientInfoMap[c].Device.Area), "[]"), " ", ",", -1) //將int[]內容轉成string
		s += ",RoomID:" + strconv.Itoa(clientInfoMap[c].Device.RoomID)                                              // int轉string
		s += ",OnlineStatus:" + strconv.Itoa(clientInfoMap[c].Device.OnlineStatus)                                  // int轉string
		s += ",DeviceStatus:" + strconv.Itoa(clientInfoMap[c].Device.DeviceStatus)                                  // int轉string
		s += ",CameraStatus:" + strconv.Itoa(clientInfoMap[c].Device.CameraStatus)                                  // int轉string
		s += ",MicStatus:" + strconv.Itoa(clientInfoMap[c].Device.MicStatus)                                        // int轉string
	}

	return s
}

// 確認是否有此帳號
func checkAccountExist(id string) (bool, *Account) {

	for _, e := range allAccountList {
		if id == e.UserID {
			return true, e
		}
	}
	//找不到帳號
	return false, nil
}

// 確認是否有此帳號並且發送驗證碼信
func checkAccountExistAndSendPassword(id string) bool {
	for _, e := range allAccountList {

		if id == e.UserID {
			//寄送驗證信

			//待補
			sendPasswordMail(id)
			return true
		}
	}
	//找不到帳號
	return false
}

// 待補
func sendPasswordMail(id string) {
	//待補

	//建立隨機密string六碼

	//密碼記在新的userIDPasswordMap中: ID --> userPassword (密碼不需要紀錄在資料庫，因為是一次性密碼)

	//寄送包含密碼的EMAIL

	//額外:進行登出時要去把對應的password移除
}

// 處理區域裝置狀態改變的廣播(device的area)
func processBroadcastingDeviceChangeStatusInArea(whatKindCommandString string, command Command, clientPointer *client, device []*Device) {

	// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: device}); err == nil {
		//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

		var numArea []int
		if e, ok := clientInfoMap[clientPointer]; ok {
			numArea = e.Device.Area
		}
		// 廣播(場域、排除個人)
		broadcastByArea(numArea, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

		// logger
		details := `執行（同場域廣播）廣播成功，場域代碼：` + strconv.Itoa(clientInfoMap[clientPointer].Device.Area[0])
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
		processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		// logger:廣播
		// allDevices := getAllDeviceByList() // 取得裝置清單-實體
		// go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
		// go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

	} else {

		// logger
		details := `執行（同場域廣播）廣播失敗：後端json轉換出錯。`
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
		processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		// logger:json轉換出錯
		// allDevices := getAllDeviceByList() // 取得裝置清單-實體
		// go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
		// go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
	}
}

// 處理區域裝置狀態改變的廣播(指定area)
func processBroadcastingDeviceChangeStatusInSomeArea(whatKindCommandString string, command Command, clientPointer *client, device []*Device, area []int) {

	// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: device}); err == nil {
		//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

		// 廣播(場域、排除個人)
		broadcastByArea(area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

		// logger
		details := `執行（指定場域）廣播成功：場域代碼` + strconv.Itoa(area[0])
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
		processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		// logger:廣播
		// allDevices := getAllDeviceByList() // 取得裝置清單-實體
		// go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
		// go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

	} else {

		// logger
		details := `執行（指定場域）廣播失敗：後端json轉換出錯`
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
		processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		// logger:json轉換出錯
		// allDevices := getAllDeviceByList() // 取得裝置清單-實體
		// go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
		// go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
	}
}

func processBroadcastingDeviceChangeStatusInRoom(whatKindCommandString string, command Command, clientPointer *client, device []*Device) {

	// (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInRoom, CommandType: CommandTypeNumberOfBroadcast, Device: device}); err == nil {

		// 房間廣播:改變麥克風/攝影機狀態
		broadcastByRoomID(clientInfoMap[clientPointer].Device.RoomID, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

		// logger:房間廣播
		allDevices := getAllDeviceByList() // 取得裝置清單-實體
		go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
		go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

	} else {

		// json出錯
		allDevices := getAllDeviceByList() // 取得裝置清單-實體
		go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
		go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
	}
}

func getOnlineIdleExpertsCountInArea(area []int) int {

	counter := 0
	for _, e := range clientInfoMap {

		//找出同場域的專家帳號＋裝置閒置
		if 1 == e.Account.IsExpert {

			intersection := intersect.Hash(e.Account.Area, area) //取交集array

			// 若場域有交集
			if len(intersection) > 0 {

				//是否閒置
				if 1 == e.Device.DeviceStatus {
					counter++
				}
			}
		}
	}

	return counter
}

//取得同房間其他人連線
func getOtherClientsInTheSameRoom(clientPoint *client, roomID int) []*client {

	results := []*client{}

	for i, e := range clientInfoMap {
		if clientPoint != i {
			if roomID == e.Device.RoomID {
				results = append(results, i)
			}
		}
	}

	return results
}

//取得同房間其他人裝置
func getOtherDevicesInTheSameRoom(clientPoint *client, roomID int) []*Device {

	results := []*Device{}

	for i, e := range clientInfoMap {
		if clientPoint != i {
			if roomID == e.Device.RoomID {
				results = append(results, e.Device)
			}
		}
	}

	return results
}

// 登入後的logger
func getLoggerParrameters(clientPointer *client) (Account, Device, client, map[*client]*Info, []Device, int) {
	myAccount := *clientInfoMap[clientPointer].Account
	myDevice := *clientInfoMap[clientPointer].Device
	myClientPointer := *clientPointer
	myClientInfoMap := clientInfoMap
	myAllDevices := getAllDeviceByList() // 取得裝置清單-實體
	nowRooId := roomID
	return myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRooId
}

// 登入前的Logger
func getLoggerParrametersBeforeLogin(clientPointer *client) (client, map[*client]*Info, []Device, int) {
	myClientPointer := *clientPointer
	myClientInfoMap := clientInfoMap
	myAllDevices := getAllDeviceByList() // 取得裝置清單-實體
	nowRooId := roomID
	return myClientPointer, myClientInfoMap, myAllDevices, nowRooId
}

func processLoggerInfof(whatKindCommandString string, details string, command Command, myAccount Account, myDevice Device, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏
	go fmt.Printf(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)
	go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)

}

func processLoggerInfofBeforeLogin(whatKindCommandString string, details string, command Command, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	go fmt.Printf(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)
	go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)

}

func processLoggerInfofBeforeReadData(whatKindCommandString string, details string, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	go fmt.Printf(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)
	go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, details, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)

}

func processLoggerWarnf(whatKindCommandString string, details string, command Command, myAccount Account, myDevice Device, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏
	go fmt.Printf(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)
	go logger.Warnf(baseLoggerInfoCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)

}

func processLoggerWarnfBeforeLogin(whatKindCommandString string, details string, command Command, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	go fmt.Printf(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)
	go logger.Warnf(baseLoggerInfoCommonMessage, whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)

}

func processLoggerErrorfBeforeLogin(whatKindCommandString string, details string, command Command, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	go fmt.Printf(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)
	go logger.Warnf(baseLoggerInfoCommonMessage, whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)

}

func processLoggerErrorf(whatKindCommandString string, details string, command Command, myAccount Account, myDevice Device, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏
	go fmt.Printf(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)
	go logger.Errorf(baseLoggerInfoCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoomID)

}

// AES加密要素
var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

// AES加密
func aesEncoder(enString string) string {

	// 需要去加密的字串 轉[]byte
	plaintext := []byte(enString)

	// 如果傳入加密串的話，plaint 就是傳入的字串
	if len(os.Args) > 1 {
		plaintext = []byte(os.Args[1])
	}

	fmt.Println("測試plaintext＝", plaintext)
	fmt.Println("測試(string)plaintext＝", (string)(plaintext))

	// 複製KEY
	keyCopy := key_AES
	if len(os.Args) > 2 {
		keyCopy = os.Args[2]
	}

	fmt.Println("測試Key長度＝", len(keyCopy))

	// 建立加密演算法 aes
	c, err := aes.NewCipher([]byte(keyCopy))

	// 發生錯誤
	if err != nil {
		fmt.Printf("Error: NewCipher(%d bytes) = %s", len(keyCopy), err)
		os.Exit(-1)
	}

	//加密字串
	cfb := cipher.NewCFBEncrypter(c, commonIV)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	fmt.Printf("測試%s=>%x\n", plaintext, ciphertext)

	return (string)(ciphertext)
}

// AES解密
func aesDecoder(deString string, key string) string {

	// // 複製KEY
	// keyCopy := key_AES
	// if len(os.Args) > 2 {
	// 	keyCopy = os.Args[2]
	// }

	// fmt.Println("測試Key長度＝", len(keyCopy))

	// // 建立加密演算法 aes
	// c, err := aes.NewCipher([]byte(keyCopy))

	// // 發生錯誤
	// if err != nil {
	// 	fmt.Printf("Error: NewCipher(%d bytes) = %s", len(keyCopy), err)
	// 	os.Exit(-1)
	// }

	// // 解密字串
	// cfbdec := cipher.NewCFBDecrypter(c, commonIV)
	// result := make([]byte, len(plaintext))
	// cfbdec.XORKeyStream(result, ciphertext)
	// fmt.Printf("測試%x=>%s\n", ciphertext, result)

	return ""
}

// 取得帳號圖片
func getAccountPicString(fileName string) string {

	content, err := ioutil.ReadFile(fileName)

	//錯誤
	if err != nil {
		log.Fatal(err)
		return ""
	}

	// Convert []byte to string and print to screen
	text := string(content)
	fmt.Println(text)

	return text
}

func EncryptAES(key []byte, plaintext string) string {

	c, err := aes.NewCipher(key)
	CheckError(err)

	out := make([]byte, len(plaintext))

	c.Encrypt(out, []byte(plaintext))

	return hex.EncodeToString(out)
}

func DecryptAES(key []byte, ct string) string {
	ciphertext, _ := hex.DecodeString(ct)

	c, err := aes.NewCipher(key)
	CheckError(err)

	pt := make([]byte, len(ciphertext))
	c.Decrypt(pt, ciphertext)

	s := string(pt[:])
	fmt.Println("DECRYPTED:", s)

	return s
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func getEncodeUserID(mystring string) string {

	key := key_AES

	result := EncryptAES([]byte(key), mystring)

	// plaintext
	fmt.Println("等待加密字串", mystring)

	// ciphertext
	fmt.Println("已經加密字串", result)

	return result
}

func getDecodeUserID(mystring string) string {

	key := key_AES

	result := DecryptAES([]byte(key), mystring)

	fmt.Println("解密後字串", result)

	return result
}

// TokenInfo - 令牌資訊
type TokenInfo struct {
	UserID string
	//Device     string
}

// var (
// 	secretByteArray = getNewSecretByteArray()
// )

// // getSecretByteArray - 取得密鑰
// func getSecretByteArray() []byte {
// 	secretByteArrayRWMutex.RLock()
// 	gotSecretByteArray := secretByteArray
// 	secretByteArrayRWMutex.RUnlock()
// 	return gotSecretByteArray
// }

// // getNewSecretByteArray - 取得新密鑰
// /*
//  * @return []byte result 結果
//  */
// func getNewSecretByteArray() (result []byte) {

// 	result = []byte(base64.StdEncoding.EncodeToString([]byte(`新的key`)))

// 	return
// }

// // CreateToken - 產生令牌
// /*
//  * @params TokenInfo tokenInfo 令牌資訊
//  * @return *string returnTokenStringPointer 令牌字串指標
//  */
// func CreateToken(tokenInfoPointer *TokenInfo) (returnTokenStringPointer *string) {

// 	if nil != tokenInfoPointer {
// 		claim := jwt.MapClaims{
// 			`UserID`: tokenInfoPointer.UserID,
// 		}

// 		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
// 		tokenString, tokenSignedStringError := token.SignedString(getSecretByteArray())

// 		logings.SendLog(
// 			[]string{`簽署令牌字串 %s `},
// 			[]interface{}{tokenString},
// 			tokenSignedStringError,
// 			logrus.InfoLevel,
// 		)

// 		if nil == tokenSignedStringError {
// 			returnTokenStringPointer = &tokenString
// 		}

// 	}

// 	return
// }

// func getSecretFunction() jwt.Keyfunc {
// 	return func(token *jwt.Token) (interface{}, error) {
// 		return getSecretByteArray(), nil
// 	}
// }

// // ParseToken - 解析令牌
// /*
//  * @params string tokenString 令牌字串
//  * @return *TokenInfo tokenInfoPointer 令牌資訊指標
//  */
// func ParseToken(tokenString string) (tokenInfoPointer *TokenInfo) {

// 	if tokenString != `` {

// 		defaultFormatSlices := []string{`解析令牌 %s `}
// 		dafaultArgs := []interface{}{tokenString}

// 		token, jwtParseError := jwt.Parse(tokenString, getSecretFunction())

// 		logings.SendLog(
// 			[]string{`解析令牌 %s `},
// 			dafaultArgs,
// 			jwtParseError,
// 			logrus.InfoLevel,
// 		)

// 		if jwtParseError != nil {
// 			return
// 		}

// 		mapClaim, ok := token.Claims.(jwt.MapClaims)

// 		if !ok {

// 			logings.SendLog(
// 				defaultFormatSlices,
// 				dafaultArgs,
// 				errors.New(`將Claim轉成MapClaim錯誤`),
// 				logrus.InfoLevel,
// 			)

// 			return
// 		}

// 		// 驗證token，如果token被修改過則為false
// 		if !token.Valid {

// 			logings.SendLog(
// 				defaultFormatSlices,
// 				dafaultArgs,
// 				errors.New(`無效令牌錯誤`),
// 				logrus.InfoLevel,
// 			)

// 			return
// 		}

// 		employeeIDValue, employeeIDOK := mapClaim[`UserID`].(string)

// 		if employeeIDOK {
// 			tokenInfoPointer = &TokenInfo{
// 				UserID: employeeIDValue,
// 			}
// 		}

// 	}

// 	return
// }

// keepReading - 保持讀取
func (clientPointer *client) keepReading() {

	if nil != clientPointer { // 若指標不為空

		defer func() {
			disconnectHub(clientPointer) // 中斷客戶端與網路中心的連線
		}()

		connectionPointer := clientPointer.getConnectionPointer() // 連線指標

		if nil != connectionPointer { // 若指標不為空

			// connection := *connectionPointer // 客戶端的連線

			// 準備偵測連線逾時
			commandTimeChannel := make(chan time.Time, 1000) // 連線逾時計算之通道(時間，1個buffer)

			// 開始偵測連線逾時
			go func() {

				fmt.Println(`【伺服器:開始偵測連線逾時】`)
				logger.Infof(`【伺服器:開始偵測連線逾時】`)

				for {

					if checkLogedInByClient(clientPointer) && clientInfoMap[clientPointer].Device == nil {
						// 若連線已經登入，並且裝置已經被刪除，就認為是<登出>狀態，就不再偵測逾時。
						go fmt.Printf(baseLoggerInfoForTimeoutWithoutNilDevice+"\n", timeout, clientInfoMap[clientPointer].Account.UserID, clientPointer, onlineDeviceList, allDeviceList, roomID)
						go logger.Infof(baseLoggerInfoForTimeoutWithoutNilDevice, timeout, clientInfoMap[clientPointer].Account.UserID, clientPointer, onlineDeviceList, allDeviceList, roomID)
						break // 跳出
					}

					commandTime := <-commandTimeChannel                                  // 當有接收到指令，則會有值在此通道
					<-time.After(commandTime.Add(time.Second * timeout).Sub(time.Now())) // 若超過時間，則往下進行
					if 0 == len(commandTimeChannel) {                                    // 若通道裡面沒有值，表示沒有收到新指令過來，則斷線

						// 暫存即將斷線的資料
						// (讓logger可以進行平行處理，怕尚未執行到，就先刪掉了連線與裝置，就無法印出了)
						var tempClientPointer client
						var tempAccountPointer Account
						var tempDevicePointer Device
						//var tempClientDevicePointer Device

						if clientPointer != nil {
							//處理nil pointer問題
							tempClientPointer = *clientPointer

							if clientInfoMap[clientPointer].Account != nil {
								tempAccountPointer = *clientInfoMap[clientPointer].Account
								//tempClientAccountPointer = clientInfoMap[clientPointer].Account
							}

							if clientInfoMap[clientPointer].Device != nil {
								tempDevicePointer = *clientInfoMap[clientPointer].Device
							}
							//tempClientDevicePointer = *clientInfoMap[clientPointer].Device
						}

						tempClientInfoMap := clientInfoMap
						tempRoomID := roomID

						// logger:此裝置發生逾時
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						details := `此裝置發生連線逾時`
						fmt.Printf(baseLoggerInfoForTimeout+"\n", details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)
						logger.Warnf(baseLoggerWarnForTimeout, details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)

						// 設定裝置在線狀態=離線
						element := clientInfoMap[clientPointer]
						if element.Device != nil {
							element.Device.OnlineStatus = 2 // 離線
						}
						clientInfoMap[clientPointer] = element // 回存

						// 通知連線:
						// Response:即將斷線
						details = `將斷線，連線逾時timeout。`
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, CommandNumberOfLogout, CommandTypeNumberOfAPIResponse, 1, details, ""))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:即將斷線
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						fmt.Printf(baseLoggerWarnForTimeout+"\n", details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)
						logger.Warnf(baseLoggerWarnForTimeout, details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)

						// 準備包成array
						//device := []Device{}

						// 若裝置存在進行包裝array + 廣播
						if element.Device != nil {

							// 包成array
							device := getArray(clientInfoMap[clientPointer].Device)

							// 【廣播】狀態變更-離線（此處仍用工具Marshal轉換，因為有陣列格式array，若轉成string過於麻煩）
							if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: device}); err == nil {

								//【待解決？】可能有 *client指標錯誤？

								// 若Area存在
								tempArea := clientInfoMap[clientPointer].Device.Area
								if tempArea != nil && clientPointer != nil {

									// 區域廣播：狀態改變
									broadcastByArea(tempArea, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行廣播

									// logger:區域廣播
									allDevices := getAllDeviceByList() // 取得裝置清單-實體
									details := `(場域)廣播：此連線已逾時，此裝置狀態已變更為:離線`
									// 此處不可用平行go處理 若斷線
									fmt.Printf(baseLoggerWarnForTimeout+"\n", details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)
									logger.Warnf(baseLoggerWarnForTimeout, details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)

								} else {

									// logger:區域廣播
									allDevices := getAllDeviceByList() // 取得裝置清單-實體
									details := `(場域)廣播時發生錯誤，未廣播: area(場域) 或 clientPointer 值為空`
									fmt.Printf(baseLoggerErrorForTimeout+"Area:%s \n", details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID, tempArea)
									logger.Errorf(baseLoggerErrorForTimeout+"Area:%s", details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID, tempArea)
								}

							} else {

								// logger:json轉換出錯
								allDevices := getAllDeviceByList() // 取得裝置清單-實體
								details := `(場域)廣播時發生錯誤，未廣播：jason轉譯出錯`
								fmt.Printf(baseLoggerErrorForTimeout+"\n", details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)
								logger.Errorf(baseLoggerErrorForTimeout, details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)
								break // 跳出

							}

						}

						// 移除連線
						delete(clientInfoMap, clientPointer) //刪除
						disconnectHub(clientPointer)         //斷線

						// 從清單中移除裝置
						fmt.Printf("新測試點GH: onlineDeviceList=%+v, tempClientDevicePointer=%+v \n", onlineDeviceList, tempDevicePointer)
						onlineDeviceList = removeDeviceFromListByDevice(onlineDeviceList, &tempDevicePointer)

						// logger:區域廣播
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						details = `已斷線(刪除連線與從裝置清單中移除)`
						fmt.Printf(baseLoggerWarnForTimeout+"\n", details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)
						logger.Infof(baseLoggerWarnForTimeout, details, timeout, tempAccountPointer, tempDevicePointer, tempClientPointer, tempClientInfoMap, allDevices, tempRoomID)

					}

					//再看要不要客戶端主動登出時，就不再繼續計算逾時
					// fmt.Println("測逾時For底部2:ID=", testTempClientUserID, "。")

				}
			}()

			for { // 循環讀取連線傳來的資料

				whatKindCommandString := `不斷從客戶端讀資料`

				inputWebsocketData, isSuccess := getInputWebsocketDataFromConnection(connectionPointer) // 不斷從客戶端讀資料

				// fmt.Printf("從客戶端讀取資料,連線訊息=%+v"+"\n", clientPointer)
				// logger.Infof(`從客戶端讀取資料,連線訊息=%+v`, clientPointer)

				if !isSuccess { //若不成功 (判斷為Socket斷線)

					// fmt.Printf(`讀取資料不成功,連線資訊=%+v`+"\n", clientPointer)
					// logger.Infof(`讀取資料不成功,連線資訊=%+v`, clientPointer)

					details := `執行失敗，即將斷線`
					myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
					processLoggerInfofBeforeReadData(whatKindCommandString, details, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					disconnectHub(clientPointer) //斷線
					// fmt.Printf(`讀取資料不成功:斷線,連線資訊=%+v`+"\n", clientPointer)
					// logger.Infof(`讀取資料不成功:斷線,連線資訊=%+v`, clientPointer)

					return // 回傳:離開for迴圈(keep Reading)
				} else {

					details := `執行成功`
					myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
					processLoggerInfofBeforeReadData(whatKindCommandString, details, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				}

				wsOpCode := inputWebsocketData.wsOpCode
				dataBytes := inputWebsocketData.dataBytes

				if ws.OpText == wsOpCode {

					var command Command

					whatKindCommandString := `收到指令，初步解譯成Json格式`

					//解譯成Json
					err := json.Unmarshal(dataBytes, &command)

					//json格式錯誤
					if err == nil {
						// logger:成功
						// logger
						details := `執行成功`
						myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
						processLoggerInfofBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					} else {
						// logger:失敗:收到json格式錯誤
						details := `執行失敗：json格式錯誤，指令結束`
						myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
						processLoggerWarnfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					}

					// 檢查command欄位是否都給了
					fields := []string{"command", "commandType", "transactionID"}
					ok, missFields := checkCommandFields(command, fields)
					if !ok {

						whatKindCommandString := `收到指令，檢查欄位`

						m := strings.Join(missFields, ",")

						// Socket Response
						if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "失敗:欄位不齊全。" + m, TransactionID: command.TransactionID}); err == nil {

							// 失敗:欄位不完全
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

							// logger
							details := `執行失敗：欄位不完全，指令結束`
							myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
							processLoggerWarnfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// break
						} else {

							// 失敗:json轉換錯誤
							details := `執行失敗：欄位不完全，指令結束，但JSON轉換錯誤`
							myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
							processLoggerErrorfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// break
						}
						//return
					}

					// 判斷指令
					switch c := command.Command; c {

					case 1: // 登入(+廣播改變狀態)

						whatKindCommandString := `登入`

						// 檢查<帳號驗證功能>欄位是否齊全
						if !checkFieldsCompleted([]string{"userID", "userPassword", "deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
						processLoggerInfofBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						//等一下從這邊開始改 確認驗證密碼功能
						// 待補:拿ID+驗證碼去資料庫比對驗證碼，若正確則進行登入
						check, account := checkPassword(command.UserID, command.UserPassword)

						// 驗證成功:
						if check {

							// 取得裝置Pointer
							device := getDevice(command.DeviceID, command.DeviceBrand)

							// 資料庫找不到此裝置
							if device == nil {

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "資料庫找不到此裝置", command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details = `執行失敗，資料庫找不到此裝置，指令結束`
								myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
								processLoggerWarnfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								break // 跳出
							}

							// 進行裝置、帳號登入 (加入Map。包含處理裝置重複登入)
							processLoginWithDuplicate(clientPointer, command, device, account)

							// Response:成功
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行成功`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// 準備廣播:包成Array:放入 Response Devices
							deviceArray := getArrayPointer(device) // 包成array
							processBroadcastingDeviceChangeStatusInArea(whatKindCommandString, command, clientPointer, deviceArray)

							// logger
							details = `執行廣播，指令結束`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						} else {
							// logger:帳密錯誤

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "無此帳號或密碼錯誤", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行失敗，無此帳號或密碼錯誤，指令結束`
							myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
							processLoggerWarnfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出
						}

					case 2: // 取得同場域所有眼鏡裝置清單

						whatKindCommandString := `取得同場域所有眼鏡裝置清單`

						// 該有欄位外層已判斷

						// 是否已與Server建立連線
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						//glassesInAreasExceptMineDevice := getDevicesByAreaAndDeviceTypeExeptOneDevice(clientInfoMap[clientPointer].Device.Area, 1, clientInfoMap[clientPointer].Device) // 取得裝置清單-實體
						//infosInAreasExceptMineDevice := getInfosByAreaAndDeviceTypeExeptOneDevice(clientInfoMap[clientPointer].Device.Area, 1, clientInfoMap[clientPointer].Device) // 取得裝置清單-實體
						infosInAreasExceptMineDevice := getDevicesWithInfoByAreaAndDeviceTypeExeptOneDevice(clientInfoMap[clientPointer].Device.Area, 1, clientInfoMap[clientPointer].Device) //待改

						// Response:成功
						// 此處json不直接轉成string,因為有 device Array型態，轉string不好轉
						if jsonBytes, err := json.Marshal(DevicesResponse{Command: 2, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID, Info: infosInAreasExceptMineDevice}); err == nil {

							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Response

							// logger
							details = `執行成功`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						} else {

							// logger
							details = `執行失敗，後端json轉換出錯。`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出

						}

					case 3: // 取得空房號

						whatKindCommandString := `取得空房號`

						// 該有欄位外層已判斷

						// 是否已登入
						if !checkLogedIn(clientPointer, command, `whatKindCommandString`) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 增加房號
						roomID = roomID + 1

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+", RoomID:%d}", command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID, roomID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 4: // 求助

						whatKindCommandString := `求助`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompleted([]string{"pic", "roomID"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 檢核:房號未被取用過則失敗
						if command.RoomID > roomID {

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "房號未被取用過", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行失敗，房號未被取用過，指令結束`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出
						}

						// 設定Pic, RoomID
						info := clientInfoMap[clientPointer]
						info.Device.Pic = command.Pic       // Pic
						info.Device.RoomID = command.RoomID // RoomID
						info.Device.DeviceStatus = 2        // 設備狀態:求助中
						clientInfoMap[clientPointer] = info // 回存Map

						// 取得在線閒置專家數量

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備廣播:包成Array:放入 Response Devices
						//deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array
						deviceArray := getArrayPointer(clientInfoMap[clientPointer].Device) // 包成array
						processBroadcastingDeviceChangeStatusInArea(whatKindCommandString, command, clientPointer, deviceArray)

						// logger
						details = `執行廣播，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 5: // 回應求助

						whatKindCommandString := `回應求助`

						//TransactionID 外層已經檢查過

						// 是否已登入
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompleted([]string{"deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 設定求助者的狀態
						device := getDevice(command.DeviceID, command.DeviceBrand)

						// 檢查有此裝置
						if device == nil {

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, "求助者不存在", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger:失敗
							details := `執行失敗，求助者不存在`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出

						}

						// 設定:求助者設備狀態
						device.DeviceStatus = 3 // 通話中
						device.CameraStatus = 1 //預設開啟相機
						device.MicStatus = 1    //預設開啟麥克風

						// 設定:回應者設備狀態+房間(自己)
						info := clientInfoMap[clientPointer]
						info.Device.DeviceStatus = 3        // 通話中
						info.Device.CameraStatus = 1        //預設開啟相機
						info.Device.MicStatus = 1           //預設開啟麥克風
						info.Device.RoomID = device.RoomID  // 求助者roomID
						clientInfoMap[clientPointer] = info // 存回Map

						//2021.5.19明天從此處開始檢查

						// Response：成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArrayPointer(clientInfoMap[clientPointer].Device) // 包成Array:放入回應者device
						deviceArray = append(deviceArray, device)                           // 包成Array:放入求助者device

						processBroadcastingDeviceChangeStatusInArea(whatKindCommandString, command, clientPointer, deviceArray)

						// logger
						details = `執行廣播，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 6: // 變更	cam+mic 狀態

						whatKindCommandString := `變更攝影機+麥克風狀態`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompleted([]string{"cameraStatus", "micStatus"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 設定攝影機、麥克風
						element := clientInfoMap[clientPointer]            // 取出device
						element.Device.CameraStatus = command.CameraStatus // 攝影機
						element.Device.MicStatus = command.MicStatus       // 麥克風
						clientInfoMap[clientPointer] = element             // 回存Map

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArrayPointer(clientInfoMap[clientPointer].Device)

						processBroadcastingDeviceChangeStatusInRoom(whatKindCommandString, command, clientPointer, deviceArray)

						// logger
						details = `進行廣播，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 7: // 掛斷通話

						whatKindCommandString := `掛斷通話`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger:收到指令
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						info := clientInfoMap[clientPointer]
						thisRoomID := info.Device.RoomID

						// 其他同房間的連線
						//otherClients := getOtherClientsInTheSameRoom(clientPointer, thisRoomID)

						// 其他同房間裝置
						otherDevices := getOtherDevicesInTheSameRoom(clientPointer, thisRoomID)

						// 若自己是一線人員掛斷: 同房間都掛斷
						if 1 == info.Account.IsFrontline {

							// 自己 離開房間
							info.Device.DeviceStatus = 1        // 裝置閒置
							info.Device.CameraStatus = 0        // 關閉
							info.Device.MicStatus = 0           // 關閉
							info.Device.RoomID = 0              // 沒有房間
							info.Device.Pic = ""                //清空
							clientInfoMap[clientPointer] = info // 回存Map

							// 其他人 離開房間
							for _, d := range otherDevices {
								d.DeviceStatus = 1 // 裝置閒置
								d.CameraStatus = 0 // 關閉
								d.MicStatus = 0    // 關閉
								d.RoomID = 0       // 沒有房間
							}
							// for _, c := range otherClients {
							// 	e := clientInfoMap[c]
							// 	e.Device.DeviceStatus = 1        // 裝置閒置
							// 	e.Device.CameraStatus = 0        // 關閉
							// 	e.Device.MicStatus = 0           // 關閉
							// 	e.Device.RoomID = 0              // 沒有房間
							// 	clientInfoMap[clientPointer] = e // 回存Map
							// 	break
							// }

						} else if 1 == info.Account.IsExpert {
							//若是自己是專家掛斷: 一線人員 變求助中

							//自己 離開房間
							info.Device.DeviceStatus = 1        // 裝置閒置
							info.Device.CameraStatus = 0        // 關閉
							info.Device.MicStatus = 0           // 關閉
							info.Device.RoomID = 0              // 沒有房間
							clientInfoMap[clientPointer] = info // 回存Map

							//一線人員 求助中
							for _, d := range otherDevices {
								//一線人員  求助中
								d.DeviceStatus = 2 // 求助中
								d.CameraStatus = 0 // 關閉
								d.MicStatus = 0    // 關閉
								// 房間不變
							}
							// for _, c := range otherClients {
							// 	//一線人員  求助中
							// 	e := clientInfoMap[c]
							// 	e.Device.DeviceStatus = 2 // 求助中
							// 	e.Device.CameraStatus = 0 // 關閉
							// 	e.Device.MicStatus = 0    // 關閉
							// 	// 房間不變
							// 	clientInfoMap[clientPointer] = e // 回存Map
							// 	break
							// }
						}

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:執行成功
						details = `執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備廣播:包成Array:放入 Response Devices
						// 要放入自己＋其他人
						deviceArray := getArrayPointer(clientInfoMap[clientPointer].Device) // 包成array
						for _, e := range otherDevices {
							deviceArray = append(deviceArray, e)
						}

						processBroadcastingDeviceChangeStatusInArea(whatKindCommandString, command, clientPointer, deviceArray)

						// logger:進行廣播
						details = `進行廣播，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 8: // 登出

						whatKindCommandString := `登出`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						// 登出就不用再重新計算了
						//commandTimeChannel <- time.Now()

						// logger:收到指令
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 設定登出者
						info := clientInfoMap[clientPointer] // 取出device
						info.Device.OnlineStatus = 2         // 離線
						info.Device.DeviceStatus = 0         // 重設
						info.Device.CameraStatus = 0         // 重設
						info.Device.MicStatus = 0            // 重設
						info.Device.RoomID = 0               // 重設
						info.Device.Pic = ""                 //重設
						clientInfoMap[clientPointer] = info  // 回存Map

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArrayPointer(clientInfoMap[clientPointer].Device) // 包成array

						// 進行廣播
						processBroadcastingDeviceChangeStatusInArea(whatKindCommandString, command, clientPointer, deviceArray)

						// 暫存即將斷線的資料
						// logger
						details = `執行廣播，指令結束，連線已登出`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 移除連線
						delete(clientInfoMap, clientPointer) //刪除
						disconnectHub(clientPointer)         //斷線

					case 9: // 心跳包

						whatKindCommandString := `心跳包`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 成功:Response
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 13: // 取得自己帳號資訊

						whatKindCommandString := `取得自己帳號資訊`

						// 該有欄位外層已判斷

						// 是否已與Server建立連線
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 隱匿密碼
						accountNoPassword := *(clientInfoMap[clientPointer].Account)
						accountNoPassword.UserPassword = ""

						// Response:成功 (此處仍使用Marshal工具轉型，因考量有 物件{} 形態，轉成string較為複雜。)
						if jsonBytes, err := json.Marshal(MyAccountResponse{
							Command:       command.Command,
							CommandType:   CommandTypeNumberOfAPIResponse,
							ResultCode:    ResultCodeSuccess,
							Results:       ``,
							TransactionID: command.TransactionID,
							Account:       accountNoPassword}); err == nil {
							//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

							fmt.Printf("測試accountNoPassword=%+v", clientInfoMap[clientPointer].Account)
							// Response(場域、排除個人)
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行成功，指令結束`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						} else {

							// logger
							details = `執行失敗，後端json轉換出錯。`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						}
						// jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+`,"account":"%+v"}`, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID, accountNoPassword))
						// clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						// details = `執行成功，指令結束`
						// myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						// processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 14: // 取得自己裝置資訊

						whatKindCommandString := `取得自己裝置資訊`

						// 該有欄位外層已判斷

						// 是否已經加到clientInfoMap中 表示已登入
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						device := getDevice(clientInfoMap[clientPointer].Device.DeviceID, clientInfoMap[clientPointer].Device.DeviceBrand) // 取得裝置清單-實體                                                                                     // 自己的裝置

						// Response:成功 (此處仍使用Marshal工具轉型，因考量有 物件{}形態，轉成string較為複雜。)
						if jsonBytes, err := json.Marshal(MyDeviceResponse{Command: command.Command, CommandType: CommandTypeNumberOfAPIResponse, ResultCode: ResultCodeSuccess, Results: ``, TransactionID: command.TransactionID, Device: *device}); err == nil {
							//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

							// Response(場域、排除個人)
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行成功，指令結束`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						} else {

							// logger
							details = `執行失敗，後端json轉換出錯。`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						}

					case 15: // 判斷帳號是否存在 若存在則寄出驗證信

						whatKindCommandString := `判斷帳號是否存在，若存在寄出驗證信`

						// 該有欄位外層已判斷

						// 不用登入就可使用

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
						processLoggerInfofBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 是否有此帳號
						haveAccount := checkAccountExistAndSendPassword(command.UserID)

						if haveAccount {
							// Response:成功
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行成功，指令結束`
							myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
							processLoggerInfofBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						} else {
							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, "無此帳號", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行失敗，無此帳號，指令結束`
							myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
							processLoggerWarnfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						}

					case 16: // 取得現在同場域空閒專家人數

						whatKindCommandString := `取得現在同場域空閒專家人數`

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 取得線上同場域閒置專家數
						onlinExperts := getOnlineIdleExpertsCountInArea(clientInfoMap[clientPointer].Device.Area)

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+`,"onlineExpertsIdle":%d}`, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID, onlinExperts))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 17: // QRcode登入 (+廣播)

						whatKindCommandString := `QRcode登入`

						// 檢查<帳號驗證功能>欄位是否齊全
						if !checkFieldsCompleted([]string{"userID", "deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// 設定KEY

						// logger
						details := `收到指令`
						myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
						processLoggerInfofBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// QRcode登入不需要密碼，只要確認是否有此帳號

						// 加密
						// claim := jwt.MapClaims{
						// 	`EmployeeID`: command.UserID,
						// }
						// token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
						// tokenString, tokenSignedStringError := token.SignedString(getSecretByteArray())

						// if nil != tokenSignedStringError {
						// 	fmt.Println("加密失敗")
						// } else {
						// 	fmt.Println("加密成功：加密結果", tokenString)
						// }

						// // 解密
						// jwt.Parse(tokenString, getSecretFunction())

						// fmt.Println("解密成功：解密結果＝", tokenString)

						// 帳號加密
						// userIDEncode := getEncodeUserID("frontLine@leapsyworld.com")

						// fmt.Println("加密後code=", userIDEncode)

						// // 帳號解碼
						// userIDdecode := getDecodeUserID(userIDEncode)

						// // 是否有此帳號
						check, account := checkAccountExist(command.UserID)

						// 有此帳號
						if check {

							// 取得裝置Pointer
							device := getDevice(command.DeviceID, command.DeviceBrand)

							// 資料庫找不到此裝置
							if device == nil {

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "資料庫找不到此裝置", command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details = `執行失敗，資料庫找不到此裝置，指令結束`
								myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
								processLoggerWarnfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								break // 跳出
							}

							// 進行裝置、帳號登入 (加入Map。包含處理裝置重複登入)
							processLoginWithDuplicate(clientPointer, command, device, account)

							// Response:成功
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行成功`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// 準備廣播:包成Array:放入 Response Devices
							deviceArray := getArrayPointer(device) // 包成array
							processBroadcastingDeviceChangeStatusInArea(whatKindCommandString, command, clientPointer, deviceArray)

							// logger
							details = `執行廣播，指令結束`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
							processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						} else {
							// logger:帳密錯誤

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "無此帳號或密碼錯誤", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							details = `執行失敗，無此帳號或密碼錯誤，指令結束`
							myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrametersBeforeLogin(clientPointer) //所有值複製一份做logger
							processLoggerWarnfBeforeLogin(whatKindCommandString, details, command, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出
						}

					case 18: // 切換場域

						whatKindCommandString := `切換場域`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 閒置中才可以切換場域（通話中、求助中無法切換場域）
						if !checkDeviceStatusIsIdle(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompleted([]string{"area"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 暫存
						var oldArea []int          //舊場域代碼
						var areaNameArray []string //新場域名
						areaNameArray = append(areaNameArray, areaNumberNameMap[command.Area[0]])

						// 設定area
						if e, ok := clientInfoMap[clientPointer]; ok {

							oldArea = e.Device.Area           //暫存舊場域
							e.Device.Area = command.Area      //換成新場域代號
							e.Device.AreaName = areaNameArray //換成新場域名
							// info := clientInfoMap[clientPointer]
							// info.Device.Pic = command.Pic       // Pic
							// info.Device.RoomID = command.RoomID // RoomID
							// info.Device.DeviceStatus = 2        // 設備狀態:求助中
							// clientInfoMap[clientPointer] = info // 回存Map
						}

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details = `執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備廣播:包成Array:放入 Response Devices
						//deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array
						deviceArray := getArrayPointer(clientInfoMap[clientPointer].Device) // 包成array

						// 廣播給舊場域的
						processBroadcastingDeviceChangeStatusInSomeArea(whatKindCommandString, command, clientPointer, deviceArray, oldArea)

						// 廣播給現在新場域的連線裝置
						processBroadcastingDeviceChangeStatusInArea(whatKindCommandString, command, clientPointer, deviceArray)

						// logger
						details = `執行廣播，指令結束`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 12: // 加入房間 //未來要做多方通話再做

						// whatKindCommandString := `加入房間`

						// // 是否已登入(基本欄位，外層已經檢查過)
						// if !checkLogedIn(clientPointer, command, whatKindCommandString) {
						// 	break // 跳出case
						// }

						// // 檢查欄位是否齊全
						// if !checkFieldsCompleted([]string{"roomID"}, clientPointer, command, whatKindCommandString) {
						// 	break // 跳出case
						// }

						// // 當送來指令，更新心跳包通道時間
						// commandTimeChannel <- time.Now()

						// // logger：收到指令
						// allDevices := getAllDeviceByList() // 取得裝置清單-實體
						// go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// // 檢核:房號未被取用過
						// if command.RoomID > roomID {

						// 	// Response：失敗
						// 	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "房號未被取用過", command.TransactionID))
						// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// 	// logger:失敗
						// 	allDevices := getAllDeviceByList() // 取得裝置清單-實體
						// 	go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "房號未被取用過", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// 	go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "房號未被取用過", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// 	break // 跳出

						// }

						// // 設定加入設備的狀態+房間(自己)
						// element := clientInfoMap[clientPointer]
						// element.Device.DeviceStatus = 3        // 通話中
						// element.Device.RoomID = command.RoomID // 想回應的RoomID
						// clientInfoMap[clientPointer] = element // 存回Map

						// // Response:成功
						// jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
						// clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// // logger:成功
						// allDevices = getAllDeviceByList() // 取得裝置清單-實體
						// go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// // 準備廣播:包成Array:放入 Response Devices
						// deviceArray := getArray(clientInfoMap[clientPointer].Device)

						// // (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
						// if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

						// 	// 區域廣播(排除個人)
						// 	broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer)

						// 	// logger:區域廣播
						// 	allDevices := getAllDeviceByList() // 取得裝置清單-實體
						// 	go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// 	go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// } else {

						// 	// logger:json出錯
						// 	allDevices := getAllDeviceByList() // 取得裝置清單-實體
						// 	go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// 	go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// 	break // 跳出
						// }

					}

				} // else if ws.OpBinary == wsOpCode {
				//
				// 	clientPointer.fileExtensionMutexPointer.RLock()           // 鎖讀
				// 	clientPointerFileExtension := clientPointer.fileExtension // 取得附檔名
				// 	clientPointer.fileExtensionMutexPointer.RUnlock()         // 解鎖讀
				//
				// 	// 廣播檔名
				// 	broadcastHubWebsocketData(
				// 		websocketData{
				// 			wsOpCode:  ws.OpText,
				// 			dataBytes: []byte(writeToFile(&connection, clientPointerFileExtension, dataBytes)),
				// 		},
				// 	)
				//
				// 	clientPointer.fileExtensionMutexPointer.Lock()   // 鎖寫
				// 	clientPointer.fileExtension = ``                 // 儲存附檔名
				// 	clientPointer.fileExtensionMutexPointer.Unlock() // 解鎖寫
				//
				// 	go broadcastHubWebsocketData(websocketData{wsOpCode: wsOpCode, dataBytes: dataBytes}) // 廣播檔案內容
				// }

			}

		}

	}

}

// keepWriting - 保持寫入連線
func (clientPointer *client) keepWriting() {

	if nil != clientPointer { // 若指標不為空

		defer func() {
			disconnectHub(clientPointer) // 中斷客戶端與網路中心的連線
		}()

		connectionPointer := clientPointer.getConnectionPointer() // 連線指標

		if nil != connectionPointer { // 若指標不為空

			for { // 循環處理通道接收

				select {

				case websocketData := <-clientPointer.outputChannel: // 若客戶端輸出通道收到資料

					if !giveOutputWebsocketDataToConnection(connectionPointer, websocketData) { // 若不成功
						return // 回傳
					}

				} // end select

			} // end for

		}

	}

}

// HandleNewConnection - 處理新連線
/**
 * @param  *client newConnectionPointer  新連線指標
 */
func HandleNewConnection(newConnectionPointer *net.Conn) {

	if nil != newConnectionPointer { // 若連線指標不為空

		// 建立新客戶端指標
		clientPointer := &client{
			inputChannel:  make(chan websocketData, channelSize),
			outputChannel: make(chan websocketData, channelSize),
		}

		clientPointer.setConnectionPointer(newConnectionPointer) // 設定新連線

		go func() {
			connectHub(clientPointer) // 將客戶端連接中心輸出
		}()

		go clientPointer.keepReading() // 保持讀取連線

		go clientPointer.keepWriting() // 保持寫入連線

	}

}

// isOriginalFileName - 判斷是否為原檔名
func isOriginalFileName(inputFileName string) bool {
	return !regexp.MustCompile(`\d{4}(-\d{2}){5}-\d{7}(\.[a-zA-Z]+)?$`).MatchString(inputFileName)
}

// writeToFile - 將連線傳來的資料寫入檔案
/**
 * @param  *net.Conn connectionPointer  連線指標
 * @param  string fileExtension         副檔名
 * @param  []byte data                  資料
 * @return  string returnFileName       檔名
 */
func writeToFile(connectionPointer *net.Conn, fileExtension string, dataBytes []byte) (returnFileName string) {

	if nil != connectionPointer { // 若連線指標不為空

		path := configurations.GetConfigValueOrPanic(`local`, `path`) // 取得預設路徑

		paths.CreateIfPathNotExisted(path) // 若路徑不存在，則建立路徑

		// 修改現在時間自串成合法檔名
		fileName := regexp.MustCompile(`\s\+0800\sCST\sm=.*$`).ReplaceAllString(time.Now().String(), "")
		fileName = regexp.MustCompile(`[\s|\.|:]`).ReplaceAllString(fileName, `-`)
		fileName += fileExtension

		connection := *connectionPointer // 連線

		formerFormatStringSlices := []string{`%s %s `, `儲存 %s 的資料為 %v `} // 記錄器前半格式片段
		latterFormatStringSlices := []string{`: `, `檔案 %s%s `}           // 記錄器後格式片段

		// 記錄器預設參數
		defaultArgs := append(
			network.GetAliasAddressPair(connection.LocalAddr().String()),
			dataBytes,
			connection.RemoteAddr().String(), path, fileName,
		)

		ioutilWriteFileError := ioutil.WriteFile(path+fileName, dataBytes, 0777) // 寫入檔案

		formatString, args := logings.GetLogFuncFormatAndArguments(
			[]string{strings.Join(formerFormatStringSlices, `準備`), strings.Join(latterFormatStringSlices, `寫入`)},
			defaultArgs,
			ioutilWriteFileError,
		)

		if nil != ioutilWriteFileError { // 若寫入新檔案錯誤
			logger.Errorf(formatString, args...) // 記錄錯誤
			return                               // 回傳
		}

		go logger.Infof(formatString, args...) // 記錄資訊
		returnFileName = fileName              // 回傳檔名

	}

	return // 回傳
}

// OutputStringToConnection - 對連線輸出字串資料
/**
 * @param  *net.Conn connectionPointer  連線指標
 * @param  string dataString                  資料字串
 */
func OutputStringToConnection(connectionPointer *net.Conn, dataString string) {

	if nil != connectionPointer { // 若連線指標不為空

		// 對連線輸出websocket資料
		giveOutputWebsocketDataToConnection(
			connectionPointer,
			websocketData{wsOpCode: ws.OpText, dataBytes: []byte(dataString)},
		)

	}

}

// OutputBytesToConnection - 對連線輸出位元組資料
/**
 * @param  *net.Conn connectionPointer  連線指標
 * @param  string dataString                  資料字串
 */
func OutputBytesToConnection(connectionPointer *net.Conn, dataBytes []byte) {

	if nil != connectionPointer { // 若連線指標不為空

		// 對連線輸出websocket資料
		giveOutputWebsocketDataToConnection(
			connectionPointer,
			websocketData{wsOpCode: ws.OpBinary, dataBytes: dataBytes},
		)

	}

}
