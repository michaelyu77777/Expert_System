package networkHub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"../configurations"
	"../jwts"
	"../logings"
	"../network"
	"../paths"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/juliangruber/go-intersect"

	gomail "gopkg.in/gomail.v2"
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

	// 裝置Info
	DeviceID    string `json:"deviceID"`    //裝置ID
	DeviceBrand string `json:"deviceBrand"` //裝置品牌(怕平板裝置的ID會重複)
	DeviceType  int    `json:"deviceType"`  //裝置類型

	Area         []int    `json:"area"`         //場域代號
	AreaName     []string `json:"areaName"`     //場域名稱
	DeviceName   []string `json:"deviceName"`   //裝置名稱
	Pic          string   `json:"pic"`          //裝置截圖(求助截圖)
	OnlineStatus int      `json:"onlineStatus"` //在線狀態
	DeviceStatus int      `json:"deviceStatus"` //設備狀態
	CameraStatus int      `json:"cameraStatus"` //相機狀態
	MicStatus    int      `json:"micStatus"`    //麥克風狀態
	RoomID       int      `json:"roomID"`       //房號

	// 加密後字串
	AreaEncryptionString string `json:"areaEncryptionString"` //場域代號加密字串

}

// 客戶端Info
type Info struct {
	AccountPointer *Account `json:"account"` //使用者帳戶資料
	DevicePointer  *Device  `json:"device"`  //使用者登入密碼
}

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
	verificationCodeTime time.Time // 最後取得驗證碼之時間
}

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

// Response-心跳包
type Heartbeat struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// Response-登入
type LoginResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// Response-取得我的帳戶
type MyAccountResponse struct {
	Command       int     `json:"command"`
	CommandType   int     `json:"commandType"`
	ResultCode    int     `json:"resultCode"`
	Results       string  `json:"results"`
	TransactionID string  `json:"transactionID"`
	Account       Account `json:"account"`
}

// Response-取得我的裝置
type MyDeviceResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
	Device        Device `json:"device"`
}

// Response-取得所有線上Info
type InfosInTheSameAreaResponse struct {
	Command       int     `json:"command"`
	CommandType   int     `json:"commandType"`
	ResultCode    int     `json:"resultCode"`
	Results       string  `json:"results"`
	TransactionID string  `json:"transactionID"`
	Info          []*Info `json:"info"`
	//Device        []*Device `json:"device"`
}

// Response-求助
type HelpResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// Broadcast(廣播)-裝置狀態改變
type DeviceStatusChange struct {
	//指令
	Command     int      `json:"command"`
	CommandType int      `json:"commandType"`
	Device      []Device `json:"device"`
}

// Broadcast(廣播)-裝置狀態改變
type DeviceStatusChangeByPointer struct {
	//指令
	Command       int       `json:"command"`
	CommandType   int       `json:"commandType"`
	DevicePointer []*Device `json:"device"`
}

// Map-連線/登入資訊
var clientInfoMap = make(map[*client]*Info)

// 所有裝置清單
var allDevicePointerList = []*Device{}

// 所有帳號清單
var allAccountPointerList = []*Account{}

// 所有 area number 對應到 area name 名稱
var areaNumberNameMap = make(map[int]string)

// 加密解密KEY(AES加密)（key 必須是 16、24 或者 32 位的[]byte）
// const key_AES = "RB7Wfa$WHssV4LZce6HCyNYpdPPlYnDn" //32位數

const (

	// 代碼-指令
	CommandNumberOfLogout             = 8  //登出
	CommandNumberOfBroadcastingInArea = 10 //區域廣播
	CommandNumberOfBroadcastingInRoom = 11 //房間廣播

	// 代碼-指令類型
	CommandTypeNumberOfAPI         = 1 // 客戶端-->Server
	CommandTypeNumberOfAPIResponse = 2 // Server-->客戶端
	CommandTypeNumberOfBroadcast   = 3 // Server廣播
	CommandTypeNumberOfHeartbeat   = 4 // 心跳包

	// 代碼-結果
	ResultCodeSuccess = 0 // 成功
	ResultCodeFail    = 1 // 失敗
)

// 連線逾時時間
// const timeout = 30
var timeout = time.Duration(configurations.GetConfigPositiveIntValueOrPanic(`local`, `timeout`)) // 轉成time.Duration型態，方便做時間乘法

// demo 模式是否開啟
var expertdemoMode = configurations.GetConfigPositiveIntValueOrPanic(`local`, `expertdemoMode`)

// 房間號(總計)
var roomID = 0

// 基底: Response Json
var baseResponseJsonString = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"}`
var baseResponseJsonStringExtend = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"` // 可延展的

// 基底: 共用(指令執行成功、指令失敗、失敗原因、廣播、指令結束)
//var baseLoggerInfoCommonMessage = `指令<%s>:%s。Command:%+v、帳號:%+v、裝置:%+v、連線:%p、連線清單:%+v、裝置清單:%+v、,房號已取到:%d` // 普通紀錄
//var baseLoggerInfoCommonMessage = "指令名稱<%s>:%s。此指令%+v、此帳號%+v、此裝置%+v、此連線%+v。連線清單%+v、裝置清單:%+v。房號已取到:%d" // 普通紀錄
var baseLoggerCommonMessage = `
指令名稱<%s>:%s。
此指令%+v、
此帳號%+v、
此裝置%+v、
此連線%+v。
所有連線清單%+s、
所有裝置清單:%+v。
房號已取到:%d。

`

var baseLoggerInfoNilMessage = "<找到空指標Nil>:指令<%s>-%s-%s。Command:%+v"

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

// 匯入所有帳號到<帳號清單>中
func importAllAccountList() {

	picExpertA := getAccountPicString("pic/picExpertA.txt")
	picExpertB := getAccountPicString("pic/picExpertB.txt")
	picFrontline := getAccountPicString("pic/picFrontline.txt")
	picDefault := getAccountPicString("pic/picDefault.txt")

	//專家帳號 場域A
	accountExpertA := Account{
		UserID:               "expertA@leapsyworld.com",
		UserPassword:         "expertA@leapsyworld.com",
		UserName:             "專家-Adora",
		IsExpert:             1,
		IsFrontline:          2,
		Area:                 []int{1},
		AreaName:             []string{"場域A"},
		Pic:                  picExpertA,
		verificationCodeTime: time.Now().AddDate(1000, 0, 0), // 驗證碼永久有效時間1000年
	}
	//專家帳號 場域B
	accountExpertB := Account{
		UserID:               "expertB@leapsyworld.com",
		UserPassword:         "expertB@leapsyworld.com",
		UserName:             "專家-Belle",
		IsExpert:             1,
		IsFrontline:          2,
		Area:                 []int{2},
		AreaName:             []string{"場域B"},
		Pic:                  picExpertB,
		verificationCodeTime: time.Now().AddDate(1000, 0, 0), // 驗證碼永久有效時間1000年
	}

	//專家帳號 場域AB
	accountExpertAB := Account{
		UserID:               "expertAB@leapsyworld.com",
		UserPassword:         "expertAB@leapsyworld.com",
		UserName:             "專家-Abel",
		IsExpert:             1,
		IsFrontline:          2,
		Area:                 []int{1, 2},
		AreaName:             []string{"場域A", "場域B"},
		Pic:                  picExpertB,
		verificationCodeTime: time.Now().AddDate(1000, 0, 0), // 驗證碼永久有效時間1000年
	}

	//專家帳號 場域AB
	accountExpertPogo := Account{
		UserID:       "pogolin@leapsyworld.com",
		UserPassword: "pogolin@leapsyworld.com",
		UserName:     "專家-Pogo",
		IsExpert:     1,
		IsFrontline:  2,
		Area:         []int{1, 2},
		AreaName:     []string{"場域A", "場域B"},
		Pic:          picExpertB,
	}

	//專家帳號 場域AB
	accountExpertMichael := Account{
		UserID:       "michaelyu77777@gmail.com",
		UserPassword: "michaelyu77777@gmail.com",
		UserName:     "專家-Michael",
		IsExpert:     1,
		IsFrontline:  2,
		Area:         []int{1, 2},
		AreaName:     []string{"場域A", "場域B"},
		Pic:          picExpertB,
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

	//一線人員帳號
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

	//一線人員帳號2
	accountFrontLine2 := Account{
		UserID:       "frontLine2@leapsyworld.com",
		UserPassword: "frontLine2@leapsyworld.com",
		UserName:     "一線人員帳號",
		IsExpert:     2,
		IsFrontline:  1,
		Area:         []int{},
		AreaName:     []string{},
		Pic:          picFrontline,
	}

	allAccountPointerList = append(allAccountPointerList, &accountExpertA)
	allAccountPointerList = append(allAccountPointerList, &accountExpertB)
	allAccountPointerList = append(allAccountPointerList, &accountExpertAB)
	allAccountPointerList = append(allAccountPointerList, &accountExpertPogo)
	allAccountPointerList = append(allAccountPointerList, &accountExpertMichael)
	allAccountPointerList = append(allAccountPointerList, &defaultAccount)
	allAccountPointerList = append(allAccountPointerList, &accountFrontLine)
	allAccountPointerList = append(allAccountPointerList, &accountFrontLine2)
}

// 匯入所有裝置到<裝置清單>中
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

	// 假資料：平板Model（沒有場域之分）
	modelTab := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   2,       // 平版
		Area:         []int{}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     []string{},
		DeviceName:   "平板裝置", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",     // <求助>時才會從客戶端得到
		OnlineStatus: 2,      // 離線
		DeviceStatus: 0,      // 未設定
		MicStatus:    0,      // 未設定
		CameraStatus: 0,      // 未設定
		RoomID:       0,      // 無房間
	}

	// 新增假資料：場域A 平版Model
	// modelTabA := Device{
	// 	DeviceID:     "",
	// 	DeviceBrand:  "",
	// 	DeviceType:   2,        // 平版
	// 	Area:         []int{1}, // 依據裝置ID+Brand，從資料庫查詢
	// 	AreaName:     []string{"場域A"},
	// 	DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
	// 	Pic:          "",           // <求助>時才會從客戶端得到
	// 	OnlineStatus: 2,            // 離線
	// 	DeviceStatus: 0,            // 未設定
	// 	MicStatus:    0,            // 未設定
	// 	CameraStatus: 0,            // 未設定
	// 	RoomID:       0,            // 無房間
	// }

	// // 新增假資料：場域B 平版Model
	// modelTabB := Device{
	// 	DeviceID:     "",
	// 	DeviceBrand:  "",
	// 	DeviceType:   2,        // 平版
	// 	Area:         []int{2}, // 依據裝置ID+Brand，從資料庫查詢
	// 	AreaName:     []string{"場域B"},
	// 	DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
	// 	Pic:          "",           // <求助>時才會從客戶端得到
	// 	OnlineStatus: 2,            // 離線
	// 	DeviceStatus: 0,            // 未設定
	// 	MicStatus:    0,            // 未設定
	// 	CameraStatus: 0,            // 未設定
	// 	RoomID:       0,            // 無房間
	// }

	// 新增假資料：場域A 眼鏡
	var glassesPointerA [5]*Device
	for i, devicePointer := range glassesPointerA {
		device := modelGlassesA
		devicePointer = &device
		devicePointer.DeviceID = "00" + strconv.Itoa(i+1)
		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+1)
		glassesPointerA[i] = devicePointer
	}
	fmt.Printf("假資料眼鏡A=%+v\n", glassesPointerA)

	// 場域B 眼鏡
	var glassesPointerB [5]*Device
	for i, devicePointer := range glassesPointerB {
		device := modelGlassesB
		devicePointer = &device
		devicePointer.DeviceID = "00" + strconv.Itoa(i+6)
		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+6)
		glassesPointerB[i] = devicePointer
	}
	fmt.Printf("假資料眼鏡B=%+v\n", glassesPointerB)

	// 平版-0011
	var tabsPointerA [1]*Device
	for i, devicePointer := range tabsPointerA {
		device := modelTab
		devicePointer = &device
		devicePointer.DeviceID = "00" + strconv.Itoa(i+11)
		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+11)
		devicePointer.DeviceName = "平板00" + strconv.Itoa(i+11)
		tabsPointerA[i] = devicePointer
	}
	fmt.Printf("假資料平版-0011=%+v\n", tabsPointerA)

	// 平版-0012
	var tabsPointerB [1]*Device
	for i, devicePointer := range tabsPointerB {
		device := modelTab
		devicePointer = &device
		devicePointer.DeviceID = "00" + strconv.Itoa(i+12)
		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+12)
		devicePointer.DeviceName = "平板00" + strconv.Itoa(i+12)
		tabsPointerB[i] = devicePointer
	}
	fmt.Printf("假資料平版-0012=%+v\n", tabsPointerB)

	// 加入眼鏡A
	for _, e := range glassesPointerA {
		allDevicePointerList = append(allDevicePointerList, e)
	}
	// 加入眼鏡B
	for _, e := range glassesPointerB {
		allDevicePointerList = append(allDevicePointerList, e)
	}
	// 加入平版A
	for _, e := range tabsPointerA {
		allDevicePointerList = append(allDevicePointerList, e)
	}
	// 加入平版B
	for _, e := range tabsPointerB {
		allDevicePointerList = append(allDevicePointerList, e)
	}

	// fmt.Printf("\n取得所有裝置%+v\n", AllDeviceList)
	for _, e := range allDevicePointerList {
		fmt.Printf("資料庫所有裝置清單:%+v\n", e)
	}

}

