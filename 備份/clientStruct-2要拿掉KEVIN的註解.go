package networkHub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
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

	// 裝置Info
	DeviceID    string `json:"deviceID"`    //裝置ID
	DeviceBrand string `json:"deviceBrand"` //裝置品牌(怕平板裝置的ID會重複)
	DeviceType  int    `json:"deviceType"`  //裝置類型

	Area         string `json:"area"`         //場域
	DeviceName   string `json:"deviceName"`   //裝置名稱
	Pic          string `json:"pic"`          //裝置截圖(求助截圖)
	OnlineStatus int    `json:"onlineStatus"` //在線狀態
	DeviceStatus int    `json:"deviceStatus"` //設備狀態
	CameraStatus int    `json:"cameraStatus"` //相機狀態
	MicStatus    int    `json:"micStatus"`    //麥克風狀態
	RoomID       int    `json:"roomID"`       //房號

	//測試用
	Argument1 string `json:"argument1"`
	Argument2 string `json:"argument2"`
}

// 心跳包 Response
type Heartbeat struct {
	Command     int `json:"command"`
	CommandType int `json:"commandType"`
}

// 客戶端資訊
type Info struct {
	UserID       string  `json:"userID"`       //使用者登入帳號
	UserPassword string  `json:"userPassword"` //使用者登入密碼
	Device       *Device `json:"datas"`        //使用者登入密碼

	//測試用
	//Name   string `json:"name"`
	//Status int    `json:"status"`
	// UserID        string `json:"userID"`        //使用者登入帳號
	// UserPassword  string `json:"userPassword"`  //使用者登入密碼
	// DeviceID      string `json:"deviceID"`      //裝置ID
	// DeviceBrand   string `json:"deviceBrand"`   //裝置品牌(怕平板裝置的ID會重複)
	// DeviceType    int    `json:"deviceType"`    //裝置類型
	// TransactionID string `json:"transactionID"` //分辨多執行緒順序不同的封包
}

// 登入
type LoginInfo struct {
	UserID        string `json:"userID"`        //使用者登入帳號
	UserPassword  string `json:"userPassword"`  //使用者登入密碼
	Device        Device `json:"datas"`         //使用者登入密碼
	TransactionID string `json:"transactionID"` //分辨多執行緒順序不同的封包
}

// 裝置資訊
type Device struct {
	DeviceID     string `json:"deviceID"`     //裝置ID
	DeviceBrand  string `json:"deviceBrand"`  //裝置品牌(怕平板裝置的ID會重複)
	DeviceType   int    `json:"deviceType"`   //裝置類型
	Area         []int  `json:"area"`         //場域
	DeviceName   string `json:"deviceName"`   //裝置名稱
	Pic          string `json:"pic"`          //裝置截圖
	OnlineStatus int    `json:"onlineStatus"` //在線狀態
	DeviceStatus int    `json:"deviceStatus"` //設備狀態
	CameraStatus int    `json:"cameraStatus"` //相機狀態
	MicStatus    int    `json:"micStatus"`    //麥克風狀態
	RoomID       int    `json:"roomID"`       //房號
}

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

// 取得空房號 - Response -
type RoomIDResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
	RoomID        int    `json:"roomID"`
}

