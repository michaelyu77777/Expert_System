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

	//指令
	Command     int `json:"command"`
	CommandType int `json:"commandType"`

	//登入Info
	UserID        string `json:"userID"`        //使用者登入帳號
	UserPassword  string `json:"userPassword"`  //使用者登入密碼
	DeviceID      string `json:"deviceID"`      //裝置ID
	DeviceBrand   string `json:"deviceBrand"`   //裝置品牌(怕平板裝置的ID會重複)
	DeviceType    int    `json:"deviceType"`    //裝置類型
	TransactionID string `json:"transactionID"` //分辨多執行緒順序不同的封包

	//測試用
	Argument1 string `json:"argument1"`
	Argument2 string `json:"argument2"`
}

// 心跳包 Response
type Heartbeat struct {
	Command     int `json:"command"`
	CommandType int `json:"commandType"`
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
	Area         string `json:"area"`         //場域
	DeviceName   string `json:"deviceName"`   //裝置名稱
	Pic          string `json:"pic"`          //裝置截圖
	OnlineStatus int    `json:"onlineStatus"` //在線狀態
	DeviceStatus int    `json:"deviceStatus"` //設備狀態
	CameraStatus int    `json:"cameraStatus"` //相機狀態
	MicStatus    int    `json:"micStatus"`    //麥克風狀態
	RoomID       int    `json:"roomID"`       //房號
}

// 登入 Response
type LoginResponse struct {
	//指令
	Command       int    `json:"command"`
	CommandType   int    `json:"commandType"`
	ResultCode    int    `json:"resultCode"`
	Results       string `json:"results"`
	TransactionID string `json:"transactionID"`
}

// 裝置狀態改變 Broadcast(廣播)
type DevicesStatusChange struct {
	//指令
	Command     int      `json:"command"`
	CommandType int      `json:"commandType"`
	Devices     []Device `json:"datas"`
}

//測試用
type Info struct {
	Name   string `json:"name"`
	Status int    `json:"status"`

	// UserID        string `json:"userID"`        //使用者登入帳號
	// UserPassword  string `json:"userPassword"`  //使用者登入密碼
	// DeviceID      string `json:"deviceID"`      //裝置ID
	// DeviceBrand   string `json:"deviceBrand"`   //裝置品牌(怕平板裝置的ID會重複)
	// DeviceType    int    `json:"deviceType"`    //裝置類型
	// TransactionID string `json:"transactionID"` //分辨多執行緒順序不同的封包

}

// 測試用
type OnlineStatus struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

// 測試用
var clientInfoMap = make(map[*client]Info)

// Map-連線/登入資訊
var clientLoginInfoMap = make(map[*client]LoginInfo)

// 所有裝置清單
var deviceList []Device

// 連線逾時時間:
//var timeout int = 30
const timeout = 10

// 從清單移除某裝置device
func removeDevice(slice []Device, s int) []Device {
	return append(slice[:s], slice[s+1:]...) //回傳移除後的array
}

// 從清單移除某連線client

// 增加裝置到清單
func addDeviceToList(device Device) bool {

	// 若裝置重複，則重新登入
	for i, _ := range deviceList {
		if deviceList[i].DeviceID == device.DeviceID {
			if deviceList[i].DeviceBrand == device.DeviceBrand {

				//移除舊的
				deviceList = removeDevice(deviceList, i)
				fmt.Println("移除後清單", deviceList)

				//新增新的
				deviceList = append(deviceList, device)
				fmt.Println("新增後清單", deviceList)

				fmt.Println("裝置重新登入")
				//fmt.Println("裝置清單:", deviceList)
				return true // 回傳
			}
		}
	}

	deviceList = append(deviceList, device) //新增裝置
	fmt.Println("裝置清單:", deviceList)
	return true
}

