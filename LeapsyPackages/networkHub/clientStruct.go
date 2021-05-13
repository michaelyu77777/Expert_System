package networkHub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
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
	UserID         string `json:"userID"`         //使用者登入帳號
	UserPassword   string `json:"userPassword"`   //使用者登入密碼
	FunctionNumber int    `json:"functionNumber"` //是否為帳號驗證功能: 只驗證帳號是否存在＋寄驗證信功能
	IDPWIsRequired int    `json:"IDPWIsRequired"` //是否為必須登入模式: 裝置需要登入才能使用其他功能

	// 裝置Info
	DeviceID    string `json:"deviceID"`    //裝置ID
	DeviceBrand string `json:"deviceBrand"` //裝置品牌(怕平板裝置的ID會重複)
	DeviceType  int    `json:"deviceType"`  //裝置類型

	Area         []int  `json:"area"`         //場域
	DeviceName   string `json:"deviceName"`   //裝置名稱
	Pic          string `json:"pic"`          //裝置截圖(求助截圖)
	OnlineStatus int    `json:"onlineStatus"` //在線狀態
	DeviceStatus int    `json:"deviceStatus"` //設備狀態
	CameraStatus int    `json:"cameraStatus"` //相機狀態
	MicStatus    int    `json:"micStatus"`    //麥克風狀態
	RoomID       int    `json:"roomID"`       //房號

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
	Device  *Device  `json:"datas"`   //使用者登入密碼
}

// 帳戶資訊
type Account struct {
	UserID         string `json:"userID"`         //使用者登入帳號
	UserPassword   string `json:"userPassword"`   //使用者登入密碼
	IDPWIsRequired int    `json:"IDPWIsRequired"` //是否為必須登入模式: 裝置需要登入才能使用其他功能
}

// 去掉Password(FOR Log)
type AccountWithoutPassword struct {
	UserID         string `json:"userID"`         //使用者登入帳號
	IDPWIsRequired int    `json:"IDPWIsRequired"` //是否為必須登入模式: 裝置需要登入才能使用其他功能
}

// 裝置資訊
type Device struct {
	DeviceID     string `json:"deviceID"`     //裝置ID
	DeviceBrand  string `json:"deviceBrand"`  //裝置品牌(怕平板裝置的ID會重複)
	DeviceType   int    `json:"deviceType"`   //裝置類型
	Area         []int  `json:"area"`         //場域
	AreaName     string `json:"areaName"`     //場域名稱
	DeviceName   string `json:"deviceName"`   //裝置名稱
	Pic          string `json:"pic"`          //裝置截圖
	OnlineStatus int    `json:"onlineStatus"` //在線狀態
	DeviceStatus int    `json:"deviceStatus"` //設備狀態
	CameraStatus int    `json:"cameraStatus"` //相機狀態
	MicStatus    int    `json:"micStatus"`    //麥克風狀態
	RoomID       int    `json:"roomID"`       //房號
}

// 客戶端資訊:(為了Logger不印出密碼)
type InfoForLogger struct {
	UserID *string `json:"userID"` //使用者登入帳號
	Device *Device `json:"datas"`  //使用者登入密碼
}

// 登入
// type LoginInfo struct {
// 	UserID        string `json:"userID"`        //使用者登入帳號
// 	UserPassword  string `json:"userPassword"`  //使用者登入密碼
// 	Device        Device `json:"datas"`         //使用者登入密碼
// 	TransactionID string `json:"transactionID"` //分辨多執行緒順序不同的封包
// }

// 登入 - Response -
type LoginResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 取得所有裝置清單 - Response -
type DevicesResponse struct {
	Command       int       `json:"command"`
	CommandType   int       `json:"commandType"`
	ResultCode    int       `json:"resultCode"`
	Results       string    `json:"results"`
	TransactionID string    `json:"transactionID"`
	Devices       []*Device `json:"datas"`
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
	Device      []Device `json:"datas"`
}

// Map-連線/登入資訊
var clientInfoMap = make(map[*client]Info)

// 所有線上裝置清單
var onlineDeviceList []*Device

// 更新:所有裝置清單
//var allDeviceList []*Device
var allDeviceList = []*Device{}

// 指令代碼
const CommandNumberOfLogout = 8
const CommandNumberOfBroadcastingInArea = 10
const CommandNumberOfBroadcastingInRoom = 11

// 指令類型代碼
const CommandTypeNumberOfAPI = 1         // 客戶端-->Server
const CommandTypeNumberOfAPIResponse = 2 // Server-->客戶端
const CommandTypeNumberOfBroadcast = 3   // Server廣播
const CommandTypeNumberOfHeartbeat = 4   // 心跳包

// 連線逾時時間:
//const timeout = 30
var timeout = time.Duration(configurations.GetConfigPositiveIntValueOrPanic(`local`, `timeout`)) // 轉成time.Duration型態，方便做時間乘法

// 房間號(總計)
var roomID = 0

// 基底:Logger 的基本 String
// 基底:Response Json 的基本 String
var baseResponseJsonString = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"}`
var baseResponseJsonStringExtend = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"` // 可延展的