// 匯入所有場域對應名稱
func importAllAreasNameToMap() {
	areaNumberNameMap[1] = "場域A"
	areaNumberNameMap[2] = "場域B"
}

// 取得某些場域的AllDeviceList
// func getAllDevicesPointerListByAreas(area []int) []*Device {

// 	result := []*Device{}

// 	for _, devicePointer := range allDevicePointerList {

// 		if devicePointer != nil {
// 			intersection := intersect.Hash(devicePointer.Area, area) //取交集array

// 			// 若場域有交集則加入
// 			if len(intersection) > 0 {
// 				result = append(result, devicePointer)
// 			}
// 		} else {
// 			// nil 發現
// 		}
// 	}
// 	return result
// }

// 登入(並處理重複登入問題)
/**
 * @param string whatKindCommandString 呼叫此函數的指令名稱
 * @param *client clientPointer 連線指標
 * @param Command command 客戶端傳來的指令
 * @param *Device newDevicePointer 登入之新裝置指標
 * @param *Account newAccountPointer 欲登入之新帳戶指標
 * @return bool isSuccess 回傳是否成功
 * @return string otherMessage 回傳詳細訊息
 */
func processLoginWithDuplicate(whatKindCommandString string, clientPointer *client, command Command, newDevicePointer *Device, newAccountPointer *Account) (isSuccess bool, messages string) {

	// 建立Info
	newInfoPointer := Info{
		AccountPointer: newAccountPointer,
		DevicePointer:  newDevicePointer,
	}

	// 登入步驟: 四種況狀判斷：相同連線、相異連線、不同裝置、相同裝置 重複登入之處理
	if infoPointer, ok := clientInfoMap[clientPointer]; ok {
		//相同連線
		messages += "-相同連線"

		//檢查裝置
		devicePointer := infoPointer.DevicePointer

		if devicePointer != nil {

			if (devicePointer.DeviceID == command.DeviceID) && (devicePointer.DeviceBrand == command.DeviceBrand) {
				// 裝置相同：同裝置重複登入
				messages += "-相同裝置"

				// 設定info
				infoPointer = &newInfoPointer

				devicePointer.OnlineStatus = 1 // 狀態為上線
				devicePointer.DeviceStatus = 1 // 裝置變閒置

			} else {
				// 裝置不同（現實中不會出現，只有測試才會出現）
				messages += "-不同裝置"

				// 將舊的裝置離線，並重設舊的裝置狀態
				if isSuccess, errMsg := resetDevicePointerStatus(devicePointer); !isSuccess {

					messages += "-重設舊裝置狀態失敗"

					//重設失敗
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
					processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
					return false, (messages + errMsg)
				}

				// 設定新info
				// 檢查
				if nil != newInfoPointer.DevicePointer {

					newInfoPointer.DevicePointer.OnlineStatus = 1 // 新的裝置＝上線
					newInfoPointer.DevicePointer.DeviceStatus = 1 // 裝置變閒置

					// 重要！將new info 指回clientInfoMap
					clientInfoMap[clientPointer] = &newInfoPointer

					//不需要斷線

				} else {
					//找不到新裝置
					messages += "-找不到新裝置"
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
					processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
					return false, messages
				}
			}
		} else {
			//失敗:找不到舊裝置
			messages += "-找不到舊裝置"
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
			processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
			return false, messages
		}

	} else {
		// 不同連線
		messages += `-發現不同連線`

		// 去找Map中有沒有已經存在相同裝置
		isExist, oldClient := isDeviceExistInClientInfoMap(newDevicePointer)

		if isExist {
			messages += `-相同裝置`
			// 裝置相同：（現實中，只有實體裝置重複ID才會實現）

			// 舊的連線：斷線＋從MAP移除＋Response舊的連線即將斷線
			if success, otherMessage := processDisconnectAndResponse(oldClient); success {
				otherMessage += messages + `-將重複裝置斷線`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, otherMessage, command, clientPointer) //所有值複製一份做logger
				processLoggerWarnf(whatKindCommandString, otherMessage, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
			} else {
				// 失敗:無此狀況
				return false, ""
			}

			// 新的連線，加入到Map，並且對應到新的裝置與帳號
			clientInfoMap[clientPointer] = &newInfoPointer
			// fmt.Printf("找到重複的連線，從Map中刪除，將此Socket斷線。\n")

			//檢查裝置
			devicePointer := clientInfoMap[clientPointer].DevicePointer

			if devicePointer != nil {
				devicePointer.OnlineStatus = 1 // 狀態為上線
				devicePointer.DeviceStatus = 1 // 裝置變閒置

			} else {
				//裝置為空

				// 警告logger
				messages += `-找不到此裝置`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
				processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				return false, messages

			}

		} else {
			//裝置不同：正常新增一的新裝置
			messages += `-不同裝置`

			clientInfoMap[clientPointer] = &newInfoPointer

			//檢查裝置
			devicePointer := clientInfoMap[clientPointer].DevicePointer
			if devicePointer != nil {

				devicePointer.OnlineStatus = 1 // 裝置狀態＝線上
				devicePointer.DeviceStatus = 1 // 裝置變閒置

			} else {
				//裝置為空

				// 警告logger
				messages += `-找不到此裝置`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
				processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				return false, messages
			}
		}
	}
	return true, messages
}

// 重設裝置狀態為預設狀態
/**
 * @param devicePointer *Device 裝置指標(想要重設的裝置)
 * @return isSuccess bool 回傳是否成功
 * @return messages string 回傳詳細訊息
 */
func resetDevicePointerStatus(devicePointer *Device) (isSuccess bool, messages string) {

	// 檢查裝置指標
	if nil != devicePointer {
		//成功
		devicePointer.OnlineStatus = 2
		devicePointer.DeviceStatus = 0
		devicePointer.CameraStatus = 0
		devicePointer.MicStatus = 0
		devicePointer.Pic = ""
		devicePointer.RoomID = 0
		return true, ``
	} else {
		//若找不到裝置指標
		return false, `-找不到裝置`
	}

}

// 設置裝置為離線
/**
 * @param devicePointer *Device 裝置指標(想要設置離線的裝置)
 * @return isSuccess bool 回傳是否成功
 * @return messages string 回傳詳細訊息
 */
func setDevicePointerOffline(devicePointer *Device) (isSuccess bool, messages string) {

	// 檢查裝置指標
	if nil != devicePointer {
		//成功
		devicePointer.OnlineStatus = 2
		devicePointer.DeviceStatus = 0
		devicePointer.CameraStatus = 0
		devicePointer.MicStatus = 0
		devicePointer.Pic = ""
		devicePointer.RoomID = 0
		return true, ``
	} else {
		//若找不到裝置指標
		return false, `-找不到裝置`
	}
}

// 將某連線斷線，並Response此連線
/**
 * @param clientPointer *client 連線指標(想要斷線的連線)
 * @return isSuccess bool 回傳是否成功
 * @return messages string 回傳詳細訊息
 */
func processDisconnectAndResponse(clientPointer *client) (isSuccess bool, messages string) {

	// 告知舊的連線，即將斷線
	// (讓logger可以進行平行處理，怕尚未執行到，就先刪掉了連線與裝置，就無法印出了)

	// Response:被斷線的連線:有裝置重複登入，已斷線
	details := `已斷線，有其他相同裝置ID登入伺服器`
	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, CommandNumberOfLogout, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, ""))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} // Response

	// logger:此斷線裝置的訊息
	details += `-此帳號與裝置為被斷線的連線裝置`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(``, details, Command{}, clientPointer) //所有值複製一份做logger
	processLoggerWarnf(``, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	fmt.Println("測試:已經進行斷線Response")

	// 舊的連線，從Map移除
	delete(clientInfoMap, clientPointer) // 此連線從Map刪除

	// 舊的連線，進行斷線
	disconnectHub(clientPointer) // 此連線斷線

	return true, `已將指定連線斷線`
}

// 查詢在線清單中(ClientInfoMap)，是否已經有相同裝置存在
/**
 * @param myDevicePointer *Device 裝置指標(想找的裝置)
 * @return bool 回傳結果(存在/不存在)
 * @return *client 回傳找到存在的連線指標(找不到則回傳nil)
 */
func isDeviceExistInClientInfoMap(myDevicePointer *Device) (bool, *client) {

	for clientPointer, infoPointer := range clientInfoMap {
		// 若找到相同裝置，回傳連線

		// 檢查傳入的裝置
		if myDevicePointer != nil {

			//檢查info
			if infoPointer != nil {

				//檢查裝置
				if infoPointer.DevicePointer != nil {

					if infoPointer.DevicePointer.DeviceID == myDevicePointer.DeviceID &&
						infoPointer.DevicePointer.DeviceBrand == myDevicePointer.DeviceBrand {
						return true, clientPointer
					}

				} else {
					//裝置為空,不額外處理
					return false, nil
				}
			} else {
				//info 為空,不額外處理
				return false, nil
			}
		} else {
			//傳入的裝置為空,不額外處理
			return false, nil
		}
	}
	return false, nil
}

// 根據裝置找到連線，關閉連線(排除自己這條連線)
// func findClientByDeviceAndCloseSocket(device *Device, excluder *client) {

// 	for client, e := range clientInfoMap {
// 		// 若找到相同裝置，關閉此client連線
// 		if e.DevicePointer.DeviceID == device.DeviceID && e.DevicePointer.DeviceBrand == device.DeviceBrand && client != excluder {
// 			delete(clientInfoMap, client) // 此連線從Map刪除
// 			disconnectHub(client)         // 此連線斷線
// 			fmt.Printf("找到重複的連線，從Map中刪除，將此Socket斷線。\n")
// 		}
// 	}

// }

// 取得帳號，透過UserID(加密的時候使用)
/**
 * @param userID string 想查找的UserID
 * @return accountPointer *Account 回傳找到的帳號指標
 */