// 求助 - Response -
type HelpResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 回應求助 - Response -
type AnswerResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 改變攝影機麥克風狀態 - Response -
type CamMicResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 掛斷 - Response -
type RingOffResponse struct {
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 登出 - Response -
type LogoutResponse struct {
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
// 測試離線用
var clientInfoMap = make(map[*client]Info)

// var clientLoginInfo = make(map[*client]LoginInfo)

// 所有裝置清單
var deviceList []*Device

// 連線逾時時間:
const timeout = 10

// 房間號(總計)
var roomID = 0

// 從清單移除某裝置
func removeDevice(slice []*Device, s int) []*Device {
	return append(slice[:s], slice[s+1:]...) //回傳移除後的array
}

//取得裝置
func getDevice(deviceID string, deviceBrand string) *Device {

	var device *Device
	// 若找到則返回
	for i, _ := range deviceList {
		if deviceList[i].DeviceID == deviceID {
			if deviceList[i].DeviceBrand == deviceBrand {

				fmt.Println("移除後清單", deviceList)
				device = deviceList[i]
			}
		}
	}

	return device // 回傳
}

// 增加裝置到清單
func addDeviceToList(device *Device) bool {

	// 若裝置重複，則重新登入
	// for i, _ := range deviceList {
	// 	if deviceList[i].DeviceID == device.DeviceID {
	// 		if deviceList[i].DeviceBrand == device.DeviceBrand {

	// 			//移除舊的
	// 			deviceList = removeDevice(deviceList, i)
	// 			fmt.Println("移除後清單", deviceList)

	// 			//新增新的
	// 			deviceList = append(deviceList, device)
	// 			fmt.Println("新增後清單", deviceList)

	// 			fmt.Println("裝置重新登入")
	// 			return true // 回傳
	// 		}
	// 	}
	// }

	for i, d := range deviceList {
		if d == device {

			//移除舊的
			deviceList = removeDevice(deviceList, i)
			fmt.Println("移除後清單", deviceList)

			//新增新的
			deviceList = append(deviceList, device)
			fmt.Println("新增後清單", deviceList)

			fmt.Println("裝置重新登入")
			return true // 回傳
		}
	}

	deviceList = append(deviceList, device) //新增裝置
	fmt.Println("裝置清單:", deviceList)
	return true
}

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
			fmt.Println("找到相同Area=", intersection)
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

	for i, e := range deviceList {
		fmt.Println(` DeviceID:` + e.DeviceID + ",DeviceBrand:" + e.DeviceBrand)
		s += `[` + (string)(i) + `],DeviceID:` + e.DeviceID
		s += `[` + (string)(i) + `],DeviceBrand:` + e.DeviceBrand
	}

	return s
}

// 判斷連線是否已經登入
func checkLogedIn(c *client) bool {

	logedIn := false

	if _, ok := clientInfoMap[c]; ok {
		logedIn = true
	}

	return logedIn
}

// 取得連線資訊物件

// 取得登入基本資訊字串
func getLoginBasicInfoString(c *client) string {

	s := " UserID:" + clientInfoMap[c].UserID

	if clientInfoMap[c].Device != nil {
		s += ",DeviceID:" + clientInfoMap[c].Device.DeviceID
		s += ",DeviceBrand:" + clientInfoMap[c].Device.DeviceBrand
		s += ",DeviceName:" + clientInfoMap[c].Device.DeviceName
		s += ",DeviceType:" + (string)(clientInfoMap[c].Device.DeviceType)
		//s += ",Area:" + strings.Join(clientInfoMap[c].Device.Area, ",")
		s += ",Area:" + strings.Replace(strings.Trim(fmt.Sprint(clientInfoMap[c].Device.Area), "[]"), " ", ",", -1) //將int[]內容轉成string
		s += ",RoomID:" + (string)(clientInfoMap[c].Device.RoomID)
		s += ",OnlineStatus:" + (string)(clientInfoMap[c].Device.OnlineStatus)
		s += ",DeviceStatus:" + (string)(clientInfoMap[c].Device.DeviceStatus)
		s += ",CameraStatus:" + (string)(clientInfoMap[c].Device.CameraStatus)
		s += ",MicStatus:" + (string)(clientInfoMap[c].Device.MicStatus)
	}

	return s
}

// 測試用
// type OnlineStatus struct {
// 	Name   string `json:"name"`
// 	Status int    `json:"status"`
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

			commandTimeChannel := make(chan time.Time, 1000) // 連線逾時計算之通道(時間，1個buffer)
			go func() {
				fmt.Println(`偵測連線逾時-開始`, getLoginBasicInfoString(clientPointer))
				logger.Infof(`偵測連線逾時-開始`, getLoginBasicInfoString(clientPointer))

				for {
					commandTime := <-commandTimeChannel                                  // 當有接收到指令，則會有值在此通道
					<-time.After(commandTime.Add(time.Second * timeout).Sub(time.Now())) // 若超過時間，則往下進行
					if 0 == len(commandTimeChannel) {                                    // 若通道裡面沒有值，表示沒有收到新指令過來，則斷線

						fmt.Println(`連線逾時-發生`, clientPointer)
						logger.Infof(`連線逾時-發生`, clientPointer)

						fmt.Println(`測試A`)

						// 設定裝置在線狀態=離線
						element := clientInfoMap[clientPointer]
						if element.Device != nil {
							element.Device.OnlineStatus = 2 // 離線
						}
						clientInfoMap[clientPointer] = element // 回存

						fmt.Println(`測試B`)

						device := []Device{} // 包成array
						if element.Device != nil {
							device = getArray(clientInfoMap[clientPointer].Device) // 包成array
						}

						fmt.Println(`測試C`)

						// 【廣播】狀態變更-離線
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: 9, CommandType: 3, Device: device}); err == nil {
							// 	fmt.Println(`json`, jsonBytes)
							// 	fmt.Println(`json string`, string(jsonBytes))
							//broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播檔案內容
							//broadcastExceptOne(clientPointer, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 排除個人進行廣播
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行廣播

							fmt.Println(`【廣播】狀態變更-離線`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`【廣播】狀態變更-離線`, getLoginBasicInfoString(clientPointer))
						} else {
							fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`json出錯`, getLoginBasicInfoString(clientPointer))
						}

						// 移除連線
						//delete(clientInfoMap, clientPointer) //刪除
						disconnectHub(clientPointer) //斷線

					}
				}
			}()

			for { // 循環讀取連線傳來的資料

				inputWebsocketData, isSuccess := getInputWebsocketDataFromConnection(connectionPointer) // 從客戶端讀資料
				fmt.Println(`從客戶端讀取資料`, getLoginBasicInfoString(clientPointer))
				logger.Infof(`從客戶端讀取資料`, getLoginBasicInfoString(clientPointer))

				if !isSuccess { //若不成功 (判斷為Socket斷線)

					fmt.Println(`讀取資料不成功`, getLoginBasicInfoString(clientPointer))
					logger.Infof(`從客戶端讀取資料`, getLoginBasicInfoString(clientPointer))

					disconnectHub(clientPointer) //斷線
					fmt.Println(`斷線`, getLoginBasicInfoString(clientPointer))
					logger.Infof(`斷線`, getLoginBasicInfoString(clientPointer))

					// myInfo, myInfoOK := clientInfoMap[clientPointer] // 查看名單的key有無客戶端

					// if myInfoOK { // 如果有名為客戶端的key

					// 	// broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: []byte(fmt.Sprintf(`%s 斷線了`, myInfo.Name))}) // 廣播檔案內容

					// 	myInfo.Status = 0 //狀態設定為斷線

					// 	clientInfoMap[clientPointer] = myInfo // 重新設定名單的key客戶端的資訊

					// 	if jsonBytes, err := json.Marshal(OnlineStatus{Name: myInfo.Name, Status: 0}); err == nil {
					// 		// fmt.Println(`json`, jsonBytes)
					// 		// fmt.Println(`json string`, string(jsonBytes))
					// 		//broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播說此key離線
					// 		//broadcastExceptOne(clientPointer, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 排除個人進行廣播
					// 		broadcastByArea(clientLoginInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

					// 		fmt.Println(myInfo.Name + `離線`)
					// 	}

					// 	fmt.Println(`刪除客戶` + myInfo.Name)
					// 	delete(clientInfoMap, clientPointer) //刪除
					// 	disconnectHub(clientPointer)         //斷線

					// }

					return // 回傳:離開for迴圈(keep Reading)
				}

				wsOpCode := inputWebsocketData.wsOpCode
				dataBytes := inputWebsocketData.dataBytes

				if ws.OpText == wsOpCode {

					// dataString := string(dataBytes) // 資料字串

					// if isOriginalFileName(dataString) { // 若資料字串為原檔名
					//
					// 	clientPointer.fileExtensionMutexPointer.Lock()         // 鎖寫
					// 	clientPointer.fileExtension = filepath.Ext(dataString) // 儲存附檔名
					// 	clientPointer.fileExtensionMutexPointer.Unlock()       // 解鎖寫
					//
					// } else {
					// go broadcastHubWebsocketData(websocketData{wsOpCode: wsOpCode, dataBytes: dataBytes}) // 廣播檔案內容

					// go broadcastHubWebsocketData(websocketData{wsOpCode: wsOpCode, dataBytes: dataBytes}) // 廣播檔案內容

					var command Command

					//解譯成Json
					err := json.Unmarshal(dataBytes, &command)

					//json格式錯誤
					if err == nil {
						fmt.Println(`json格式正確。command:`, command)
					} else {
						fmt.Println(`json格式錯誤。Error Message:`, err)
					}

					//判斷指令
					// if `name` == command.Command {
					// 	clientInfoMap[clientPointer] = Info{Name: command.Argument1, Status: 1}

					// 	// myInfo, myInfoOK := clientInfoMap[clientPointer]

					// 	// if myInfoOK {
					// 	// broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: []byte(fmt.Sprintf(`%s 斷線了`, myInfo.Name))}) // 廣播檔案內容

					// 	if jsonBytes, err := json.Marshal(OnlineStatus{Name: clientInfoMap[clientPointer].Name, Status: 1}); err == nil {
					// 		// 	fmt.Println(`json`, jsonBytes)
					// 		// 	fmt.Println(`json string`, string(jsonBytes))
					// 		broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播檔案內容
					// 	}

					// 	// }

					// 	fmt.Println(`clientInfoMap`, clientInfoMap)
					// 	// go broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: []byte(command.Argument1 + `上線了`)}) // 廣播檔案內容
					// } else if `message` == command.Command {

					// 	myInfo, myInfoOK := clientInfoMap[clientPointer]

					// 	if myInfoOK {
					// 		for client, info := range clientInfoMap {

					// 			if command.Argument1 == info.Name {
					// 				client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: []byte(myInfo.Name + `說:` + command.Argument2)}
					// 				break
					// 			}

					// 		}
					// 	}

					// }

					// 檢查command欄位是否都給了
					fields := []string{"command", "commandType", "transactionID"}
					ok, missFields := checkCommandFields(command, fields)
					if !ok {

						m := strings.Join(missFields, ",")
						fmt.Println(`[欄位不齊全]`, clientPointer, m)
						logger.Warnf(`[欄位不齊全]`, clientPointer, m)

						// Socket Response
						if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						} else {
							fmt.Println(`json出錯`)
							logger.Warnf(`json出錯`, clientPointer, m)
						}
						//return
					}

					// 判斷指令
					switch c := command.Command; c {

					case 1: // 登入+廣播改變狀態

						// 檢查欄位是否齊全
						fields := []string{"userID", "userPassword", "deviceID", "deviceBrand", "deviceType"}
						ok, missFields := checkCommandFields(command, fields)
						if !ok {

							m := strings.Join(missFields, ",")
							fmt.Println("[欄位不齊全]", clientPointer, m)
							logger.Warnf(`[欄位不齊全]`, clientPointer, m)

							if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							} else {
								fmt.Println(`json出錯`, clientPointer)
								logger.Warnf(`json出錯`, clientPointer)
							}
							//return
							break
						}

						// 當送來指令，更新心跳包通道時間
						fmt.Println(`收到指令<登入>`, getLoginBasicInfoString(clientPointer))
						commandTimeChannel <- time.Now()
						logger.Infof(`收到指令<登入>`, getLoginBasicInfoString(clientPointer))

						// 判斷密碼是否正確
						userid := command.UserID
						userpassword := command.UserPassword

						resultCode := 0 //結果代碼
						results := ""   //錯誤內容

						//測試帳號 id:001 pw:test
						if userid == "001" { //從資料庫找
							check := userpassword == "test" //從資料庫找比對
							if check {

								// 帳號正確
								resultCode = 0 //正常
								fmt.Println(`密碼正確`, clientPointer)
								logger.Infof(`密碼正確`, clientPointer)
								device := Device{

									DeviceID:     command.DeviceID,
									DeviceBrand:  command.DeviceBrand,
									DeviceType:   command.DeviceType,
									Area:         []int{777},                // 依據裝置ID+Brand，從資料庫查詢
									DeviceName:   "001+" + command.DeviceID, // 依據裝置ID+Brand，從資料庫查詢
									Pic:          "",                        // <求助>時才會從客戶端得到
									OnlineStatus: 1,                         // 在線
									DeviceStatus: 1,                         // 閒置
									MicStatus:    1,                         // 開啟
									CameraStatus: 1,                         // 開啟
									RoomID:       0,                         // 無房間
								}

								// Map <client連線> <登入Info>
								clientInfoMap[clientPointer] = Info{UserID: command.UserID, UserPassword: command.UserPassword, Device: &device}

								// 裝置加入清單
								if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: resultCode, Results: results, TransactionID: command.TransactionID}); err == nil {

									//deviceList = append(deviceList, device) //新增裝置
									if addDeviceToList(clientInfoMap[clientPointer].Device) {

										// 加入成功
										list := printDeviceList() // 裝置清單
										fmt.Println(`加入裝置清單成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)
										logger.Infof(`加入裝置清單成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)

										clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
										logger.Infof(`回覆<登入>成功`, getLoginBasicInfoString(clientPointer))

										deviceArray := getArray(&device) // 包成array

										// 【廣播-場域】狀態變更
										if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: 9, CommandType: 3, Device: deviceArray}); err == nil {
											// 	fmt.Println(`json`, jsonBytes)
											// 	fmt.Println(`json string`, string(jsonBytes))

											//broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播檔案內容
											//broadcastExceptOne(clientPointer, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 排除個人進行廣播
											broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

											fmt.Println(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
											logger.Infof(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))

										} else {
											fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
											logger.Warnf(`json出錯`, getLoginBasicInfoString(clientPointer))

										}

									} else {
										fmt.Println(`加入裝置清單時錯誤`, getLoginBasicInfoString(clientPointer))
										logger.Warnf(`加入裝置清單時錯誤`, getLoginBasicInfoString(clientPointer))

										if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: 1, Results: `加入裝置清單時錯誤`, TransactionID: command.TransactionID}); err == nil {
											clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

											// 從clientInfoMap移除此連線訊息
											delete(clientInfoMap, clientPointer)
											logger.Infof(`移除此連線`, getLoginBasicInfoString(clientPointer))
											logger.Infof(`回覆<登入>失敗`, getLoginBasicInfoString(clientPointer))

										} else {
											fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
											logger.Warnf(`json出錯`, getLoginBasicInfoString(clientPointer))

										}
									}
								} else {
									fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
									logger.Warnf(`json出錯`, getLoginBasicInfoString(clientPointer))

								}

							} else {
								fmt.Println(`密碼錯誤`, getLoginBasicInfoString(clientPointer))
								logger.Warnf(`密碼錯誤`, getLoginBasicInfoString(clientPointer))

								if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: 1, Results: `密碼錯誤`, TransactionID: command.TransactionID}); err == nil {

									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
									logger.Infof(`回覆<登入>失敗`, getLoginBasicInfoString(clientPointer))

								} else {
									fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
									logger.Warnf(`json出錯`, getLoginBasicInfoString(clientPointer))

								}
								// disconnectHub(clientPointer)
							}
						} else {
							fmt.Println(`無此帳號`, getLoginBasicInfoString(clientPointer))
							logger.Warnf(`無此帳號`, getLoginBasicInfoString(clientPointer))

							if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: 1, Results: "無此帳號", TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								logger.Infof(`回覆<登入>失敗`, getLoginBasicInfoString(clientPointer))

							} else {
								fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
								logger.Warnf(`json出錯`, getLoginBasicInfoString(clientPointer))

							}
						}

					case 2: // 取得所有裝置清單

						// 該有欄位外層已判斷

						// 是否已登入
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 2, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<取得所有裝置清單>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<取得所有裝置清單>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()
						fmt.Println(`收到指令<取得所有裝置清單>`, getLoginBasicInfoString(clientPointer))
						logger.Infof(`收到指令<取得所有裝置清單>`, getLoginBasicInfoString(clientPointer))

						// Response
						if jsonBytes, err := json.Marshal(DevicesResponse{Command: 2, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID, Devices: deviceList}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							list := printDeviceList() // 裝置清單
							fmt.Println(`回覆指令<取得所有裝置清單>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)
							logger.Infof(`回覆指令<取得所有裝置清單>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)

						} else {
							fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
							logger.Warnf(`json出錯`, getLoginBasicInfoString(clientPointer))
						}

					case 3: // 取得空房號

						// 該有欄位外層已判斷

						// 是否已登入
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 3, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<取得空房號>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<取得空房號>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 基本資訊
						//basicInfo := getLoginBasicInfoString(clientPointer)

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()
						fmt.Println(`收到指令<取得空房號>`, getLoginBasicInfoString(clientPointer))
						logger.Infof(`收到指令<取得空房號>`, getLoginBasicInfoString(clientPointer))

						roomID = roomID + 1
						fmt.Println(`房號:`, roomID, getLoginBasicInfoString(clientPointer))
						logger.Infof(`房號:`, roomID, getLoginBasicInfoString(clientPointer))

						if jsonBytes, err := json.Marshal(RoomIDResponse{Command: 3, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID, RoomID: roomID}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

							fmt.Println(`回覆指令<取得空房號>成功`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`回覆指令<取得空房號>成功`, getLoginBasicInfoString(clientPointer))

						} else {
							fmt.Println(`json出錯`, getLoginBasicInfoString(clientPointer))
							logger.Warnf(`json出錯`, getLoginBasicInfoString(clientPointer))
						}

					case 4: // 求助

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 4, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<求助>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<求助>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 檢查欄位是否齊全
						fields := []string{"pic", "roomID"}
						ok, missFields := checkCommandFields(command, fields)
						if !ok {

							m := strings.Join(missFields, ",")
							fmt.Println("[欄位不齊全]", m)
							if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()
						fmt.Println(`收到指令<求助>`, getLoginBasicInfoString(clientPointer))
						logger.Infof(`收到指令<求助>`, getLoginBasicInfoString(clientPointer))

						// 設定Pic, RoomID
						element := clientInfoMap[clientPointer]
						element.Device.Pic = command.Pic       // Pic
						element.Device.DeviceStatus = 2        // 設備狀態:求助中
						element.Device.RoomID = command.RoomID // RoomID
						clientInfoMap[clientPointer] = element // 回存Map

						//Response
						if jsonBytes, err := json.Marshal(HelpResponse{Command: 4, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							fmt.Println(`回覆指令<求助>成功`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`回覆指令<求助>成功`, getLoginBasicInfoString(clientPointer))
						} else {
							fmt.Println(`json出錯`)
							logger.Warnf(`json出錯`)
							//return
							break
						}

						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array

						fmt.Println("deviceArray=", deviceArray)
						// 對區域廣播-求助
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: 9, CommandType: 3, Device: deviceArray}); err == nil {
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播
							fmt.Println(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
						}

					case 5: // 回應求助

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 5, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<回應求助>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<回應求助>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 檢查欄位是否齊全
						fields := []string{"roomID", "deviceID", "deviceBrand"}
						ok, missFields := checkCommandFields(command, fields)
						if !ok {

							m := strings.Join(missFields, ",")
							fmt.Println("[欄位不齊全]", m)
							logger.Warnf(`[欄位不齊全]`, m)
							if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
								return
							}
							//return
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()
						fmt.Println(`收到指令<回應求助>`, getLoginBasicInfoString(clientPointer))
						logger.Infof(`收到指令<回應求助>`, getLoginBasicInfoString(clientPointer))

						// 設定回應者的設備狀態+房間(自己)
						element := clientInfoMap[clientPointer]
						element.Device.DeviceStatus = 3        // 通話中
						element.Device.RoomID = command.RoomID // 想回應的RoomID
						clientInfoMap[clientPointer] = element // 存回Map

						// 設定求助者的狀態
						devicePointer := getDevice(command.DeviceID, command.DeviceBrand)
						if devicePointer != nil {
							devicePointer.DeviceStatus = 3 // 通話中
						} else {
							// 求助者不存在
							// Response
							if jsonBytes, err := json.Marshal(AnswerResponse{Command: 5, CommandType: 2, ResultCode: 1, Results: `求助者不存在`, TransactionID: command.TransactionID}); err == nil {
								//Response
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}
								fmt.Println(`回覆指令<回應求助>失敗-求助者不存在`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<回應求助>失敗-求助者不存在`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)

							}
						}

						fmt.Println(`回應者:`, getLoginBasicInfoString(clientPointer), ` 求助者:DeviceID:`, command.DeviceID, `,DeviceBrand:`, command.DeviceBrand)
						logger.Infof(`回應者:`, getLoginBasicInfoString(clientPointer), ` 求助者:DeviceID:`, command.DeviceID, `,DeviceBrand:`, command.DeviceBrand)

						// Response
						if jsonBytes, err := json.Marshal(AnswerResponse{Command: 5, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
							//Response
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}
							fmt.Println(`回覆指令<回應求助>成功`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`回覆指令<回應求助>成功`, getLoginBasicInfoString(clientPointer))
						} else {
							fmt.Println(`json出錯`)
							logger.Warnf(`json出錯`)
							//return
							break
						}

						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成Array:放入回應者device
						deviceArray = append(deviceArray, *devicePointer)            // 包成Array:放入求助者device

						fmt.Println("datas[]=", deviceArray)

						// 對區域廣播-回應求助
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: 9, CommandType: 3, Device: deviceArray}); err == nil {
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播
							fmt.Println(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
						}

					case 6: // 變更	cam+mic 狀態

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 6, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<變更cam+mic狀態>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<變更cam+mic狀態>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 檢查欄位是否齊全
						fields := []string{"cameraStatus", "micStatus"}
						ok, missFields := checkCommandFields(command, fields)
						if !ok {
							m := strings.Join(missFields, ",")
							fmt.Println("[欄位不齊全]", m)
							if jsonBytes, err := json.Marshal(LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: "欄位不齊全 " + m, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 基本資訊
						//basicInfo := getLoginBasicInfoString(clientPointer)

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()
						fmt.Println(`收到指令<變更攝影機+麥克風狀態>`, getLoginBasicInfoString(clientPointer))
						logger.Infof(`收到指令<變更攝影機+麥克風狀態>`, getLoginBasicInfoString(clientPointer))

						// 設定攝影機、麥克風
						element := clientInfoMap[clientPointer]            // 取出device
						element.Device.CameraStatus = command.CameraStatus // 攝影機
						element.Device.MicStatus = command.MicStatus       // 麥克風
						clientInfoMap[clientPointer] = element             // 回存Map

						//fmt.Println("更新後設備清單", deviceList)

						//Response
						if jsonBytes, err := json.Marshal(CamMicResponse{Command: 6, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
						} else {
							fmt.Println(`json出錯`)
							logger.Warnf(`json出錯`)
							//return
							break
						}

						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array

						fmt.Println("deviceArray=", deviceArray)
						// 對房間廣播-改變麥克風/攝影機狀態
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: 10, CommandType: 3, Device: deviceArray}); err == nil {
							broadcastByRoomID(clientInfoMap[clientPointer].Device.RoomID, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播
							fmt.Println(`【廣播】(房間)狀態變更`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`【廣播】(房間)狀態變更`, getLoginBasicInfoString(clientPointer))
						}

					case 7: // 掛斷通話

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 7, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<掛斷通話>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<掛斷通話>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()
						fmt.Println(`收到指令<掛斷通話>`, getLoginBasicInfoString(clientPointer))
						logger.Infof(`收到指令<掛斷通話>`, getLoginBasicInfoString(clientPointer))

						// 設定掛斷者cam, mic
						element := clientInfoMap[clientPointer] // 取出device
						element.Device.DeviceStatus = 1         // 閒置
						element.Device.RoomID = 0               // 沒有房間
						clientInfoMap[clientPointer] = element  // 回存Map

						//fmt.Println("更新後設備清單", &deviceList)

						//Response
						if jsonBytes, err := json.Marshal(RingOffResponse{Command: 7, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							fmt.Println(`回覆指令<掛斷通話>成功`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`回覆指令<掛斷通話>成功`, getLoginBasicInfoString(clientPointer))
						} else {
							fmt.Println(`json出錯`)
							logger.Warnf(`json出錯`)
							//return
							break
						}

						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array

						//fmt.Println("deviceArray=", deviceArray)

						// 對區域廣播-裝置狀態變成<閒置>
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: 9, CommandType: 3, Device: deviceArray}); err == nil {
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播
							fmt.Println(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
						}

					case 8: // 登出

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 8, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<登出>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<登出>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()
						fmt.Println(`收到指令<登出>`, getLoginBasicInfoString(clientPointer))
						logger.Infof(`收到指令<登出>`, getLoginBasicInfoString(clientPointer))

						// 設定登出者
						element := clientInfoMap[clientPointer] // 取出device
						element.Device.OnlineStatus = 2         // 離線
						clientInfoMap[clientPointer] = element  // 回存Map

						//fmt.Println("更新後設備清單", deviceList)

						// Response
						if jsonBytes, err := json.Marshal(LogoutResponse{Command: 8, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID}); err == nil {
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

							list := printDeviceList() // 裝置清單
							fmt.Println(`回覆指令<登出>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)
							logger.Infof(`回覆指令<登出>成功`, getLoginBasicInfoString(clientPointer), ` 裝置清單:`, list)
						} else {
							fmt.Println(`json出錯`)
							logger.Warnf(`json出錯`)
							//return
							break
						}

						deviceArray := getArray(clientInfoMap[clientPointer].Device) // 包成array

						//fmt.Println("deviceArray=", deviceArray)

						// 對區域廣播-裝置狀態變成<閒置>
						if jsonBytes, err := json.Marshal(DeviceStatusChange{Command: 9, CommandType: 3, Device: deviceArray}); err == nil {
							broadcastByArea(clientInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播
							fmt.Println(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`【廣播】(場域)狀態變更`, getLoginBasicInfoString(clientPointer))
						}

						// 移除連線
						//delete(clientInfoMap, clientPointer) //刪除
						disconnectHub(clientPointer) //斷線
						list := printDeviceList()    //印出裝置清單
						fmt.Println("已移除Socket連線於在線清單。目前清單為:", list)

					case 9: // 心跳包

						// 是否已登入(TransactionID 外層已經檢查過)
						if !checkLogedIn(clientPointer) {
							// 無登入資料
							if jsonBytes, err := json.Marshal(HelpResponse{Command: 9, CommandType: 2, ResultCode: 1, Results: `連線尚未登入`, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								fmt.Println(`回覆指令<心跳包>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
								logger.Infof(`回覆指令<心跳包>失敗-連線尚未登入`, getLoginBasicInfoString(clientPointer))
							} else {
								fmt.Println(`json出錯`)
								logger.Warnf(`json出錯`)
							}
							//return
							break
						}

						// 該有欄位外層已判斷

						// 基本資訊
						//basicInfo := getLoginBasicInfoString(clientPointer)

						commandTimeChannel <- time.Now() // 當送來指令，更新心跳包通道時間
						fmt.Println(`收到指令<心跳包>`, getLoginBasicInfoString(clientPointer))

						/* 待補:檢查有沒有漏欄位*/

						if jsonBytes, err := json.Marshal(Heartbeat{Command: 8, CommandType: 4}); err == nil {
							// fmt.Println(`json`, jsonBytes)
							// broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播檔案內容
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //依照原socket客戶端 返回內容
							fmt.Println(`回覆指令<心跳包>成功`, getLoginBasicInfoString(clientPointer))
							logger.Infof(`回覆指令<心跳包>成功`, getLoginBasicInfoString(clientPointer))
						} else {
							fmt.Println(`json出錯`)
							logger.Warnf(`json出錯`)
							//return
							break
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