// 基底:收到Json格式
var baseLoggerServerReceiveJsonString = `伺服器收到Json:%s。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`
var baseLoggerMissFieldsString = `<檢查Json欄位>:失敗-Json欄位不齊全,以下欄位不齊全:%s。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`
var baseLoggerReceiveJsonRetrunErrorString = `伺服器內部jason轉譯失敗，當轉譯(Json欄位不齊全,以下欄位不齊全:%s)時。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d` // Server轉譯json出錯

// 基底:登入時發生的狀況Response(尚未建立Account/Device資料時)
var baseLoggerWhenLoginString = `指令<%s>:%s。客戶端Command:%+v、此連線Pointer:%p、連線清單:%+v、所有裝置:%+v、線上裝置:%+v、房號已取到:%d`

// 基底:共用(成功、失敗、廣播)
var baseLoggerServerReceiveCommnad = `收到<%s>指令。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`                          // 收到指令
var baseLoggerNotLoggedInWarnString = `指令<%s>失敗:連線尚未登入。客戶端Command:%+v、此連線帳號:%+s、此連線裝置ID:%+s、此連線裝置Brand:%+s、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d` // 連線尚未登入          // 失敗:連線尚未登入
var baseLoggerNotCompletedFieldsWarnString = `指令<%s>失敗:以下欄位不齊全:%s。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`       // 失敗:欄位不齊全
var baseLoggerSuccessString = `指令<%s>成功。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`                                 // 成功
var baseLoggerInfoBroadcastInArea = `指令<%s>-(場域)廣播-狀態變更。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、,房號已取到:%d`                // 場域廣播
var baseLoggerInfoCommonMessage = `指令<%s>-%s。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、,房號已取到:%d`                           // 普通紀錄
var baseLoggerWarnReasonString = `指令<%s>失敗:%s。客戶端Command:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`                                               // 失敗:原因
var baseLoggerErrorJsonString = `指令<%s>jason轉譯出錯。客戶端Command:%+v、此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、房號已取到:%d`               // Server轉譯json出錯

// 基底:連線逾時專用
var baseLoggerInfoForTimeout = `<偵測連線逾時>%s，timeout=%d。此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、,房號已取到:%d` // 場域廣播（逾時 timeout)
var baseLoggerWarnForTimeout = `<偵測連線逾時>%s，timeout=%d。此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、,房號已取到:%d` // 主動告知client（逾時 timeout)
var baseLoggerErrorForTimeout = `<偵測連線逾時>%s，timeout=%d。此連線帳號:%+v、此連線裝置:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、線上裝置:%+v、房號已取到:%d` // Server轉譯json出錯
var baseLoggerInfoForTimeoutWithoutNilDevice = `連線已登入，並已從清單移除裝置，判定為<登出>，不再繼續偵測逾時，timeout=%d。此連線帳號:%+v、此連線Pointer:%p、所有連線清單:%+v、所有裝置清單:%+v、房號已取到:%d`