func getAccountByUserID(userID string) (accountPointer *Account) {

	for _, accountPointer := range allAccountPointerList {
		// 檢查帳號
		if nil != accountPointer {
			// 若找到，直接返回帳號指標
			if userID == accountPointer.UserID {
				return accountPointer
			}
		}
	}

	// 沒找到帳號，直接回一個空的
	accountPointer = &Account{}
	return
}

// 確認密碼是否正確
/**
 * @param userID string 帳號
 * @param userPassword string 密碼
 * @return bool 回傳是否正確
 * @return *Account 回傳找到的帳號資料
 */
func checkPassword(userID string, userPassword string) (bool, *Account) {
	for _, accountPointer := range allAccountPointerList {

		// 帳號不為空
		if accountPointer != nil {
			if userID == accountPointer.UserID {

				//若為demo模式,且為測試帳號直接通過
				if (1 == expertdemoMode) &&
					("expertA@leapsyworld.com" == userID || "expertB@leapsyworld.com" == userID || "expertAB@leapsyworld.com" == userID) {
					return true, accountPointer
				} else {
					//非測試帳號 驗證密碼
					if userPassword == accountPointer.UserPassword {
						return true, accountPointer
					} else {
						return false, nil
					}
				}
			}
		}
	}
	return false, nil
}

// 判斷某連線是否已登入指令(若沒有登入，則直接RESPONSE給客戶端，並說明因尚未登入執行失敗)
/**
 * @param clientPointer *client 連線
 * @param command Command 客戶端傳來的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @return isLogedIn bool 回傳是否已登入
 */
func checkLogedInAndResponseIfFail(clientPointer *client, command Command, whatKindCommandString string) (isLogedIn bool) {

	// 若登入過
	if _, ok := clientInfoMap[clientPointer]; ok {

		isLogedIn = true
		return

	} else {
		// 尚未登入
		isLogedIn = false
		details := `-執行失敗，連線尚未登入`

		// 失敗:Response
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

		// 警告logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return
	}
}

// 判斷某連線裝置是否為閒置(判斷依據:eviceStatus)
/**
 * @param client *client 連線
 * @param command Command 客戶端傳來的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @param details string 之前已經處理的細節
 * @return bool 回傳是否為閒置
 */
func checkDeviceStatusIsIdleAndResponseIfFail(client *client, command Command, whatKindCommandString string, details string) bool {

	// 若連線存在
	if e, ok := clientInfoMap[client]; ok {
		details += `-找到連線`

		//檢查裝置
		devicePointer := e.DevicePointer
		if devicePointer != nil {
			details += `-找到裝置`

			// 若為閒置
			if devicePointer.DeviceStatus == 1 {
				// 狀態為閒置
				return true
			} else {
				// 狀態非閒置
				details += `-裝置狀態非閒置`

				// 失敗:Response
				jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
				client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

				// logger
				details += `-執行指令失敗`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, client) //所有值複製一份做logger
				processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return false
			}
		} else {
			// 裝置不存在
			// 失敗:Response
			details += `-找不到裝置`
			processResponseDeviceNil(client, whatKindCommandString, command, details)
			return false

		}

	} else {
		// 此連線不存在
		details += `-找不到連線`
		processResponseInfoNil(client, whatKindCommandString, command, details)
		return false

	}
}

// 判斷此裝置是否為眼鏡端(若不是的話，則直接RESPONSE給客戶端，並說明無法切換場域)
/**
 * @param clientPointer *client 連線指標
 * @param command Command 客戶端傳來的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @param details string 之前已經處理的細節
 * @return 是否為眼鏡端
 */
func checkDeviceTypeIsGlassesAndResponseIfFail(clientPointer *client, command Command, whatKindCommandString string, details string) bool {

	// 取連線
	if infoPointer, ok := clientInfoMap[clientPointer]; ok {
		details += "-找到連線"

		// 取裝置
		devicePointer := infoPointer.DevicePointer
		if devicePointer != nil {
			details += "-找到裝置"

			// 若為眼鏡端
			if devicePointer.DeviceType == 1 {
				// 成功
				return true
			} else {
				// 失敗Response:非眼鏡端
				details += "-非眼鏡端無法切換區域"

				// 失敗:Response
				jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
				clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

				// logger
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
				processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return false
			}
		} else {
			// 失敗Response:找不到裝置
			details += "-找不到裝置"
			processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
			return false
		}

	} else {
		//失敗:連線不存在
		details += "-找不到連線"
		processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
		return false
	}
}

// 判斷某連線是否已登入(用連線存不存在決定)
/**
 * @param client *client 連線指標
 * @return bool 回傳結果
 */
func checkLogedInByClient(client *client) bool {

	logedIn := false

	if _, ok := clientInfoMap[client]; ok {
		logedIn = true
	}

	return logedIn

}

// // 從清單移除某裝置
// func removeDeviceFromList(slice []*Device, s int) []*Device {
// 	return append(slice[:s], slice[s+1:]...) //回傳移除後的array
// }

// // 取得所有裝置清單 By clientInfoMap（For Logger）
// func getOnlineDevicesByClientInfoMap() []Device {

// 	deviceArray := []Device{}

// 	for _, info := range clientInfoMap {
// 		device := info.DevicePointer
// 		deviceArray = append(deviceArray, *device)
// 	}

// 	return deviceArray
// }

// 取得所有匯入的裝置清單COPY副本(For Logger)
func getAllDeviceByList() []Device {
	deviceArray := []Device{}
	for _, d := range allDevicePointerList {
		deviceArray = append(deviceArray, *d)
	}
	return deviceArray
}

// 取得某裝置指標
/**
 * @param deviceID string 裝置ID
 * @param deviceBrand string 裝置Brand
 * @return result *Device 回傳裝置指標
 */
func getDevice(deviceID string, deviceBrand string) (result *Device) {

	// 若找到則返回
	for _, devicePointer := range allDevicePointerList {
		if nil != devicePointer {
			if devicePointer.DeviceID == deviceID {
				if devicePointer.DeviceBrand == deviceBrand {
					result = devicePointer
					// fmt.Println("找到裝置", result)
					// fmt.Println("所有裝置清單", allDevicePointerList)
				}
			}
		}
	}

	return // 回傳
}

// // 取得裝置:同區域＋同類型＋去掉某一裝置（自己）
// func getDevicesByAreaAndDeviceTypeExeptOneDevice(area []int, deviceType int, device *Device) []*Device {

// 	result := []*Device{}

// 	// 若找到則返回
// 	for _, e := range allDevicePointerList {
// 		intersection := intersect.Hash(e.Area, area) //取交集array

// 		// 同區域＋同類型＋去掉某一裝置（自己）
// 		if len(intersection) > 0 && e.DeviceType == deviceType && e != device {

// 			result = append(result, e)

// 		}
// 	}

// 	fmt.Printf("找到指定場域＋指定類型＋排除自己的所有裝置:%+v \n", result)
// 	return result // 回傳
// }

// 針對<專家＋平版端>，組合出所有連線指標InfoPointer(裝置+帳號組合):符合與專家同場域(某場域)＋裝置為眼鏡(某裝置類型)＋去掉自己(某一裝置)
/**
 * @param myArea []int 想查的場域
 * @param someDeviceType int 想查的裝置類型(眼鏡/平板)
 * @param myDevice *Device 排除廣播的自己裝置(眼鏡/平板)
 * @return resultInfoPointers []*Info 回傳組合IngoPointer的結果
 * @return otherMeessage string 回傳詳細狀況
 */
func getDevicesWithInfoByAreaAndDeviceTypeExeptOneDevice(myArea []int, someDeviceType int, myDevice *Device) (resultInfoPointers []*Info, otherMeessage string) {

	fmt.Printf("測試：目標要找 myArea =%+v", myArea)

	// 若找到則返回
	for _, devicePointer := range allDevicePointerList {

		if nil != devicePointer {
			// 找到裝置
			otherMeessage += "-找到裝置"

			intersection := intersect.Hash(devicePointer.Area, myArea) //場域交集array
			fmt.Printf("\n\n 找交集 intersection =%+v, device=%s", intersection, devicePointer.DeviceID)

			// 有相同場域 + 同類型 + 去掉某一裝置（自己）
			if (len(intersection) > 0) && (devicePointer.DeviceType == someDeviceType) && (devicePointer != myDevice) {

				otherMeessage += "-找到相同場域"
				fmt.Printf("\n\n 找到交集 intersection =%+v, device=%s", intersection, devicePointer.DeviceID)

				// 準備進行同場域info包裝，針對空Account進行處理

				// 裝置在線，取出info
				if 1 == devicePointer.OnlineStatus {
					infoPointer := getInfoByOnlineDevice(devicePointer)

					//若有找到則加入結果清單
					if nil != infoPointer {
						resultInfoPointers = append(resultInfoPointers, infoPointer) // 加到結果
						fmt.Printf("\n\n 找到在線裝置=%+v,帳號=%+v", devicePointer, infoPointer.AccountPointer)
					}

				} else {
					//裝置離線，給空的Account
					emptyAccountPointer := &Account{}
					infoPointer := &Info{AccountPointer: emptyAccountPointer, DevicePointer: devicePointer}
					resultInfoPointers = append(resultInfoPointers, infoPointer) // 加到結果
					fmt.Printf("\n\n 找到離線裝置=%+v,帳號=%+v", devicePointer, infoPointer.AccountPointer)
				}
			}
		} else {
			// 裝置為空 不做事
			otherMeessage += "-裝置為空"
		}

	}

	fmt.Printf("找到指定場域＋指定裝置類型＋排除自己的所有裝置，結果為:%+v \n", resultInfoPointers)
	return // 回傳
}

// 取得在線裝置所對應的Info
/**
 * @param devicePointer *Device 某在線裝置指標
 * @return *Info 回傳對應的連線指標
 */
func getInfoByOnlineDevice(devicePointer *Device) *Info {

	for _, infoPointer := range clientInfoMap {

		// 若找到裝置對應的info
		if devicePointer == infoPointer.DevicePointer {
			return infoPointer
		}
	}

	//找不到就回傳NIL
	return nil

	//找不到就回傳空的
	//return &Info{}
}

// // 排除某連線進行廣播 (excluder 被排除的client)
// func broadcastExceptOne(excluder *client, websocketData websocketData) {

// 	//Response to all
// 	for clientPointer, _ := range clientInfoMap {

// 		// 僅排除一個連線
// 		if nil != clientPointer && clientPointer != excluder {

// 			clientPointer.outputChannel <- websocketData //Socket Response

// 		}
// 	}
// }

// // 針對指定群組進行廣播，排除某連線(自己)
// func broadcastByGroup(clientPinters []*client, websocketData websocketData, excluder *client) {

// 	for _, clientPointer := range clientPinters {

// 		//排除自己
// 		if nil != clientPointer && clientPointer != excluder {
// 			clientPointer.outputChannel <- websocketData //Socket Respone
// 		}
// 	}
// }

// 針對某場域(Area)進行廣播，排除某連線(自己)
/**
 * @param area []int 想廣播的區域代碼
 * @param websocketData websocketData 想廣播的內容
 * @param whatKindCommandString string 是哪個指令呼叫此函式
 * @param command Command 客戶端的指令
 * @param excluder *client 排除廣播的客戶端連線指標(通常是自己)
 * @param details string 詳細資訊
 */