// 修改裝置狀態
func changeDeviceStatus(deviceID string, deviceBrand string, onlineStatus int, deviceStatus int, cameraStatus int, micStatus int) {

	// 找裝置
	for i, _ := range deviceList {
		if deviceList[i].DeviceID == deviceID {
			if deviceList[i].DeviceBrand == deviceBrand {

				//fmt.Println("找到裝置 deviceID=", deviceID, " deviceBrand=", deviceBrand)
				if onlineStatus != -1 {
					deviceList[i].OnlineStatus = onlineStatus
					fmt.Println("修改 onlineStatus=", onlineStatus)
				}
				if deviceStatus != -1 {
					deviceList[i].DeviceStatus = deviceStatus
					fmt.Println("修改 deviceStatus=", deviceStatus)
				}
				if cameraStatus != -1 {
					deviceList[i].CameraStatus = cameraStatus
					fmt.Println("修改 cameraStatus=", cameraStatus)
				}
				if micStatus != -1 {
					deviceList[i].MicStatus = micStatus
					fmt.Println("修改 micStatus=", micStatus)
				}
				//fmt.Println("修改完成")
			}
		}
		//fmt.Println("修改後裝置清單:")
		//fmt.Println(fmt.Sprintf("%d: %s", i+1, e))
	}
}