// 待補:(定時)更新DB裝置清單，好讓後台增加裝置時，也可以再依定時間內同步補上
func UpdateAllDevicesList() {
	// 待補:固定時間內
	// for {
	importAllDevicesList()
	// }
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
		AreaName:     "場域A",
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 在線
		DeviceStatus: 1,            // 閒置
		MicStatus:    1,            // 開啟
		CameraStatus: 1,            // 開啟
		RoomID:       0,            // 無房間
	}

	// 新增假資料：場域B 眼鏡Model
	modelGlassesB := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   1,        //眼鏡
		Area:         []int{2}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     "場域B",
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 在線
		DeviceStatus: 1,            // 閒置
		MicStatus:    1,            // 開啟
		CameraStatus: 1,            // 開啟
		RoomID:       0,            // 無房間
	}

	// 新增假資料：場域A 平版Model
	modelTabA := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   2,        // 平版
		Area:         []int{1}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     "場域A",
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 在線
		DeviceStatus: 1,            // 閒置
		MicStatus:    1,            // 開啟
		CameraStatus: 1,            // 開啟
		RoomID:       0,            // 無房間
	}

	// 新增假資料：場域B 平版Model
	modelTabB := Device{
		DeviceID:     "",
		DeviceBrand:  "",
		DeviceType:   2,        // 平版
		Area:         []int{2}, // 依據裝置ID+Brand，從資料庫查詢
		AreaName:     "場域B",
		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
		Pic:          "",           // <求助>時才會從客戶端得到
		OnlineStatus: 2,            // 在線
		DeviceStatus: 1,            // 閒置
		MicStatus:    1,            // 開啟
		CameraStatus: 1,            // 開啟
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

// 取得某些場域的AllDeviceList
func getAllDevicesListByAreas(areas []int) []*Device {

	result := []*Device{}

	for _, e := range allDeviceList {
		intersection := intersect.Hash(e.Area, areas) //取交集array
		// 若場域有交集則加入
		if len(intersection) > 0 {
			result = append(result, e)
		}
	}
	return result
}

// 登入邏輯重寫
func processLoginWithDuplicate(clientPointer *client, command Command, device *Device) {

	// 建立帳號
	account := Account{
		UserID:         command.UserID,
		UserPassword:   command.UserPassword,
		IDPWIsRequired: command.IDPWIsRequired,
	}

	// 建立Info
	info := Info{
		Account: &account,
		Device:  device,
	}

	// 登入步驟:
	// 四種況狀判斷：相同連線、相異連線、不同裝置、相同裝置 重複登入之處理
	if _, ok := clientInfoMap[clientPointer]; ok {
		//相同連線

		if clientInfoMap[clientPointer].Device.DeviceID == command.DeviceID && clientInfoMap[clientPointer].Device.DeviceBrand == command.DeviceBrand {
			// 裝置相同：同裝置重複登入

			// 設定info
			clientInfoMap[clientPointer] = info

			// 狀態為上線
			clientInfoMap[clientPointer].Device.OnlineStatus = 1

		} else {
			//裝置不同（現實中不會出現，只有測試才會出現）

			//舊的裝置＝離線
			clientInfoMap[clientPointer].Device.OnlineStatus = 2

			//設定新的info
			clientInfoMap[clientPointer] = info

			//新的裝置＝上線
			clientInfoMap[clientPointer].Device.OnlineStatus = 1

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
			clientInfoMap[clientPointer] = info
			fmt.Printf("找到重複的連線，從Map中刪除，將此Socket斷線。\n")

			// 狀態為上線
			clientInfoMap[clientPointer].Device.OnlineStatus = 1

		} else {
			//裝置不同：正常新增一的新裝置
			clientInfoMap[clientPointer] = info

			//裝置狀態＝線上
			clientInfoMap[clientPointer].Device.OnlineStatus = 1
		}

	}

}

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

// 待補:判斷帳密是否正確
func checkLoginPassword(id string, pw string) bool {

	if id == pw {
		return true
	} else {
		return false
	}
}

// 判斷某連線是否已經登入(透過 clientInfoMap[client] 看Info是否有值 )
func checkLogedIn(client *client, command Command, whatKindCommandString string) bool {

	logedIn := false

	if _, ok := clientInfoMap[client]; ok {
		logedIn = true
	}

	if !logedIn {

		// 尚未登入
		// 失敗:Response
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 1, `連線尚未登入`, command.TransactionID))
		client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

		allDevices := getAllDeviceByList() // 取得裝置清單-實體
		go fmt.Printf(baseLoggerNotLoggedInWarnString+"\n", whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)
		go logger.Warnf(baseLoggerNotLoggedInWarnString, whatKindCommandString, command, command.UserID, command.DeviceID, command.DeviceBrand, client, clientInfoMap, allDevices, roomID)

		return logedIn

	} else {
		// 已登入
		return logedIn
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
	clientInfoMap[newClientPointer] = Info{Account: accountNew, Device: device}
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
func getAccountWithoutPassword(account *Account) *AccountWithoutPassword {
	accountNew := &AccountWithoutPassword{
		UserID:         account.UserID,
		IDPWIsRequired: account.IDPWIsRequired,
	}
	return accountNew
}

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

		case "functionNumber":
			if command.FunctionNumber == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "IDPWIsRequired":
			if command.IDPWIsRequired == 0 {
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
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 1, `以下欄位不齊全`+m, command.TransactionID))
		client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

		allDevices := getAllDeviceByList() // 取得裝置清單-實體
		go fmt.Printf(baseLoggerNotCompletedFieldsWarnString+"\n", whatKindCommandString, m, command, clientInfoMap[client].Account.UserID, clientInfoMap[client].Device, client, clientInfoMap, allDevices, roomID)
		go logger.Warnf(baseLoggerNotCompletedFieldsWarnString, whatKindCommandString, m, command, clientInfoMap[client].Account.UserID, clientInfoMap[client].Device, client, clientInfoMap, allDevices, roomID)
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