func broadcastByArea(area []int, websocketData websocketData, whatKindCommandString string, command Command, excluder *client, details string) {

	for clientPointer, infoPointer := range clientInfoMap {

		// 檢查nil
		if nil != infoPointer.DevicePointer {

			// 找相同的場域
			myArea := getMyAreaByClientPointer(whatKindCommandString, command, clientPointer, details) //取每個clientPointer的場域
			intersection := intersect.Hash(myArea, area)                                               //取交集array

			fmt.Println("測試中：同場域有哪些場域：intersection=", intersection, "裝置ID＝", infoPointer.DevicePointer.DeviceID)

			// 若有找到
			if len(intersection) > 0 {

				fmt.Println("測試中：找到有交集的裝置為：intersection=", intersection, "裝置ID＝", infoPointer.DevicePointer.DeviceID)

				//if clientInfoMap[client].DevicePointer.Area == area {
				if clientPointer != excluder { //排除自己
					fmt.Println("測試中：正在廣播：intersection=", intersection)

					// 廣播
					otherMessage := "-對某裝置進行廣播-裝置ID=" + infoPointer.DevicePointer.DeviceID
					clientPointer.outputChannel <- websocketData //Socket Response

					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details+otherMessage, command, clientPointer) //所有值複製一份做logger
					processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				}
			}
		}
	}
}

// 針對某房間(RoomID)進行廣播，排除某連線(自己)
/**
 * @param roomID int 欲廣播的房間號
 * @param websocketData websocketData 欲廣播的內容
 * @param excluder *client 欲排除的連線指標(通常是自己)
 */
func broadcastByRoomID(roomID int, websocketData websocketData, excluder *client) {

	for clientPointer, infoPointer := range clientInfoMap {
		// 檢查nil
		if nil != infoPointer.DevicePointer {
			// 找到相同房間的連線
			if infoPointer.DevicePointer.RoomID == roomID {

				if clientPointer != excluder { //排除自己

					// 廣播
					clientPointer.outputChannel <- websocketData //Socket Response
				}
			}
		}
	}
}

// 取得某clientPointer的場域：(眼鏡端：取眼鏡場域，平版端：取專家場域)
/**
 * @param whatKindCommandString string 是哪個指令呼叫此函式
 * @param command Command 客戶端的指令
 * @param clientPointer *client 連線指標
 * @param details string 之前已處理的詳細訊息
 * @return area []int 回傳查詢到的場域代碼
 */
func getMyAreaByClientPointer(whatKindCommandString string, command Command, clientPointer *client, details string) (area []int) {

	if infoPointer, ok := clientInfoMap[clientPointer]; ok {

		devicePointer := infoPointer.DevicePointer

		if nil != devicePointer {
			// 若自己為眼鏡端
			details += "-找到連線"
			if devicePointer.DeviceType == 1 {
				// 取眼鏡場域
				details += "-取眼鏡場端場域"
				area = devicePointer.Area
				fmt.Println("測試：我是眼鏡端取裝置場域＝", area)
				return

				// 若自己為平板端
			} else if devicePointer.DeviceType == 2 {
				//取專家帳號場域
				details += "-取專家場域"
				accountPointer := infoPointer.AccountPointer
				if nil != accountPointer {
					details += "-找到專家帳號"
					area = accountPointer.Area
					fmt.Println("測試：我是專家端取專家場域", area)
					return
				} else {
					//找不到帳號
					details += "-找不到帳號"

					// 警告logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
					processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				}
			}
		} else {
			//找不到裝置
			details += "-找不到裝置"

			// 警告logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		}
	} else {
		//找不到連線
		details += "-找不到連線"

		// 警告logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	}

	return []int{}
}

// 將裝置指標 包裝成 裝置實體array
/**
 * @param 裝置指標
 * @return 裝置實體array
 */
func getArray(device *Device) []Device {
	var array = []Device{}
	array = append(array, *device)
	return array
}

// 將裝置指標 包裝成 裝置指標array
/**
 * @param 裝置指標
 * @return 裝置指標array
 */
func getArrayPointer(device *Device) []*Device {
	var array = []*Device{}
	array = append(array, device)
	return array
}

// 檢查Command的指定欄位是否齊全(command 非指標 不用檢查Nil問題)
/**
 * @param command Command 客戶端的指令
 * @param fields []string 檢查的欄位名稱array
 * @return ok bool 回傳是否齊全
 * @return missFields []string 回傳遺漏的欄位名稱
 */
func checkCommandFields(command Command, fields []string) (ok bool, missFields []string) {

	missFields = []string{} // 遺失的欄位
	ok = true               // 是否齊全

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

		case "area":
			if command.Area == nil {
				missFields = append(missFields, field)
				ok = false
			}

		case "areaName":
			if command.AreaName == nil {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceName":
			if command.DeviceName == nil {
				missFields = append(missFields, field)
				ok = false
			}

		case "pic":
			if command.Pic == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "onlineStatus":
			if command.OnlineStatus == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceStatus":
			if command.DeviceStatus == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "cameraStatus":
			if command.CameraStatus == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "micStatus":
			if command.MicStatus == 0 {
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

// 判斷欄位Command(個別指令專屬欄位)是否齊全，若不齊全直接Response給客戶端
/**
 * @param fields []string 檢查的欄位名稱array
 * @param clientPointer *client 連線指標
 * @param command Command 客戶端的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @return bool 回傳是否齊全
 */
func checkFieldsCompletedAndResponseIfFail(fields []string, clientPointer *client, command Command, whatKindCommandString string) bool {

	//fields := []string{"roomID"}
	ok, missFields := checkCommandFields(command, fields)

	if !ok {

		m := strings.Join(missFields, ",")

		details := `-欄位不齊全:` + m

		// Response: 失敗
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

		// 警告logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return false

	} else {

		return true
	}

}

// // 取得連線資訊物件
// // 取得登入基本資訊字串
// func getLoginBasicInfoString(c *client) string {

// 	s := " UserID:" + clientInfoMap[c].AccountPointer.UserID

// 	if clientInfoMap[c].DevicePointer != nil {
// 		s += ",DeviceID:" + clientInfoMap[c].DevicePointer.DeviceID
// 		s += ",DeviceBrand:" + clientInfoMap[c].DevicePointer.DeviceBrand
// 		s += ",DeviceName:" + clientInfoMap[c].DevicePointer.DeviceName
// 		//s += ",DeviceType:" + (string)(clientInfoMap[c].DevicePointer.DeviceType)
// 		s += ",DeviceType:" + strconv.Itoa(clientInfoMap[c].DevicePointer.DeviceType)                                      // int轉string
// 		s += ",Area:" + strings.Replace(strings.Trim(fmt.Sprint(clientInfoMap[c].DevicePointer.Area), "[]"), " ", ",", -1) //將int[]內容轉成string
// 		s += ",RoomID:" + strconv.Itoa(clientInfoMap[c].DevicePointer.RoomID)                                              // int轉string
// 		s += ",OnlineStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.OnlineStatus)                                  // int轉string
// 		s += ",DeviceStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.DeviceStatus)                                  // int轉string
// 		s += ",CameraStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.CameraStatus)                                  // int轉string
// 		s += ",MicStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.MicStatus)                                        // int轉string
// 	}

// 	return s
// }

// 確認是否有此帳號
/**
 * @param id string 使用者帳號
 * @return bool 回傳是否存在此帳號
 * @return *Account 回傳帳號指標,若不存在回傳nil
 */
func checkAccountExist(id string) (bool, *Account) {

	for _, accountPointer := range allAccountPointerList {
		if nil != accountPointer {
			if id == accountPointer.UserID {
				return true, accountPointer
			}
		}
	}
	//找不到帳號
	return false, nil
}

// Struct結構: 儲存email之夾帶內容
type mailInfo struct {
	VerificationCode string
}

// 寄送郵件功能
/**
 * @receiver myInfo mailInfo 可以使用此函數的主體結構為 mailInfo 的實體變數，可使用實體變數名稱.sendMail() 來直接使用此函數)
 * @param accountPointer *Account 帳戶指標
 * @return success bool 回傳寄送成功或失敗
 * @return otherMessage string 回傳處理的細節
 */
func (myInfo mailInfo) sendMail(accountPointer *Account) (success bool, otherMessage string) {

	var err error

	// 帳號不為空
	if nil != accountPointer {

		//建立樣板物件
		t := template.New(`templateVerificationCode.html`) //物件名稱
		fmt.Println("測試", t)

		//轉檔
		t, err = t.ParseFiles(`./template/templateVerificationCode.html`)
		if err != nil {
			log.Println(err)
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, myInfo); err != nil {
			log.Println(err)
		}

		//取得樣板文字結果
		result := tpl.String()

		//建立與設定電子郵件
		m := gomail.NewMessage()
		m.SetHeader("From", "sw@leapsyworld.com")
		m.SetHeader("To", accountPointer.UserID) //Email 就是 userID

		//副本
		//m.SetAddressHeader("Cc", "<RECIPIENT CC>", "<RECIPIENT CC NAME>")

		m.SetHeader("Subject", "Leapsy專家系統-驗證通知信")
		m.SetBody("text/html", result)

		//夾帶檔案
		//m.Attach("template.html") // attach whatever you want

		d := gomail.NewDialer("smtp.qiye.aliyun.com", 25, "sw@leapsyworld.com", "Leapsy123!")

		//寄發電子郵件
		if err := d.DialAndSend(m); err != nil {
			// 寄信發生錯誤
			//panic(err)
			otherMessage = "-寄信發生錯誤"
			success = false
		} else {
			otherMessage = "-順利寄出"
			// 紀錄驗證信發送時間
			accountPointer.verificationCodeTime = time.Now()
			success = true
			fmt.Println("順利寄出", success)
		}
	} else {
		// 帳號為空
		otherMessage = "-帳號為空"
		success = false
	}

	return
}

// 產生驗證碼並寄送郵件
/**
 * @param accountPointer *Account 帳戶指標
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @param details string 之前已經處理過的細節
 * @param command Command 客戶端的指令
 * @param clientPointer *client 連線指標
 * @return success bool 回傳寄送成功或失敗
 * @return returnMessages string 回傳處理的細節
 */
func processSendVerificationCodeMail(accountPointer *Account, whatKindCommandString string, details string, command Command, clientPointer *client) (success bool, returnMessages string) {

	// 建立隨機密string六碼
	verificationCode := strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10))

	//給logger用的
	details += `-建立隨機六碼,verificationCode=` + verificationCode

	//可能會回給前端用的
	returnMessages += "-建立驗證碼六碼"

	// 正常logger
	fmt.Println(details)
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	//密碼記在account中:(未來驗證碼不用紀錄在資料庫，因為設計中，密碼是一次性密碼。即使主機當機，所有USER也需要重新登入，自然要重新取一次驗證信)
	if nil != accountPointer {

		details += "-找到帳號,userID=" + accountPointer.UserID
		returnMessages += "-找到帳號,userID=" + accountPointer.UserID

		// 儲存隨機六碼到帳戶
		accountPointer.UserPassword = verificationCode

		details += "-帳戶已更新驗證碼"
		returnMessages += "-帳戶已更新驗證碼"

		// 準備寄送包含密碼的EMAIL
		d := mailInfo{verificationCode}
		// emailString := accountPointer.UserID

		// 寄送郵件
		ok, errMsg := d.sendMail(accountPointer)

		if ok {
			// 已寄出
			success = true
			details += "-驗證碼已成功寄出"
			returnMessages += "-驗證碼已成功寄出"

			// 正常logger
			fmt.Println(details)
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return
		} else {
			// 寄出失敗:請確認您的電子信箱是否正確
			success = false
			details += "-驗證碼寄出失敗:請確認您的電子信箱是否正確-錯誤訊息:" + errMsg
			returnMessages += "-驗證碼寄出失敗:請確認您的電子信箱是否正確-錯誤訊息:" + errMsg

			// 警告logger
			fmt.Println(details)
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return
		}
	} else {
		// accountPointer 為nil
		success = false
		details += "-帳號不存在"
		returnMessages += "-帳號不存在"

		// 警告logger
		fmt.Println(details)
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return
	}

	// return
	//額外:進行登出時要去把對應的password移除
}

// 處理<我的裝置區域>的廣播，廣播內容為某些裝置的狀態變更
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command Command 客戶端的指令
* @param clientPointer *client 連線指標(我的裝置區域從此變數來)
* @param devicePointerArray []*Device 要廣播出去的所有Device內容
* @param details string 之前已經處理過的細節
* @return string 回傳處理的細節
**/
func processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString string, command Command, clientPointer *client, devicePointerArray []*Device, details string) string {

	// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, DevicePointer: devicePointerArray}); err == nil {
		//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

		// 準備廣播

		// 取出自己的場域
		area := getMyAreaByClientPointer(whatKindCommandString, command, clientPointer, details)

		if len(area) > 0 {
			// 若有找到場域
			strArea := fmt.Sprintln(area)
			details += `-執行（場域）廣播成功,場域代碼=` + strArea

			// 廣播(場域、排除個人)
			broadcastByArea(area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, whatKindCommandString, command, clientPointer, details) // 排除個人進行Area廣播

			// 一般logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return details

		} else {
			// 若沒找到場域
			details += `-執行（場域）廣播失敗,沒找到自己的場域`

			// 警告logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
			return details
		}

	} else {

		// 錯誤logger
		details += `-執行（場域）廣播失敗,後端json轉換出錯`
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return details
	}
}

// 處理<指定區域>的廣播，廣播內容為某些裝置狀態的變更
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command Command 客戶端的指令
* @param clientPointer *client 連線指標(我的區域從此變數來)
* @param device []*Device 要廣播出去的所有Device內容
* @param area []int 想廣播的區域
* @param details string 之前已經處理過的細節
**/
func processBroadcastingDeviceChangeStatusInSomeArea(whatKindCommandString string, command Command, clientPointer *client, device []*Device, area []int, details string) {

	// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, DevicePointer: device}); err == nil {
		//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

		// 廣播(場域、排除個人)
		broadcastByArea(area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, whatKindCommandString, command, clientPointer, details) // 排除個人進行Area廣播

		// logger
		details := `執行（指定場域）廣播成功-場域代碼=` + strconv.Itoa(area[0])
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	} else {

		// logger
		details := `執行（指定場域）廣播失敗：後端json轉換出錯`
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	}
}