// 排除某連線進行廣播 (excluder 被排除的client)
func broadcastExceptOne(excluder *client, websocketData websocketData) {

	//Response to all
	for client, _ := range clientLoginInfoMap {

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
func broadcastByArea(area string, websocketData websocketData, excluder *client) {

	for client, _ := range clientLoginInfoMap {

		// 找到相同場域的連線
		if clientLoginInfoMap[client].Device.Area == area {
			if client != excluder { //排除自己
				client.outputChannel <- websocketData //Socket Response
			}
		}
	}
}

// 針對某房間(RoomID)進行廣播，排除某連線(自己)
func broadcastByRoomID(roomID int, websocketData websocketData, excluder *client) {

	for client, _ := range clientLoginInfoMap {

		// 找到相同房間的連線
		if clientLoginInfoMap[client].Device.RoomID == roomID {
			if client != excluder { //排除自己
				client.outputChannel <- websocketData //Socket Response
			}
		}
	}
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

			commandTimeChannel := make(chan time.Time, 1) // 連線逾時計算之通道(時間，1個buffer)
			go func() {
				fmt.Println("開始偵測連線逾時")
				for {
					commandTime := <-commandTimeChannel                                  // 當有接收到指令，則會有值在此通道
					<-time.After(commandTime.Add(time.Second * timeout).Sub(time.Now())) // 若超過時間，則往下進行
					if 0 == len(commandTimeChannel) {                                    // 若通道裡面沒有值，表示沒有收到新指令過來，則斷線

						fmt.Println("接收超時")
						disconnectHub(clientPointer) //斷線

						// 設定裝置在線狀態=離線
						id := clientLoginInfoMap[clientPointer].Device.DeviceID
						brand := clientLoginInfoMap[clientPointer].Device.DeviceBrand
						changeDeviceStatus(id, brand, 0, -1, -1, -1)

						// 移除連線
						delete(clientLoginInfoMap, clientPointer) //刪除

						// 【廣播】狀態變更
						if jsonBytes, err := json.Marshal(DevicesStatusChange{Command: 9, CommandType: 3, Devices: deviceList}); err == nil {
							// 	fmt.Println(`json`, jsonBytes)
							// 	fmt.Println(`json string`, string(jsonBytes))
							//broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播檔案內容
							broadcastExceptOne(clientPointer, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 排除個人進行廣播

							fmt.Println("廣播<狀態變更>")
						} else {
							fmt.Println(`json出錯`)
						}
					}
				}
			}()

			for { // 循環讀取連線傳來的資料

				inputWebsocketData, isSuccess := getInputWebsocketDataFromConnection(connectionPointer) // 從客戶端讀資料
				fmt.Println(`從客戶端讀取資料`)

				if !isSuccess { //若不成功 (Socket斷線)

					fmt.Println(`讀取資料不成功`)

					myInfo, myInfoOK := clientInfoMap[clientPointer] // 查看名單的key有無客戶端

					if myInfoOK { // 如果有名為客戶端的key

						// broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: []byte(fmt.Sprintf(`%s 斷線了`, myInfo.Name))}) // 廣播檔案內容

						myInfo.Status = 0 //狀態設定為斷線

						clientInfoMap[clientPointer] = myInfo // 重新設定名單的key客戶端的資訊

						if jsonBytes, err := json.Marshal(OnlineStatus{Name: myInfo.Name, Status: 0}); err == nil {
							// fmt.Println(`json`, jsonBytes)
							// fmt.Println(`json string`, string(jsonBytes))
							//broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播說此key離線
							//broadcastExceptOne(clientPointer, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 排除個人進行廣播
							broadcastByArea(clientLoginInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

							fmt.Println(myInfo.Name + `離線`)
						}

						fmt.Println(`刪除客戶` + myInfo.Name)
						delete(clientInfoMap, clientPointer) //刪除

					}

					return // 回傳
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
					json.Unmarshal(dataBytes, &command)

					fmt.Println(`Command`, command)

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

					switch c := command.Command; c {

					case 8: // 心跳包

						commandTimeChannel <- time.Now() // 當送來指令，更新心跳包通道時間
						fmt.Println(`收到心跳包指令`)

						if jsonBytes, err := json.Marshal(Heartbeat{Command: 8, CommandType: 4}); err == nil {
							// fmt.Println(`json`, jsonBytes)
							// fmt.Println(`json string`, string(jsonBytes))
							// broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播檔案內容
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //依照原socket客戶端 返回內容
						} else {
							fmt.Println(`json出錯`)
						}

					case 1: //登入+廣播改變狀態

						commandTimeChannel <- time.Now() // 當送來指令，更新心跳包通道時間
						fmt.Println(`收到登入指令`)

						// 判斷密碼是否正確
						userid := command.UserID
						userpassword := command.UserPassword

						resultCode := 0 //結果代碼
						results := ""   //錯誤內容

						// 裝置
						var device Device

						//測試帳號 id:001 pw:test
						if userid == "001" { //從資料庫找
							check := userpassword == "test" //從資料庫找比對
							if check {

								// 帳號正確
								resultCode = 0 //正常
								fmt.Println("密碼正確")

								// 設定裝置
								device = Device{
									DeviceID:     command.DeviceID,
									DeviceBrand:  command.DeviceBrand,
									DeviceType:   command.DeviceType,
									Area:         "Leapsy",                  //從資料庫撈
									DeviceName:   "001+" + command.DeviceID, //從資料庫撈
									Pic:          "",                        //<求助>時才會從客戶端得到
									OnlineStatus: 1,                         //在線
									DeviceStatus: 0,                         //閒置
									MicStatus:    0,                         //麥克風
									CameraStatus: 0,                         //相機
									RoomID:       0,                         //房號
								}

								// 裝置加入清單
								if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: resultCode, Results: results, TransactionID: command.TransactionID}); err == nil {

									//deviceList = append(deviceList, device) //新增裝置
									if addDeviceToList(device) {

										// 加入成功
										fmt.Println("加入裝置清單成功")

										clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

										// 記錄<連線>與<登入資訊>的配對
										clientLoginInfoMap[clientPointer] = LoginInfo{UserID: command.UserID, UserPassword: command.UserPassword, Device: device, TransactionID: command.TransactionID}

										// 將自己包成 array
										var d = []Device{}
										d = append(d, device)

										// 【廣播-場域】狀態變更
										if jsonBytes, err := json.Marshal(DevicesStatusChange{Command: 9, CommandType: 3, Devices: d}); err == nil {
											// 	fmt.Println(`json`, jsonBytes)
											// 	fmt.Println(`json string`, string(jsonBytes))

											//broadcastHubWebsocketData(websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 廣播檔案內容
											//broadcastExceptOne(clientPointer, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}) // 排除個人進行廣播
											broadcastByArea(clientLoginInfoMap[clientPointer].Device.Area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

											fmt.Println("廣播-場域-狀態變更")
										} else {
											fmt.Println(`json出錯`)
										}

									} else {

										// 加入錯誤
										resultCode = 1
										results += "加入裝置清單時錯誤"

										if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: resultCode, Results: results, TransactionID: command.TransactionID}); err == nil {
											clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
										} else {
											fmt.Println(`json出錯`)
										}
									}
								} else {
									fmt.Println(`json出錯`)
								}

							} else {
								resultCode = 1
								results = "密碼錯誤"
								fmt.Println("密碼錯誤")

								if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: resultCode, Results: results, TransactionID: command.TransactionID}); err == nil {
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
								} else {
									fmt.Println(`json出錯`)
									//return
								}
								// disconnectHub(clientPointer)
							}
						} else {
							resultCode = 1
							results = "無此帳號"
							fmt.Println("無此帳號")

							if jsonBytes, err := json.Marshal(LoginResponse{Command: 1, CommandType: 2, ResultCode: resultCode, Results: results, TransactionID: command.TransactionID}); err == nil {
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response
							} else {
								fmt.Println(`json出錯`)
								//return
							}
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