// keepReading - 保持讀取
func (clientPointer *client) keepReading() {

	if nil != clientPointer { // 若指標不為空

		defer func() {
			disconnectHub(clientPointer) // 中斷客戶端與網路中心的連線
		}()

		connectionPointer := clientPointer.getConnectionPointer() // 連線指標

		if nil != connectionPointer { // 若指標不為空

			// connection := *connectionPointer // 客戶端的連線

			commandTimeChannel := make(chan time.Time, 1000) // 連線逾時計算之通道(時間，1個buffer)
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
						device := []Device{}

						// 若裝置存在進行包裝array + 廣播
						if element.Device != nil {

							// 包成array
							device = getArray(clientInfoMap[clientPointer].Device)

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

				inputWebsocketData, isSuccess := getInputWebsocketDataFromConnection(connectionPointer) // 從客戶端讀資料
				fmt.Printf(`從客戶端讀取資料,client基本資訊=%+v`, clientPointer)
				logger.Infof(`從客戶端讀取資料,client基本資訊=%+v`, clientPointer)

				if !isSuccess { //若不成功 (判斷為Socket斷線)

					fmt.Println(`讀取資料不成功,client基本資訊=`, clientPointer)
					logger.Infof(`從客戶端讀取資料,client基本資訊=%s`, clientPointer)

					disconnectHub(clientPointer) //斷線
					fmt.Println(`斷線,client基本資訊=`, clientPointer)
					logger.Infof(`斷線,client基本資訊=%s`, clientPointer)

					return // 回傳:離開for迴圈(keep Reading)
				}

				wsOpCode := inputWebsocketData.wsOpCode
				dataBytes := inputWebsocketData.dataBytes

				if ws.OpText == wsOpCode {

					var command Command

					//解譯成Json
					err := json.Unmarshal(dataBytes, &command)

					//json格式錯誤
					if err == nil {
						// logger:成功
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						fmt.Printf(baseLoggerServerReceiveJsonString+"\n", "成功", command, clientPointer, clientInfoMap, allDevices, roomID)
						logger.Infof(baseLoggerServerReceiveJsonString, "成功", command, clientPointer, clientInfoMap, allDevices, roomID)
					} else {
						// logger:失敗
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						fmt.Printf(baseLoggerServerReceiveJsonString+"\n", `失敗:收到json格式錯誤。`, command, clientPointer, clientInfoMap, allDevices, roomID)
						logger.Warnf(baseLoggerServerReceiveJsonString, `失敗:收到的json格式錯誤。`, command, clientPointer, clientInfoMap, allDevices, roomID)
					}

					// 檢查command欄位是否都給了
					fields := []string{"command", "commandType", "transactionID"}
					ok, missFields := checkCommandFields(command, fields)
					if !ok {

						m := strings.Join(missFields, ",")
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						fmt.Printf(baseLoggerMissFieldsString+"\n", m, command, clientPointer, clientInfoMap, allDevices, roomID)
						logger.Warnf(baseLoggerMissFieldsString, m, command, clientPointer, clientInfoMap, allDevices, roomID)

						// Socket Response
						if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "<檢查json欄位>失敗:欄位不齊全。" + m, TransactionID: command.TransactionID}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						} else {
							fmt.Printf(baseLoggerReceiveJsonRetrunErrorString+"\n", m, command, clientPointer, clientInfoMap, allDevices, roomID)
							logger.Warnf(baseLoggerReceiveJsonRetrunErrorString, m, command, clientPointer, clientInfoMap, allDevices, roomID)
						}
						//return
					}

					// 判斷指令
					switch c := command.Command; c {

					case 1: // 登入(+廣播改變狀態)

						whatKindCommandString := `登入`

						// 檢查<帳號驗證功能>欄位是否齊全
						if !checkFieldsCompleted([]string{"userID", "userPassword", "deviceID", "deviceBrand", "deviceType", "functionNumber", "IDPWIsRequired"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger:收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "收到指令", command, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "收到指令", command, clientPointer, clientInfoMap, allDevices, roomID)

						// 若功能為<帳號驗證>
						if command.FunctionNumber == 1 {

							// 待補:去資料庫找是否有此ID

							// 待補:有此帳號
							haveAccount := true

							if haveAccount {
								// 有此userID(email)

								// 待補:發送驗證碼到此人email

								// Response:驗證帳號成功並發送驗證信
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger:驗證帳號成功並發送驗證信
								allDevices = getAllDeviceByList() // 取得裝置清單-實體
								go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "", command, clientPointer, clientInfoMap, allDevices, roomID)
								go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "", command, clientPointer, clientInfoMap, allDevices, roomID)

							} else {

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "無此帳號", command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger:失敗
								allDevices := getAllDeviceByList()                                         // 所有裝置
								account := getAccountWithoutPassword(clientInfoMap[clientPointer].Account) // 去掉密碼
								device := clientInfoMap[clientPointer].Device                              // 此裝置
								onlineDevice := getAllDeviceByList()                                       // 線上裝置
								go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "無此帳號", command, account, device, clientPointer, clientInfoMap, allDevices, onlineDevice, roomID)
								go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "無此帳號", command, account, device, clientPointer, clientInfoMap, allDevices, onlineDevice, roomID)
								break // 跳出

							}

						} else if command.FunctionNumber == 2 {
							// 若功能為<登入>

							// 不需要<登入>
							if command.IDPWIsRequired == 2 {
								// 若不須登入:直接當作登入，並建立物件

								// 取得裝置Pointer
								device := getDevice(command.DeviceID, command.DeviceBrand)

								// 資料庫找不到此裝置
								if device == nil {

									// Response：失敗
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "資料庫找不到此裝置", command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger:失敗
									allDevices := getAllDeviceByList()                                         // 所有裝置
									account := getAccountWithoutPassword(clientInfoMap[clientPointer].Account) // 去掉密碼
									device := clientInfoMap[clientPointer].Device                              // 此裝置
									onlineDevice := getAllDeviceByList()                                       // 線上裝置
									go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "失敗:資料庫找不到此裝置", command, account, device, clientPointer, clientInfoMap, allDevices, onlineDevice, roomID)
									go logger.Warnf(baseLoggerWhenLoginString, whatKindCommandString, "失敗:資料庫找不到此裝置", command, account, device, clientPointer, clientInfoMap, allDevices, onlineDevice, roomID)
									break // 跳出
								}

								// 進行裝置登入+帳號登入(包含:處理裝置重複登入、裝置加入online清單)
								//if !processLogin(whatKindCommandString, clientPointer, command, device) {
								//	break //登入失敗
								//}

								processLoginWithDuplicate(clientPointer, command, device)

								fmt.Printf("【已登入】帳號=%+v、裝置=%+v。"+"\n", clientInfoMap[clientPointer].Device, clientInfoMap[clientPointer].Account)

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger:成功
								onlineDevices := getOnlineDevicesByClientInfoMap() // 取得裝置清單-實體
								allDevices := getAllDeviceByList()
								go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "登入成功", command, clientPointer, clientInfoMap, allDevices, onlineDevices, roomID)
								go logger.Infof(baseLoggerWhenLoginString, whatKindCommandString, "登入成功", command, clientPointer, clientInfoMap, allDevices, onlineDevices, roomID)

								// 準備廣播:包成Array:放入 Response Devices
								deviceArray := getArray(device) // 包成array

								// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。)
								if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

									fmt.Printf("「測試」clientInfoMap[clientPointer].Device = %+v", clientInfoMap[clientPointer].Device)
									// 廣播(場域、排除個人)
									broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

									// logger:廣播
									allDevices := getAllDeviceByList()                                         // 所有裝置
									account := getAccountWithoutPassword(clientInfoMap[clientPointer].Account) // 去掉密碼
									device := clientInfoMap[clientPointer].Device                              // 此裝置
									onlineDevices := getAllDeviceByList()                                      // 線上裝置
									go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, account, device, clientPointer, clientInfoMap, allDevices, onlineDevices, roomID)
									go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, account, device, clientPointer, clientInfoMap, allDevices, onlineDevices, roomID)

								} else {

									// logger:json轉換出錯
									accountWithoutPassword := getAccountWithoutPassword(clientInfoMap[clientPointer].Account)
									device := clientInfoMap[clientPointer].Device
									allDevices := getAllDeviceByList()                // all devices
									onlineDevices = getOnlineDevicesByClientInfoMap() // 取得裝置清單-實體
									go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, accountWithoutPassword, device, clientPointer, clientInfoMap, allDevices, onlineDevices, roomID)
									go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, accountWithoutPassword, device, clientPointer, clientInfoMap, allDevices, onlineDevices, roomID)
									break // 跳出
								}

							} else if command.IDPWIsRequired == 1 {
								// 需要<登入>

								// 待補:拿ID+驗證碼去資料庫比對驗證碼，若正確則進行登入
								check := checkLoginPassword(command.UserID, command.UserPassword)

								// 驗證成功:
								if check {

									// 待補:

									// logger:帳密正確
									// allDevices = getAllDeviceByList() // 取得裝置清單-實體
									// go fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, "帳密正確", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
									// go logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, "帳密正確", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

									// 取得裝置Pointer
									device := getDevice(command.DeviceID, command.DeviceBrand)

									// 資料庫找不到此裝置
									if device == nil {

										// Response：失敗
										jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "資料庫找不到此裝置", command.TransactionID))
										clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

										// logger:失敗

										go fmt.Printf(baseLoggerWhenLoginString+"\n", whatKindCommandString, "失敗:資料庫找不到此裝置", command, clientPointer, clientInfoMap, allDevices, roomID)
										go logger.Warnf(baseLoggerWhenLoginString, whatKindCommandString, "失敗:資料庫找不到此裝置", command, clientPointer, clientInfoMap, allDevices, roomID)
										break // 跳出
									}

									// 進行裝置、帳號登入 (加入Map。包含處理裝置重複登入)
									processLoginWithDuplicate(clientPointer, command, device)

									// Response:成功
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger:成功
									allDevices = getAllDeviceByList() // 取得裝置清單-實體
									go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
									go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

									// 準備廣播:包成Array:放入 Response Devices
									deviceArray := getArray(device) // 包成array

									// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。)
									if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

										// 廣播(場域、排除個人)
										broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

										// logger:廣播
										allDevices := getAllDeviceByList() // 取得裝置清單-實體
										go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
										go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

									} else {

										// logger:json轉換出錯
										allDevices := getAllDeviceByList() // 取得裝置清單-實體
										go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
										go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
										break // 跳出
									}

								} else {
									// logger:帳密錯誤

									// Response：失敗
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "密碼錯誤", command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger:失敗
									allDevices := getAllDeviceByList() // 取得裝置清單-實體
									go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "密碼錯誤", command, clientPointer, clientInfoMap, allDevices, roomID)
									go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "密碼錯誤", command, clientPointer, clientInfoMap, allDevices, roomID)
									break // 跳出
								}

								// 待補:驗證失敗:

							} else {
								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "無此 IDPWIsRequired 代號", command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger:失敗
								allDevices := getAllDeviceByList() // 取得裝置清單-實體
								go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "無此 IDPWIsRequired 代號", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
								go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "無此 IDPWIsRequired 代號", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
								break // 跳出
							}

						} else {
							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "無此功能", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger:失敗
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "無此功能", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "無此功能", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出
						}

					case 2: // 取得所有裝置清單

						whatKindCommandString := `取得裝置清單`

						// 該有欄位外層已判斷

						// 是否已登入
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger:收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// Response:成功
						// 此處json不直接轉成string,因為有 device Array型態，轉string不好轉
						if jsonBytes, err := json.Marshal(DevicesResponse{Command: 2, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID, Devices: onlineDeviceList}); err == nil {

							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Response

							// logger:成功
							allDevices = getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

							// list := printDeviceList() // 裝置清單
							// fmt.Println(`回覆指令<取得所有裝置清單>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)
							// logger.Infof(`回覆指令<取得所有裝置清單>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)

						} else {

							// logger:json出錯
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
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

						// logger:收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// 增加房號
						roomID = roomID + 1

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+", RoomID:%d}", command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID, roomID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// if jsonBytes, err := json.Marshal(RoomIDResponse{Command: 3, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID, RoomID: roomID + 1}); err == nil {

						// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

						// 	roomID = roomID + 1
						// 	go fmt.Println(`房號:`, roomID, getLoginBasicInfoString(clientPointer))
						// 	go logger.Infof(`房號:`, roomID, getLoginBasicInfoString(clientPointer))

						// 	fmt.Println(`回覆指令<取得空房號>成功`, getLoginBasicInfoString(clientPointer))
						// 	logger.Infof(`回覆指令<取得空房號>成功`, getLoginBasicInfoString(clientPointer))

						// } else {
						// 	fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
						// 	logger.Warn(`json出錯`, getLoginBasicInfoString(clientPointer))
						// }

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
						// fields := []string{"pic", "roomID"}
						// ok, missFields := checkCommandFields(command, fields)
						// if !ok {

						// 	m := strings.Join(missFields, ",")
						// 	fmt.Printf("[欄位不齊全]欄位:%s,登入基本資訊:%s\n", m, getLoginBasicInfoString(clientPointer))
						// 	logger.Warnf("[欄位不齊全]欄位:%s,登入基本資訊:%s", m, getLoginBasicInfoString(clientPointer))

						// 	if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
						// 		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						// 	} else {
						// 		fmt.Println(`json出錯`)
						// 		logger.Warn(`json出錯`)
						// 	}

						// 	break //跳出
						// }

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger:收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// fmt.Println(`收到指令<求助>,登入基本資訊:%s`, getLoginBasicInfoString(clientPointer))
						// logger.Infof(`收到指令<求助>,登入基本資訊:%s`, getLoginBasicInfoString(clientPointer))

						// 檢核:房號未被取用過
						if command.RoomID > roomID {

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "房號未被取用過", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger:失敗
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "房號未被取用過", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "房號未被取用過", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出

							// if jsonBytes, err := json.Marshal(HelpResponse{Command: 4, CommandType: 2, ResultCode: 1, Results: `房號未被取用過`, TransactionID: command.TransactionID}); err == nil {
							// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							// 	fmt.Println(`回覆指令<求助>失敗-房號未被取用過,登入基本資訊:`, getLoginBasicInfoString(clientPointer))
							// 	logger.Warnf(`回覆指令<求助>失敗-房號未被取用過,登入基本資訊:%s`, getLoginBasicInfoString(clientPointer))
							// } else {
							// 	fmt.Println(`json出錯`)
							// 	logger.Warn(`json出錯`)
							// }

							// break //跳出
						}

						// 設定Pic, RoomID
						element := clientInfoMap[clientPointer]
						element.Device.Pic = command.Pic       // Pic
						element.Device.DeviceStatus = 2        // 設備狀態:求助中
						element.Device.RoomID = command.RoomID // RoomID
						clientInfoMap[clientPointer] = element // 回存Map

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// if jsonBytes, err := json.Marshal(HelpResponse{Command: 4, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
						// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						// 	fmt.Println(`回覆指令<求助>成功`, getLoginBasicInfoString(clientPointer))
						// 	logger.Infof(`回覆指令<求助>成功`, getLoginBasicInfoString(clientPointer))
						// } else {
						// 	fmt.Println(`json出錯`)
						// 	logger.Warn(`json出錯`)
						// 	//return
						// 	break
						// }

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array

						// (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

							// 區域廣播(排除個人)
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

							// logger:區域廣播
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						} else {

							// logger:json出錯
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出
						}

					case 5: // 回應求助

						whatKindCommandString := `回應求助`

						//TransactionID 外層已經檢查過

						// 是否已登入
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompleted([]string{"roomID", "deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// fields := []string{"roomID", "deviceID", "deviceBrand"}
						// ok, missFields := checkCommandFields(command, fields)
						// if !ok {

						// 	m := strings.Join(missFields, ",")
						// 	fmt.Println("[欄位不齊全]", m)
						// 	logger.Warn(`[欄位不齊全]`, m)
						// 	if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
						// 		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						// 	} else {
						// 		fmt.Println(`json出錯`)
						// 		logger.Errorf(`json出錯`)
						// 		return
						// 	}
						// 	//return
						// 	break
						// }

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger：收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// 設定回應者的設備狀態+房間(自己)
						element := clientInfoMap[clientPointer]
						element.Device.DeviceStatus = 3        // 通話中
						element.Device.RoomID = command.RoomID // 想回應的RoomID
						clientInfoMap[clientPointer] = element // 存回Map

						// 設定求助者的狀態
						devicePointer := getDevice(command.DeviceID, command.DeviceBrand)

						// 檢查有此裝置，且房間號一樣
						if devicePointer == nil {

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "求助者不存在", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger:失敗
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "求助者不存在", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "求助者不存在", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出

							// if jsonBytes, err := json.Marshal(AnswerResponse{Command: 5, CommandType: 2, ResultCode: 1, Results: `求助者不存在`, TransactionID: command.TransactionID}); err == nil {

							// 	// Response：失敗-求助者不存在
							// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}
							// 	// fmt.Println(`回覆指令<回應求助>失敗-求助者不存在,登入基本資訊:`, getLoginBasicInfoString(clientPointer))
							// 	// logger.Warnf(`回覆指令<回應求助>失敗-求助者不存在,登入基本資訊:%s`, getLoginBasicInfoString(clientPointer))
							// } else {
							// 	fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
							// 	logger.Errorf(`json出錯,登入基本資訊 %s`, getLoginBasicInfoString(clientPointer))

							// }
							// break // 跳出

						} else if devicePointer.RoomID != command.RoomID {

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "房號錯誤", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger:失敗
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "房號錯誤", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "房號錯誤", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出

							// // 房號錯誤
							// // Response:失敗
							// if jsonBytes, err := json.Marshal(AnswerResponse{Command: 5, CommandType: 2, ResultCode: 1, Results: `房號錯誤`, TransactionID: command.TransactionID}); err == nil {
							// 	//Response
							// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}
							// 	fmt.Println(`回覆指令<回應求助>失敗-房號錯誤,登入基本資訊:`, getLoginBasicInfoString(clientPointer))
							// 	logger.Warnf(`回覆指令<回應求助>失敗-房號錯誤,登入基本資訊:%s`, getLoginBasicInfoString(clientPointer))
							// } else {
							// 	fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
							// 	logger.Errorf(`json出錯,登入基本資訊 %s`, getLoginBasicInfoString(clientPointer))

							// }
							// break // 跳出

						} else {
							// 正常
							devicePointer.DeviceStatus = 3 // 通話中
						}

						// fmt.Println(`回應者:`, getLoginBasicInfoString(clientPointer), ` 求助者:DeviceID:`, command.DeviceID, `,DeviceBrand:`, command.DeviceBrand)
						// logger.Infof(`回應者:`, getLoginBasicInfoString(clientPointer), ` 求助者:DeviceID:`, command.DeviceID, `,DeviceBrand:`, command.DeviceBrand)

						// Response：成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// if jsonBytes, err := json.Marshal(AnswerResponse{Command: 5, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
						// 	//Response
						// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}
						// 	fmt.Println(`回覆指令<回應求助>成功`, getLoginBasicInfoString(clientPointer))
						// 	logger.Infof(`回覆指令<回應求助>成功`, getLoginBasicInfoString(clientPointer))
						// } else {
						// 	fmt.Println(`json出錯`)
						// 	logger.Errorf(`json出錯`)
						// 	//return
						// 	break
						// }

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成Array:放入回應者device
						deviceArray = append(deviceArray, *devicePointer)            // 包成Array:放入求助者device

						// (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

							// 區域廣播(排除個人)
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播
							go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						} else {

							// json出錯
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出
						}

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
						// fields := []string{"cameraStatus", "micStatus"}
						// ok, missFields := checkCommandFields(command, fields)
						// if !ok {
						// 	m := strings.Join(missFields, ",")
						// 	fmt.Println("[欄位不齊全]", m)
						// 	if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
						// 		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						// 	} else {
						// 		fmt.Println(`json出錯`)
						// 		logger.Warn(`json出錯`)
						// 	}
						// 	//return
						// 	break
						// }

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger:收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// fmt.Println(`收到指令<變更攝影機+麥克風狀態>`, getLoginBasicInfoString(clientPointer))
						// logger.Infof(`收到指令<變更攝影機+麥克風狀態>`, getLoginBasicInfoString(clientPointer))

						// 設定攝影機、麥克風
						element := clientInfoMap[clientPointer]            // 取出device
						element.Device.CameraStatus = command.CameraStatus // 攝影機
						element.Device.MicStatus = command.MicStatus       // 麥克風
						clientInfoMap[clientPointer] = element             // 回存Map

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// if jsonBytes, err := json.Marshal(CamMicResponse{Command: 6, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
						// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						// } else {
						// 	fmt.Println(`json出錯`)
						// 	logger.Warn(`json出錯`)
						// 	//return
						// 	break
						// }

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArray(clientInfoMap[clientPointer].Device)

						// (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInRoom, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

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
							break // 跳出
						}

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
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						// fmt.Println(`收到指令<掛斷通話>`, getLoginBasicInfoString(clientPointer))
						// logger.Infof(`收到指令<掛斷通話>`, getLoginBasicInfoString(clientPointer))

						// 設定掛斷者cam, mic
						element := clientInfoMap[clientPointer] // 取出device
						element.Device.DeviceStatus = 1         // 閒置
						element.Device.RoomID = 0               // 沒有房間
						clientInfoMap[clientPointer] = element  // 回存Map

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// if jsonBytes, err := json.Marshal(RingOffResponse{Command: 7, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
						// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						// 	fmt.Println(`回覆指令<掛斷通話>成功`, getLoginBasicInfoString(clientPointer))
						// 	logger.Infof(`回覆指令<掛斷通話>成功`, getLoginBasicInfoString(clientPointer))
						// } else {
						// 	fmt.Println(`json出錯`)
						// 	logger.Warn(`json出錯`)
						// 	//return
						// 	break
						// }

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array

						// 此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

							// 場域廣播(排除個人):裝置狀態變成<閒置>
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						}

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
						// logger:登出與逾時：logger、fmt都不使用平行處理（因為會涉及刪除連線與裝置，可能列印會碰到nullpointer問題）
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// 設定登出者
						element := clientInfoMap[clientPointer] // 取出device
						element.Device.OnlineStatus = 2         // 離線
						clientInfoMap[clientPointer] = element  // 回存Map

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						// logger:登出與逾時：logger、fmt都不使用平行處理（因為會涉及刪除連線與裝置，可能列印會碰到nullpointer問題）
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// if jsonBytes, err := json.Marshal(LogoutResponse{Command: 8, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
						// 	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

						// 	list := printDeviceList() // 裝置清單
						// 	fmt.Println(`回覆指令<登出>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)
						// 	logger.Infof(`回覆指令<登出>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)
						// } else {
						// 	fmt.Println(`json出錯`)
						// 	logger.Warn(`json出錯`)
						// 	//return
						// 	break
						// }

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array

						// 此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

							// 場域廣播(排除個人)
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

							// logger:登出與逾時：logger、fmt都不使用平行處理（因為會涉及刪除連線與裝置，可能列印會碰到nullpointer問題）
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						} else {

							// logger:json出錯
							// logger:登出與逾時：logger、fmt都不使用平行處理（因為會涉及刪除連線與裝置，可能列印會碰到nullpointer問題）
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出
						}

						// 暫存即將斷線的資料
						tempClientPointer := &clientPointer
						tempClientUserID := clientInfoMap[clientPointer].Account.UserID
						tempClientDevice := clientInfoMap[clientPointer].Device

						// 移除連線
						delete(clientInfoMap, clientPointer) //刪除
						disconnectHub(clientPointer)         //斷線

						// 從清單中移除裝置
						onlineDeviceList = removeDeviceFromListByDevice(onlineDeviceList, tempClientDevice)

						// logger:指令完成
						// logger:登出與逾時：logger、fmt都不使用平行處理（因為會涉及刪除連線與裝置，可能列印會碰到nullpointer問題）
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						details := `此連線已登出(刪除連線與從裝置清單中移除)`
						fmt.Println(baseLoggerInfoCommonMessage+"\n", whatKindCommandString, details, command, tempClientUserID, tempClientDevice, tempClientPointer, clientInfoMap, allDevices, roomID)
						logger.Infof(baseLoggerInfoCommonMessage, whatKindCommandString, details, command, tempClientUserID, tempClientDevice, tempClientPointer, clientInfoMap, allDevices, roomID)

					case 9: // 心跳包

						whatKindCommandString := `心跳包`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// 成功:Response
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, 2, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

					case 12: // 加入房間

						whatKindCommandString := `加入房間`

						// 是否已登入(基本欄位，外層已經檢查過)
						if !checkLogedIn(clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 檢查欄位是否齊全
						if !checkFieldsCompleted([]string{"roomID"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger：收到指令
						allDevices := getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Println(baseLoggerServerReceiveCommnad+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerServerReceiveCommnad, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// 檢核:房號未被取用過
						if command.RoomID > roomID {

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 1, "房號未被取用過", command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger:失敗
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerWarnReasonString+"\n", whatKindCommandString, "房號未被取用過", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Warnf(baseLoggerWarnReasonString, whatKindCommandString, "房號未被取用過", command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出

						}

						// 設定加入設備的狀態+房間(自己)
						element := clientInfoMap[clientPointer]
						element.Device.DeviceStatus = 3        // 通話中
						element.Device.RoomID = command.RoomID // 想回應的RoomID
						clientInfoMap[clientPointer] = element // 存回Map

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, 0, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger:成功
						allDevices = getAllDeviceByList() // 取得裝置清單-實體
						go fmt.Printf(baseLoggerSuccessString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
						go logger.Infof(baseLoggerSuccessString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						// 準備廣播:包成Array:放入 Response Devices
						deviceArray := getArray(clientInfoMap[clientPointer].Device)

						// (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, Device: deviceArray}); err == nil {

							// 區域廣播(排除個人)
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer)

							// logger:區域廣播
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerInfoBroadcastInArea+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Infof(baseLoggerInfoBroadcastInArea, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)

						} else {

							// logger:json出錯
							allDevices := getAllDeviceByList() // 取得裝置清單-實體
							go fmt.Printf(baseLoggerErrorJsonString+"\n", whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							go logger.Errorf(baseLoggerErrorJsonString, whatKindCommandString, command, clientInfoMap[clientPointer].Account.UserID, clientInfoMap[clientPointer].Device, clientPointer, clientInfoMap, allDevices, roomID)
							break // 跳出
						}

					}

					// }

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