// 處理<我的裝置房間>的廣播，廣播內容為我的裝置狀態的變更
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command Command 客戶端的指令
* @param clientPointer *client 連線指標(我的房號從此變數來)
* @param devicePointerArray []*Device 要廣播出去的所有Device內容
* @param details string 之前已經處理過的細節
* @return string 回傳處理的細節
**/
func processBroadcastingDeviceChangeStatusInRoom(whatKindCommandString string, command Command, clientPointer *client, devicePointerArray []*Device, details string) string {

	// (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInRoom, CommandType: CommandTypeNumberOfBroadcast, DevicePointer: devicePointerArray}); err == nil {

		var roomID = 0

		if infoPointer, ok := clientInfoMap[clientPointer]; ok {
			devicePointer := infoPointer.DevicePointer
			if nil != devicePointer {
				// 找到房號
				roomID = devicePointer.RoomID
				details += `-找到裝置ID=` + devicePointer.DeviceID + `,裝置品牌=` + devicePointer.DeviceBrand + `,與裝置房號=` + strconv.Itoa(roomID)

				// 房間廣播:改變麥克風/攝影機狀態
				broadcastByRoomID(roomID, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

				// 一般logger
				details += `-執行（房間）廣播成功,於房間號碼=` + strconv.Itoa(roomID)
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
				processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return details
			} else {
				//找不到裝置
				details += `-找不到裝置,執行（房間）廣播失敗`

				// 警告logger
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
				processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return details
			}
		} else {
			//找不到連線
			details += `-找不到連線,執行（房間）廣播失敗`

			// 警告logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return details
		}

	} else {
		// 後端json轉換出錯

		details += `-後端json轉換出錯,執行（房間）廣播失敗`

		// 錯誤logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return details
	}
}

// 取得某場域的線上閒置專家數
/**
* @param area []int 想要計算的場域代碼array
* @param whatKindCommandString string 是哪個指令呼叫此函數 (for log)
* @param command Command 客戶端的指令 (for log)
* @param clientPointer *client 連線指標 (for log)
* @return int 回傳結果
**/
func getOnlineIdleExpertsCountInArea(area []int, whatKindCommandString string, command Command, clientPointer *client) int {

	counter := 0
	for _, e := range clientInfoMap {

		//找出同場域的專家帳號＋裝置閒置
		accountPointer := e.AccountPointer
		if accountPointer != nil {

			if 1 == accountPointer.IsExpert {

				intersection := intersect.Hash(accountPointer.Area, area) //取交集array

				// 若場域有交集
				if len(intersection) > 0 {

					//是否閒置
					if 1 == e.DevicePointer.DeviceStatus {
						counter++
					}
				}
			}
		} else {
			//帳號為空 不Response特別處理 僅記錄
			//logger
			details := `-發現accountPointer為空`
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		}
	}

	return counter
}

//取得同房間其他人連線
// func getOtherClientsInTheSameRoom(clientPoint *client, roomID int) []*client {

// 	results := []*client{}

// 	for i, e := range clientInfoMap {
// 		if clientPoint != i {
// 			if roomID == e.DevicePointer.RoomID {
// 				results = append(results, i)
// 			}
// 		}
// 	}

// 	return results
// }

// 取得同房號的其他所有人裝置指標
/**
* @param roomID int (房號)
* @param clientPoint *client 排除的連線指標(通常為自己)
* @return []*Device 回傳結果
**/
func getOtherDevicesInTheSameRoom(roomID int, clientPoint *client) []*Device {

	results := []*Device{}

	for cPointer, infoPointer := range clientInfoMap {

		// 排除自己
		if clientPoint != cPointer {

			//取連線
			if nil != infoPointer {
				//取裝置
				if nil != infoPointer.DevicePointer {
					//找同房間
					if roomID == infoPointer.DevicePointer.RoomID {
						//加入結果清單
						results = append(results, infoPointer.DevicePointer)
					}
				} else {
					//找不到裝置
				}
			} else {
				// 找不到連線
			}
		}
	}

	return results
}

// 取得所有Log需要的參數(都取COPY，為了記錄當時狀況，避免平行處理值改變了)
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param details string 之前已經處理過的細節
* @param command Command 客戶端的指令
* @param clientPointer *client 連線指標

* @return myAccount Account 帳戶實體COPY
* @return myDevice Device 裝置實體COPY
* @return myClient client 連線實體COPY
* @return myClientInfoMap map[*client]*Info 連線與Info對應Map實體COPY
* @return myAllDevices []Device 所有裝置清單實體COPY (為了印出log，先而取出所有實體，若使用pointer無法直接透過%+v印出)
* @return nowRoomId int 最後取到的房號COPY
**/
func getLoggerParrameters(whatKindCommandString string, details string, command Command, clientPointer *client) (myAccount Account, myDevice Device, myClient client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomId int) {

	if clientInfoMap != nil {
		myClientInfoMap = clientInfoMap

		if clientPointer != nil {
			myClient = *clientPointer

			if infoPointer, ok := clientInfoMap[clientPointer]; ok {

				myAccount = *infoPointer.AccountPointer
				myDevice = *infoPointer.DevicePointer

				// if e.AccountPointer != nil {
				// 	myAccount = *e.AccountPointer
				// }

				// if e.DevicePointer != nil {
				// 	myDevice = *e.DevicePointer
				// }
			}
		}
	}

	myAllDevices = getAllDeviceByList() // 取得裝置清單-實體(為了印出log，先而取出所有實體，若使用pointer無法直接透過%+v印出)
	nowRoomId = roomID                  // 目前房號

	return
}

// 處理發現nil的logger
// func processNilLoggerInfof(whatKindCommandString string, details string, otherMessages string, command Command) {

// 	go fmt.Printf(baseLoggerInfoNilMessage+"\n", whatKindCommandString, details, otherMessages, command)
// 	go logger.Infof(baseLoggerInfoNilMessage, whatKindCommandString, details, otherMessages, command)

// }

// 處理<一般logger>
/**
* @param whatKindCommandString string 哪個指令發出的
* @param details string 詳細訊息
* @param command Command 客戶端的指令
* @param myAccount Account 帳戶(連線本身)
* @param myDevice Device 裝置(連線本身)
* @param myClientPointer client 連線
* @param myClientInfoMap map[*client]*Info 所有在線連線的info(裝置+帳戶之配對)
* @param myAllDevices []Device 所有已匯入裝置的狀態訊息
* @param nowRoomID int 線在房號
**/
func processLoggerInfof(whatKindCommandString string, details string, command Command, myAccount Account, myDevice Device, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏

	strClientInfoMap := getStringOfClientInfoMap() //所有連線、裝置、帳號資料
	go fmt.Printf(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)
	go logger.Infof(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)

}

// 處理<警告logger>
/**
* @param 參數同處理<一般logger>
**/
func processLoggerWarnf(whatKindCommandString string, details string, command Command, myAccount Account, myDevice Device, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏

	strClientInfoMap := getStringOfClientInfoMap() //所有連線、裝置、帳號資料
	go fmt.Printf(baseLoggerCommonMessage+"\n\n", whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)
	go logger.Warnf(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)

}

// 處理<錯誤logger>
/**
* @param 參數同處理<一般logger>
**/
func processLoggerErrorf(whatKindCommandString string, details string, command Command, myAccount Account, myDevice Device, myClientPointer client, myClientInfoMap map[*client]*Info, myAllDevices []Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏

	strClientInfoMap := getStringOfClientInfoMap() //所有連線、裝置、帳號資料
	go fmt.Printf(baseLoggerCommonMessage+"\n\n", whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)
	go logger.Errorf(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)

}

// 取得帳號圖片
/**
* @param fileName string 檔名
* @return string 回傳檔案內容
**/
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

// // 檢查clientInfoMap 是否有nil pointer狀況
// func checkAndGetClientInfoMapNilPoter(whatKindCommandString string, details string, command Command, clientPointer *client) (myInfoPointer *Info, myDevicePointer *Device, myAccountPointer *Account) {

// 	myInfoPointer = &Info{}
// 	myDevicePointer = &Device{}
// 	myAccountPointer = &Account{}

// 	if e, ok := clientInfoMap[clientPointer]; ok {

// 		if e.DevicePointer != nil {
// 			myDevicePointer = e.DevicePointer
// 		} else {

// 			// logger
// 			details := `發現DevicePointer為nil`
// 			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
// 			processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

// 		}

// 		if e.AccountPointer != nil {
// 			myAccountPointer = e.AccountPointer
// 		} else {

// 		}

// 	} else {
// 		//info is nil
// 	}

// 	return
// }

// 處理連線Info為空Response給客戶端
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func processResponseInfoNil(clientPointer *client, whatKindCommandString string, command Command, details string) {
	// Response:失敗
	details += `-執行失敗-找不到連線`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現infoPointer為空`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 處理帳號為空Response給客戶端
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func processResponseAccountNil(clientPointer *client, whatKindCommandString string, command Command, details string) {
	// Response:失敗
	details += `-執行失敗-找不到帳號`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現accountPointer為空`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 處理裝置為空Response給客戶端
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func processResponseDeviceNil(clientPointer *client, whatKindCommandString string, command Command, details string) {
	// Response:失敗
	details += `-執行失敗-找不到裝置`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現devicePointer為空`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 處理某指標為空Response
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func processResponseNil(clientPointer *client, whatKindCommandString string, command Command, details string) {
	// Response:失敗
	details += `-執行失敗`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現空指標`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 取得log字串:針對ClientInfoMap(即所有在線的連線、裝置、帳號配對)
func getStringOfClientInfoMap() (results string) {

	for myClient, myInfo := range clientInfoMap {
		results += fmt.Sprintf(`【連線%v,裝置%v,帳號%v】
		
		`, myClient, myInfo.DevicePointer, myInfo.AccountPointer)
	}

	return
}

// 取得log字串:針對指定的 info array(可用來取得紀錄平查詢到的所有連線、裝置、帳號配對)
func getStringByInfoPointerArray(infoPointerArray []*Info) (results string) {

	for i, infoPointer := range infoPointerArray {
		if nil != infoPointer {

			results += `第` + strconv.Itoa(i+1) + `組裝置與帳號:[`

			devicePointer := infoPointer.DevicePointer
			if nil != devicePointer {
				//str := fmt.Sprint(devicePointer)
				stringArea := fmt.Sprint(devicePointer.Area)
				stringAreaName := fmt.Sprint(devicePointer.AreaName)

				results += `裝置{` +
					`裝置ID=` + devicePointer.DeviceID +
					`,裝置Brand=` + devicePointer.DeviceBrand +
					`,裝置OnlineStatus=` + strconv.Itoa(devicePointer.OnlineStatus) +
					`,裝置DeviceStatus=` + strconv.Itoa(devicePointer.DeviceStatus) +
					`,裝置場域代號=` + stringArea +
					`,裝置場域名稱=` + stringAreaName +
					`}`
			} else {
				results += `裝置{找不到此裝置}`
			}
			accountPointer := infoPointer.AccountPointer
			if nil != accountPointer {
				//str := fmt.Sprint(accountPointer)
				results += `,帳號{` +
					`,帳號userID=` + accountPointer.UserID +
					`,帳號是否為專家=` + strconv.Itoa(accountPointer.IsExpert) +
					`,帳號是否為一線人員=` + strconv.Itoa(accountPointer.IsFrontline) +
					`}`
			} else {
				results += `,帳號{尚未登入}`
			}

			results += `]`

		}
	}
	results += `以上為所有結果。`

	return
}

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

				whatKindCommandString := `伺服器-開始偵測連線逾時`

				fmt.Printf("【此連線開始偵測逾時】連線=%v \n", clientPointer)
				logger.Infof("【此連線開始偵測逾時】連線=%v", clientPointer)

				for {

					// 偵測連線自動離線 直接結束此偵測逾時之執行序
					if infoPointer, ok := clientInfoMap[clientPointer]; ok {
						devicePointer := infoPointer.DevicePointer
						if devicePointer != nil {

							// 若裝置為離線，就認為是<登出>狀態，就不再偵測逾時。
							if devicePointer.OnlineStatus == 2 {

								details := `-已登入,但裝置已離線,離開連線逾時之偵測`

								// 一般logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, Command{}, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								break // 跳出
							}
						}
					}

					commandTime := <-commandTimeChannel                                  // 當有接收到指令，則會有值在此通道
					<-time.After(commandTime.Add(time.Second * timeout).Sub(time.Now())) // 若超過時間，則往下進行
					if 0 == len(commandTimeChannel) {                                    // 若通道裡面沒有值，表示沒有收到新指令過來，則斷線

						details := `-此裝置發生逾時,即將斷線`

						// Response:通知連線即將斷線
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, CommandNumberOfLogout, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, ""))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, Command{}, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 設定裝置在線狀態=離線
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {
							devicePointer := infoPointer.DevicePointer
							if nil != devicePointer {

								_, message := setDevicePointerOffline(devicePointer)
								details += `-設置裝置為離線狀態` + message

								// devicePointer.OnlineStatus = 2 // 離線
								// devicePointer.DeviceStatus = 0 // 重設
								// devicePointer.CameraStatus = 0 // 重設
								// devicePointer.MicStatus = 0    // 重設
								// devicePointer.RoomID = 0       // 重設
								// devicePointer.Pic = ""         //重設
							}
						}

						// 若連線存在
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {

							// 檢查裝置非NIL：進行包裝array + 廣播
							devicePointer := infoPointer.DevicePointer
							if nil != devicePointer {

								// 準備廣播:包成Array:放入 Response Devices
								deviceArray := getArrayPointer(devicePointer) // 包成array

								// (場域)廣播：狀態改變
								messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, Command{}, clientPointer, deviceArray, details)
								// broadcastByArea(tempArea, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, whatKindCommandString, Command{}, clientPointer, details) // 排除個人進行廣播

								details += `-(場域)廣播,此連線已逾時,此裝置狀態已變更為:離線,詳細訊息:` + messages
								// 一般logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, Command{}, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							}
						}

						// 移除連線
						delete(clientInfoMap, clientPointer) //刪除
						disconnectHub(clientPointer)         //斷線

						details += `-已斷線`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, Command{}, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					}

				}
			}()

			for { // 循環讀取連線傳來的資料

				whatKindCommandString := `不斷從客戶端讀資料`

				inputWebsocketData, isSuccess := getInputWebsocketDataFromConnection(connectionPointer) // 不斷從客戶端讀資料

				if !isSuccess { //若不成功 (判斷為Socket斷線)

					details := `-從客戶讀取資料失敗,即將斷線`

					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, Command{}, clientPointer) //所有值複製一份做logger
					processLoggerInfof(whatKindCommandString, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					disconnectHub(clientPointer) //斷線

					details += `-已斷線`
					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, Command{}, clientPointer) //所有值複製一份做logger
					processLoggerInfof(whatKindCommandString, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					return // 回傳:離開for迴圈(keep Reading)

				} else {

					details := `-從客戶讀取資料成功`

					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, Command{}, clientPointer) //所有值複製一份做logger
					processLoggerInfof(whatKindCommandString, details, Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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
						// 執行成功
						details := `-指令已成功解譯為json格式`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					} else {
						details := `-指令解譯失敗,json格式錯誤`

						// 警告logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					}

					// 檢查command欄位是否都給了
					fields := []string{"command", "commandType", "transactionID"}
					ok, missFields := checkCommandFields(command, fields)
					if !ok {

						whatKindCommandString := `收到指令檢查欄位`

						m := strings.Join(missFields, ",")

						// Socket Response
						if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "失敗:欄位不齊全。" + m, TransactionID: command.TransactionID}); err == nil {

							// 失敗:欄位不完全
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

							details := `-執行失敗,欄位不完全`

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// break
						} else {
							// 錯誤logger
							details := `-執行失敗,欄位不完全,但JSON轉換錯誤`
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// break
						}
						//return
					}

					// 判斷指令
					switch c := command.Command; c {

					case 1: // 登入(+廣播改變狀態)

						whatKindCommandString := `登入`

						// 檢查<帳號驗證功能>欄位是否齊全
						if !checkFieldsCompletedAndResponseIfFail([]string{"userID", "userPassword", "deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備驗證密碼:拿ID+密碼去資料庫比對密碼，若正確則進行登入

						check := false
						accountPointer := &Account{}

						// 若為demo模式，且為demo帳號，不驗證密碼，直接成功
						// fmt.Println("測試中A！expertdemoMode＝", expertdemoMode)
						// if (1 == expertdemoMode) &&
						// 	("expertA@leapsyworld.com" == command.UserID ||
						// 		"expertB@leapsyworld.com" == command.UserID ||
						// 		"expertAB@leapsyworld.com" == command.UserID) {
						// 	check = true
						// 	// 取帳號指標
						// 	accountPointer = getAccountByUserID(command.UserID)

						// } else {
						// 若為一般帳號，進行密碼驗證

						// demo 帳號完全可以登入
						check, accountPointer = checkPassword(command.UserID, command.UserPassword)

						// }

						// 驗證密碼成功:
						if check {

							details += `-驗證密碼成功`

							if accountPointer != nil {

								// 看專家帳號的驗證碼是否過期（資料庫有此帳號,且為專家帳號,才會用驗證信）
								// 僅檢查專家帳號，一線人員帳號，不會用驗證碼
								if 1 == accountPointer.IsExpert {

									details += `-此為專家帳號`

									// 看驗證碼是否過期
									m, _ := time.ParseDuration("10m")                      // 驗證碼有效時間
									deadline := accountPointer.verificationCodeTime.Add(m) // 此帳號驗證碼最後有效期限期限
									isBefore := time.Now().Before(deadline)                // 看是否還在期限內

									fmt.Println("還在期限內？", isBefore)
									if !isBefore {
										// 已過期

										details += `-驗證碼已過期,驗證失敗`

										// Response：失敗
										jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
										clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

										// 警告logger
										myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
										processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
										break // 跳出
									}
								}
							} else {

								details += `-找不到帳號`

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, "資料庫找不到此帳號", command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出
							}

							// 取得裝置Pointer
							devicePointer := getDevice(command.DeviceID, command.DeviceBrand)

							// 資料庫找不到此裝置
							if devicePointer == nil {

								details += `-找不裝置`

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出
							}

							// 進行裝置、帳號登入 (加入Map。包含處理裝置重複登入)
							if success, otherMessage := processLoginWithDuplicate(whatKindCommandString, clientPointer, command, devicePointer, accountPointer); success {
								// 成功

								details += `-登入成功,回應客戶端`

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 一般logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								deviceArray := getArrayPointer(devicePointer) // 包成array
								messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// 一般logger
								details += `-執行廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								// 失敗

								details += `-登入失敗`

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details+otherMessage, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details+otherMessage, command, clientPointer) //所有值複製一份做logger
								processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出
							}

						} else {
							// logger:帳密錯誤

							details += `-驗證密碼失敗-無此帳號或密碼錯誤`

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
							break // 跳出
						}

					case 15: // 判斷帳號是否存在，若存在則寄出驗證信

						whatKindCommandString := `判斷帳號是否存在，若存在寄出驗證信`

						// 該有欄位外層已判斷

						// 不用登入就可使用

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 是否有此帳號
						haveAccount, accountPointer := checkAccountExist(command.UserID)

						if haveAccount && (nil != accountPointer) {

							details += `-找到帳號,userID=` + accountPointer.UserID

							var success bool
							var otherMessages string

							// 準備寄送寄信

							fmt.Println("測試中15！expertdemoMode＝", expertdemoMode)

							// 若為demo模式,且為demo帳號，不驗證、不寄信、直接成功
							if (1 == expertdemoMode) &&
								("expertA@leapsyworld.com" == command.UserID || "expertB@leapsyworld.com" == command.UserID || "expertAB@leapsyworld.com" == command.UserID) {
								success = true
								otherMessages = ""

							} else {
								// 若為一般帳號，進行驗證並寄信
								success, otherMessages = processSendVerificationCodeMail(accountPointer, whatKindCommandString, details, command, clientPointer)
							}

							if success {

								details += `-驗證信已寄出`

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 一般logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {

								details += `-驗證信寄出失敗,訊息:` + otherMessages

								// Response:失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							}

						} else {
							details += `-找不到此帳號`

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						}

					case 17: // QRcode登入 (+廣播)

						whatKindCommandString := `QRcode登入`

						// 檢查<帳號驗證功能>欄位是否齊全
						if !checkFieldsCompletedAndResponseIfFail([]string{"userID", "deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// QRcode登入不需要密碼，只要確認是否有此帳號

						// // 準備進行加密 封裝資料
						// userIDToken := jwts.TokenInfo{
						// 	Data: "frontLine@leapsyworld.com",
						// }

						// // 加密結果
						// encryptionString := jwts.CreateToken(&userIDToken)
						// fmt.Println("加密後<", *encryptionString, ">前面是加密結果")
						// //fmt.Println("加密後<", encryptionString, ">前面是加密結果")

						//myString := "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJEYXRhIjoiZnJvbnRMaW5lQGxlYXBzeXdvcmxkLmNvbSJ9.WT84ZLamrfc8tZQRrhysIoglIpWgf3A69-tM-TjzH_jsp-5EQ5K4B5WXHxA_-7D9I8sCxyknr-AujxSrzgOS_g"

						// 進行解密
						token := jwts.ParseToken(command.UserID)
						//token := jwts.ParseToken(*stringPointer)

						fmt.Println("解密後token：", token)

						// 解密後字串
						var decryptedString string

						//若非nil 取出
						if token != nil {
							// 解密成功
							details += `-QR cdoe 解密成功`

							decryptedString = token.Data
						} else {
							// 解密出現錯誤，找不到token
							details += `-QR code 解密錯誤:解密找不到token`

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 錯誤logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出

						}

						fmt.Println("-解密後字串decryptedString:", decryptedString)

						userid := decryptedString
						fmt.Println("-解密後取出userid：", userid)

						// 是否有此帳號
						check, accountPointer := checkAccountExist(userid)

						// 找到帳號
						if check {

							details += `-找到帳號`

							// 找裝置Pointer
							devicePointer := getDevice(command.DeviceID, command.DeviceBrand)

							// 找到裝置
							if devicePointer != nil {

								details += `-找到裝置`

								// 登入:裝置、帳號 (加入Map。包含處理裝置重複登入)
								if success, otherMeessage := processLoginWithDuplicate(whatKindCommandString, clientPointer, command, devicePointer, accountPointer); success {
									// 登入成功

									details += `-登入成功`

									// Response:成功
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// 一般logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									// 準備廣播:包成Array:放入 Response Devices
									deviceArray := getArrayPointer(devicePointer) // 包成array
									messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

									// logger
									details += `-執行(區域)廣播,詳細訊息:` + messages
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								} else {
									// 登入失敗
									details += `-登入失敗:` + otherMeessage

									// Response:失敗
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// 一般logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									break // 跳出

								}

							} else {
								// 找不到裝置
								details += `-找不到裝置`

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								break // 跳出
							}

						} else {
							// logger:帳密錯誤
							details += `-找不到帳號或密碼錯誤`

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出
						}

					case 2: // 專家帳號要取得跟專家自己同場域的所有眼鏡裝置清單

						whatKindCommandString := `專家帳號要取得跟專家自己同場域的所有眼鏡裝置清單`

						// 該有欄位外層已判斷

						// 是否已與Server建立連線
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 取得指定場域裝置清單
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {

							// 檢查裝置
							if nil != infoPointer.DevicePointer {

								//取得跟專家場域一樣的、眼鏡裝置、並排除自己的所有裝置info
								infosInAreasExceptMineDevice, otherMessage := getDevicesWithInfoByAreaAndDeviceTypeExeptOneDevice(infoPointer.AccountPointer.Area, 1, infoPointer.DevicePointer) //待改

								// Response:成功
								// 此處json不直接轉成string,因為有 device Array型態，轉string不好轉
								if jsonBytes, err := json.Marshal(InfosInTheSameAreaResponse{Command: 2, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID, Info: infosInAreasExceptMineDevice}); err == nil {

									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Response

									infoPointerString := getStringByInfoPointerArray(infosInAreasExceptMineDevice) // 將查詢結果轉成字串

									// 一般logger
									details += `-指令執行成功,取得清單為:` + infoPointerString + `,執行細節:` + otherMessage
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {

									// logger
									details += `-指令執行失敗,後端json轉換出錯,執行細節:` + otherMessage
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
									break // 跳出

								}
							} else {
								// 找不到裝置
								details += `-找不到裝置`

								// Response:失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出

							}
						} else {
							// 尚未建立連線
							details += `-執行失敗，尚未建立連線`

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出
						}

					case 3: // 取得空房號

						whatKindCommandString := `取得空房號`

						// 該有欄位外層已判斷

						// 是否已登入
						if !checkLogedInAndResponseIfFail(clientPointer, command, `whatKindCommandString`) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 增加房號
						roomID = roomID + 1

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+`, "roomID":%d}`, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID, roomID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details += `-指令執行成功,取得房號為` + strconv.Itoa(roomID)
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 4: // 求助

						whatKindCommandString := `求助`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompletedAndResponseIfFail([]string{"pic", "roomID"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 檢核:房號未被取用過則失敗
						if command.RoomID > roomID {

							details += `-執行失敗:房號未被取用過`

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
							break // 跳出

						}

						// 設定Pic, RoomID, 裝置狀態
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {

							devicePointer := infoPointer.DevicePointer

							if nil != devicePointer {

								devicePointer.Pic = command.Pic       // Pic還原預設
								devicePointer.RoomID = command.RoomID // RoomID還原預設
								devicePointer.DeviceStatus = 2        // 設備狀態:閒置

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details += `-指令執行成功`
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								//deviceArray := getArray(clientInfoMap[clientPointer].DevicePointer) // 包成array
								deviceArray := getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array
								messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// logger
								details += `-執行(區域)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, ``)
								break
							}

						} else {
							// Response:失敗
							processResponseInfoNil(clientPointer, whatKindCommandString, command, ``)
							break
						}

					case 5: // 回應求助

						whatKindCommandString := `回應求助`

						// 是否已登入
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompletedAndResponseIfFail([]string{"deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備設定-求助者狀態

						// (求助者)裝置
						askerDevicePointer := getDevice(command.DeviceID, command.DeviceBrand)

						// 檢查有此裝置
						if askerDevicePointer != nil {

							details += `-找到(求助者)裝置`

							// 設定:求助者設備狀態
							askerDevicePointer.DeviceStatus = 3 // 通話中
							askerDevicePointer.CameraStatus = 1 //預設開啟相機
							askerDevicePointer.MicStatus = 1    //預設開啟麥克風

							// 準備設定-回應者設備狀態+房間(自己)

							// (回應者)info
							giverInfoPointer := clientInfoMap[clientPointer]
							if nil != giverInfoPointer {
								details += `-找到(回應者)連線Info`

								// (回應者)裝置
								giverDeivcePointer := giverInfoPointer.DevicePointer
								if nil != giverDeivcePointer {
									details += `-找到(回應者)裝置ID=` + giverDeivcePointer.DeviceID + `,(回應者)裝置Brand=` + giverDeivcePointer.DeviceBrand

									giverDeivcePointer.DeviceStatus = 3                   // 通話中
									giverDeivcePointer.CameraStatus = 1                   // 預設開啟相機
									giverDeivcePointer.MicStatus = 1                      // 預設開啟麥克風
									giverDeivcePointer.RoomID = askerDevicePointer.RoomID // 求助者roomID

									// Response：成功
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger
									details += `-指令執行成功` +
										`,(回應者)房號=` + strconv.Itoa(giverDeivcePointer.RoomID) +
										`,(回應者)裝置狀態DeviceStatus=` + strconv.Itoa(giverDeivcePointer.DeviceStatus) +
										`,(求助者)房號=` + strconv.Itoa(askerDevicePointer.RoomID) +
										`,(求助者)裝置狀態DeviceStatus=` + strconv.Itoa(askerDevicePointer.DeviceStatus) +
										`,(求助者)裝置ID=` + askerDevicePointer.DeviceID +
										`,(求助者)裝置Brand=` + askerDevicePointer.DeviceBrand
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									// 準備廣播:包成Array:放入 Response Devices
									deviceArray := getArrayPointer(giverDeivcePointer)    // 包成Array:放入回應者device
									deviceArray = append(deviceArray, askerDevicePointer) // 包成Array:放入求助者device

									// 進行區域廣播
									messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

									// logger
									details += `-執行(區域)廣播,詳細訊息:` + messages
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {
									details += `-(回應者)裝置不存在`
									processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
									break // 跳出
								}

							} else {
								details += `-(回應者)連線Info不存在`
								processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
								break // 跳出

							}

						} else {
							details += `-(求助者)裝置不存在`
							processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
							break // 跳出
						}

					case 6: // 變更	cam+mic 狀態

						whatKindCommandString := `變更攝影機+麥克風狀態`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompletedAndResponseIfFail([]string{"cameraStatus", "micStatus"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 設定攝影機、麥克風
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {

							devicePointer := infoPointer.DevicePointer // 取裝置
							if nil != devicePointer {

								details += `-找到裝置`

								devicePointer.CameraStatus = command.CameraStatus // 攝影機
								devicePointer.MicStatus = command.MicStatus       // 麥克風

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details += `-指令執行成功,變更裝置ID=` + devicePointer.DeviceID + `,裝置品牌=` + devicePointer.DeviceBrand + `,攝影機狀態改為=` + strconv.Itoa(devicePointer.CameraStatus) + `,麥克風狀態改為=` + strconv.Itoa(devicePointer.MicStatus)

								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								deviceArray := getArrayPointer(clientInfoMap[clientPointer].DevicePointer)

								messages := processBroadcastingDeviceChangeStatusInRoom(whatKindCommandString, command, clientPointer, deviceArray, details)

								// logger
								details += `-進行(房間)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								//找不到裝置
								details += `-找不到裝置`
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							// 找不到要求端連線info
							details += `-找不到要求端連線info`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 7: // 掛斷通話

						whatKindCommandString := `掛斷通話`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger:收到指令
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						infoPointer := clientInfoMap[clientPointer] // 取info

						// 找到要求端連線info
						if nil != infoPointer {
							details += `-找到要求端連線Info`

							devicePointer := infoPointer.DevicePointer   // 取device
							accountPointer := infoPointer.AccountPointer // 取account

							// 找到裝置
							if nil != devicePointer {
								details += `-找到裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								thisRoomID := devicePointer.RoomID

								// 其他同房間的連線
								//otherClients := getOtherClientsInTheSameRoom(clientPointer, thisRoomID)

								// 其他同房間裝置
								otherDevicesPointer := getOtherDevicesInTheSameRoom(thisRoomID, clientPointer)

								// 帳號非空
								if nil != accountPointer {
									details += `-找到帳號userID=` + accountPointer.UserID

									// 若自己是一線人員掛斷: 同房間都掛斷
									if 1 == accountPointer.IsFrontline {

										// 自己 離開房間
										devicePointer.DeviceStatus = 1 // 裝置閒置
										devicePointer.CameraStatus = 0 // 關閉
										devicePointer.MicStatus = 0    // 關閉
										devicePointer.RoomID = 0       // 沒有房間
										devicePointer.Pic = ""         //清空

										// 其他人 離開房間
										for _, dPointer := range otherDevicesPointer {
											if nil != dPointer {
												dPointer.DeviceStatus = 1 // 裝置閒置
												dPointer.CameraStatus = 0 // 關閉
												dPointer.MicStatus = 0    // 關閉
												dPointer.RoomID = 0       // 沒有房間
											}
										}

									} else if 1 == accountPointer.IsExpert {
										//若是自己是專家掛斷: 一線人員 變求助中

										//自己 離開房間
										devicePointer.DeviceStatus = 1 // 裝置閒置
										devicePointer.CameraStatus = 0 // 關閉
										devicePointer.MicStatus = 0    // 關閉
										devicePointer.RoomID = 0       // 沒有房間

										//一線人員 求助中
										for _, dPointer := range otherDevicesPointer {
											if nil != dPointer {
												dPointer.DeviceStatus = 2 // 求助中
												dPointer.CameraStatus = 0 // 關閉
												dPointer.MicStatus = 0    // 關閉
												// 房間不變
											}
										}
									}

									// Response:成功
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger:執行成功
									details += `-指令執行成功,從房號=` + strconv.Itoa(thisRoomID) + `中退出`
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									// 準備廣播:包成Array:放入 Response Devices
									// 要放入自己＋其他人
									deviceArray := getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array
									for _, e := range otherDevicesPointer {
										deviceArray = append(deviceArray, e)
									}

									messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

									// logger:進行廣播
									details += `-執行(區域)廣播,詳細訊息:` + messages
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {
									//找不到帳號
									details += `-找不到帳號`
									processResponseAccountNil(clientPointer, whatKindCommandString, command, details)
									break
								}

							} else {
								//找不到裝置
								details += `-找不到裝置`
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							//找不到要求端連線info
							details += `-找不到要求端連線info`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 8: // 登出

						whatKindCommandString := `登出`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間，登出就不用再重新計算了
						//commandTimeChannel <- time.Now()

						// logger:收到指令
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備設定登出者

						// 取出連線info
						infoPointer := clientInfoMap[clientPointer]
						if nil != infoPointer {
							details += `-找到要求端連線資訊Info`

							// 取出裝置
							devicePointer := infoPointer.DevicePointer
							if nil != devicePointer {
								details += `-找到裝置,裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								// 重設裝置為預設離線狀態
								_, message := setDevicePointerOffline(devicePointer)
								details += `-設置裝置為離線狀態` + message

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details += `-指令執行成功，連線已登出`
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								// deviceArray := getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array
								deviceArray := getArrayPointer(devicePointer) // 包成array

								// 進行廣播
								messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// 暫存即將斷線的資料
								// logger
								details += `-執行(區域)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 移除連線
								// 帳號包在clientInfoMap[clientPointer]裡面,會一併進行清空
								delete(clientInfoMap, clientPointer) //刪除
								disconnectHub(clientPointer)         //斷線

							} else {
								details += `-找不到裝置`
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
							}

						} else {
							details += `-找不到要求端連線Info`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
						}

					case 9: // 心跳包

						whatKindCommandString := `心跳包`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 成功:Response
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details += `-指令執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 13: // 取得自己帳號資訊

						whatKindCommandString := `取得自己帳號資訊`

						// 該有欄位外層已判斷

						// 是否已與Server建立連線
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備隱匿密碼

						// 取出連線info
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {
							details += `-找到要求端連線info`

							//取出帳號
							accountPointer := infoPointer.AccountPointer
							if nil != accountPointer {
								details += `-找到自己帳號資訊,userID=` + accountPointer.UserID

								accountNoPassword := *(accountPointer) //取帳號copy複本
								accountNoPassword.UserPassword = ""    //密碼隱藏

								details += `-指令執行成功`
								// Response:成功 (此處仍使用Marshal工具轉型，因考量有 物件Account{}形態，轉成string較為複雜。)
								if jsonBytes, err := json.Marshal(MyAccountResponse{
									Command:       command.Command,
									CommandType:   CommandTypeNumberOfAPIResponse,
									ResultCode:    ResultCodeSuccess,
									Results:       ``,
									TransactionID: command.TransactionID,
									Account:       accountNoPassword}); err == nil {

									// fmt.Printf("測試accountNoPassword=%+v", clientInfoMap[clientPointer].AccountPointer)
									// Response(場域、排除個人)
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {
									details += `-後端json轉換出錯`

									// logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								}
							} else {
								//帳戶為空
								details += `-找不到帳戶`
								processResponseAccountNil(clientPointer, whatKindCommandString, command, details)
							}

						} else {
							//info為空
							details += `-找不到要求端連線info`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
						}

					case 14: // 取得自己裝置資訊

						whatKindCommandString := `取得自己裝置資訊`

						// 該有欄位外層已判斷

						// 是否已經加到clientInfoMap中 表示已登入
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						if infoPointer, ok := clientInfoMap[clientPointer]; ok { //取info

							devicePointer := infoPointer.DevicePointer //取裝置

							if nil != devicePointer {

								details += `-找到裝置,裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								device := getDevice(devicePointer.DeviceID, devicePointer.DeviceBrand) // 取得裝置清單-實體                                                                                     // 自己的裝置

								// Response:成功 (此處仍使用Marshal工具轉型，因考量有 物件{}形態，轉成string較為複雜。)
								if jsonBytes, err := json.Marshal(MyDeviceResponse{Command: command.Command, CommandType: CommandTypeNumberOfAPIResponse, ResultCode: ResultCodeSuccess, Results: ``, TransactionID: command.TransactionID, Device: *device}); err == nil {
									//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

									// Response(場域、排除個人)
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger
									details += `-指令執行成功`
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {

									// logger
									details += `-執行指令失敗，後端json轉換出錯。`
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								}
							} else {
								// 找不到裝置
								details += `-找不到裝置`
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							// 找不到info
							details += `-找不到要求端連線`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 16: // 取得現在同場域空閒專家人數

						whatKindCommandString := `取得現在同場域空閒專家人數`

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						if infoPointer, ok := clientInfoMap[clientPointer]; ok {
							details += `-找到要求端連線`

							devicePointer := infoPointer.DevicePointer //取裝置
							if nil != devicePointer {
								details += `-找到裝置,裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								// 取得線上同場域閒置專家數
								onlinExperts := getOnlineIdleExpertsCountInArea(devicePointer.Area, whatKindCommandString, command, clientPointer)

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+`,"onlineExpertsIdle":%d}`, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID, onlinExperts))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details += `-指令執行成功,取得現在同場域空閒專家人數=` + strconv.Itoa(onlinExperts)
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								//找不到裝置
								details += `-找不到裝置`
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							//找不到info
							details += `-找不到要求端連線`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 18: // 眼鏡切換場域

						whatKindCommandString := `眼鏡切換場域`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 閒置中才可以切換場域（通話中、求助中無法切換場域）
						if !checkDeviceStatusIsIdleAndResponseIfFail(clientPointer, command, whatKindCommandString, "-檢查設備狀態為閒置才可切換場域") {
							break //跳出
						}

						// 眼鏡端才可以切換場域
						if !checkDeviceTypeIsGlassesAndResponseIfFail(clientPointer, command, whatKindCommandString, "-檢查設備為眼鏡端才可以切換場域") {
							break //跳出
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompletedAndResponseIfFail([]string{"area"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// QRCode解密

						// //  準備進行加密 封裝資料
						// userIDToken := jwts.TokenInfo{
						// 	Data: "2",
						// }

						// // 加密結果
						// encryptionString := jwts.CreateToken(&userIDToken)
						// fmt.Println("加密後<", *encryptionString, ">前面是加密結果")
						// //fmt.Println("加密後<", encryptionString, ">前面是加密結果")

						// 進行解密
						token := jwts.ParseToken(command.AreaEncryptionString)
						//token := jwts.ParseToken(*encryptionString)

						fmt.Println("-解密後token：", token)

						// 要取出的 data
						var newAreaString string

						// 取出內容 data
						if token != nil {
							newAreaString = token.Data
						}

						fmt.Println("-解密後Data:", newAreaString)

						// 新場域代碼：data轉成 場域數字代碼
						newAreaNumber, err := strconv.Atoi(newAreaString)
						if err == nil {
							// 轉換數字成功
							details += `-字符或數字轉換成功`
						} else {
							// 失敗：轉換數字失敗
							details += `-執行指令失敗-字符轉換失敗-字符或數字轉換錯誤或解密錯誤`

							fmt.Println(details)

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break //跳出case
						}

						// 查找是否有此場域代碼
						if areaName, ok := areaNumberNameMap[newAreaNumber]; ok {
							details += `-找到此場域代碼,場域代碼=` + strconv.Itoa(newAreaNumber) + `,場域名稱=` + areaName
						} else {
							//失敗：沒有此場域區域
							details += `-執行指令失敗-找不到此場域代碼與其對應名稱,場域代碼=` + strconv.Itoa(newAreaNumber)

							fmt.Println(details)

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break //跳出case
						}

						// 暫存
						var oldArea []int        //舊場域代碼
						var oldAreaName []string //舊場域名

						var newAreaNumberArray []int                                   //新場域代碼
						newAreaNumberArray = append(newAreaNumberArray, newAreaNumber) //封裝成array

						var newAreaNameArray []string                                                 //新場域名
						newAreaNameArray = append(newAreaNameArray, areaNumberNameMap[newAreaNumber]) //封裝成array

						// 檢查Info
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {
							details += `-找到要求端連線`

							// 檢查裝置
							devicePointer := infoPointer.DevicePointer
							if devicePointer != nil {
								details += `-找到裝置,裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								// 若場域代碼與現在場域不相同
								if newAreaNumber != devicePointer.Area[0] {
									// 成功

									oldArea = devicePointer.Area              //暫存舊場域
									oldAreaName = devicePointer.AreaName      //暫存舊場域名
									devicePointer.Area = newAreaNumberArray   //換成新場域代號
									devicePointer.AreaName = newAreaNameArray //換成新場域名

									// Response:成功
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger
									// int [] string[] 轉換成string
									newAreaString := fmt.Sprint(devicePointer.Area)
									newAreaNameString := fmt.Sprint(devicePointer.AreaName)
									oldAreaString := fmt.Sprint(oldArea)
									oldAreaNameString := fmt.Sprint(oldAreaName)

									details += `-指令執行成功,已換成新場域,新場域代號=` + newAreaString + `,新場域名=` + newAreaNameString + `,舊場域代號=` + oldAreaString + `,舊場域名=` + oldAreaNameString
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									// 準備廣播:包成Array:放入 Response Devices
									//deviceArray := getArray(clientInfoMap[clientPointer].DevicePointer) // 包成array
									deviceArray := getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array

									// 廣播給舊場域的
									processBroadcastingDeviceChangeStatusInSomeArea(whatKindCommandString, command, clientPointer, deviceArray, oldArea, details)

									// 廣播給現在新場域的連線裝置
									messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

									// logger
									details += `-執行(區域)廣播,詳細訊息:` + messages
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {
									//失敗：此裝置已經在這個場域，不進行切換
									details += `-指令執行失敗,此裝置已經在這個場域(` + areaNumberNameMap[newAreaNumber] + `)，不進行切換`
									fmt.Println(details)

									// Response：失敗
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
									break
								}
							} else {
								//找不到裝置
								details += `-找不到裝置`
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}
						} else {
							//找不到Info
							details += `-找不到要求端連線`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 19: // 取消求助

						whatKindCommandString := `取消求助`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 設定Pic, RoomID, 裝置狀態
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {
							details += `-找到要求端連線`

							devicePointer := infoPointer.DevicePointer
							if devicePointer != nil {
								//成功
								details += `-找到裝置,裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								devicePointer.Pic = ""         // Pic還原預設
								devicePointer.RoomID = 0       // RoomID還原預設
								devicePointer.DeviceStatus = 1 // 設備狀態:閒置

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details += `-指令執行成功,取消了求助,裝置RoomID=` + strconv.Itoa(devicePointer.RoomID) + `,設備狀態DeviceStatus=` + strconv.Itoa(devicePointer.DeviceStatus)
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								//deviceArray := getArray(clientInfoMap[clientPointer].DevicePointer) // 包成array
								deviceArray := getArrayPointer(devicePointer) // 包成array
								messages := processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// logger
								details += `-執行(區域)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								// 裝置為空
								details += `-找不到裝置`
								processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break

							}

						} else {
							// Info 為空
							details += `-找不到要求端連線`
							processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break

						}

					case 12: // 加入房間 //未來要做多方通話再做

						// whatKindCommandString := `加入房間`

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
