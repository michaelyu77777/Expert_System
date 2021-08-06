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

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"leapsy.com/packages/configurations"
	"leapsy.com/packages/jwts"
	"leapsy.com/packages/logings"
	"leapsy.com/packages/network"
	"leapsy.com/packages/paths"
	"leapsy.com/packages/serverDataStruct"
	"leapsy.com/packages/serverResponseStruct"
	// "leapsy.com/packages/model"
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

// Map-連線/登入資訊
var clientInfoMap = make(map[*client]*serverDataStruct.Info)

// 加密解密KEY(AES加密)（key 必須是 16、24 或者 32 位的[]byte）
// const key_AES = "RB7Wfa$WHssV4LZce6HCyNYpdPPlYnDn" //32位數

// 連線逾時時間
var timeout = time.Duration(configurations.GetConfigPositiveIntValueOrPanic(`local`, `timeout`)) // 轉成time.Duration型態，方便做時間乘法

// demo 模式是否開啟
var expertdemoMode = configurations.GetConfigPositiveIntValueOrPanic(`local`, `expertdemoMode`)

// 房間號(總計)
var roomID = 0

// 基底: Response Json
var baseResponseJsonString = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"}`
var baseResponseJsonStringExtend = `{"command":%d,"commandType":%d,"resultCode":%d,"results":"%s","transactionID":"%s"` // 可延展的

// 基底: 共用(指令執行成功、指令失敗、失敗原因、廣播、指令結束)
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

// keepReading - 保持讀取
func (clientPointer *client) keepReading() {

	// 各樣指令操作工具
	cTool := &CommandTool{}

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
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								break // 跳出
							}
						}
					}

					commandTime := <-commandTimeChannel                                  // 當有接收到指令，則會有值在此通道
					<-time.After(commandTime.Add(time.Second * timeout).Sub(time.Now())) // 若超過時間，則往下進行
					if 0 == len(commandTimeChannel) {                                    // 若通道裡面沒有值，表示沒有收到新指令過來，則斷線

						details := `<此裝置發生逾時,即將斷線>`

						// Response:通知連線即將斷線
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, CommandNumberOfLogout, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, ""))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 設定裝置在線狀態=離線
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {
							devicePointer := infoPointer.DevicePointer
							if nil != devicePointer {

								_, message := cTool.setDevicePointerOffline(devicePointer)
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
								deviceArray := cTool.getArrayPointer(devicePointer) // 包成array

								// (場域)廣播：狀態改變
								messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, serverResponseStruct.Command{}, clientPointer, deviceArray, details)
								// broadcastByArea(tempArea, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, whatKindCommandString, serverResponseStruct.Command{}, clientPointer, details) // 排除個人進行廣播

								details += `-(場域)廣播,此連線已逾時,此裝置狀態已變更為:離線,詳細訊息:` + messages
								// 一般logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							}
						}

						// 移除連線
						delete(clientInfoMap, clientPointer) //刪除
						disconnectHub(clientPointer)         //斷線

						details += `-已斷線`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					}

				}
			}()

			for { // 循環讀取連線傳來的資料

				whatKindCommandString := `不斷從客戶端讀資料`

				inputWebsocketData, isSuccess := getInputWebsocketDataFromConnection(connectionPointer) // 不斷從客戶端讀資料

				if !isSuccess { //若不成功 (判斷為Socket斷線)

					details := `-從客戶讀取資料失敗,即將斷線`

					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
					cTool.processLoggerInfof(whatKindCommandString, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					disconnectHub(clientPointer) //斷線

					details += `-已斷線`
					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
					cTool.processLoggerInfof(whatKindCommandString, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					return // 回傳:離開for迴圈(keep Reading)

				} else {

					details := `-從客戶讀取資料成功`

					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
					cTool.processLoggerInfof(whatKindCommandString, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				}

				wsOpCode := inputWebsocketData.wsOpCode
				dataBytes := inputWebsocketData.dataBytes

				if ws.OpText == wsOpCode {

					var command serverResponseStruct.Command

					whatKindCommandString := `收到指令，初步解譯成Json格式`

					//解譯成Json
					err := json.Unmarshal(dataBytes, &command)

					//json格式錯誤
					if err == nil {
						// 執行成功
						details := `-指令已成功解譯為json格式`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					} else {
						details := `-指令解譯失敗,json格式錯誤`

						// 警告logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					}

					// 檢查command欄位是否都給了
					fields := []string{"command", "commandType", "transactionID"}
					ok, missFields := cTool.checkCommandFields(command, fields)
					if !ok {

						whatKindCommandString := `收到指令檢查欄位`

						//遺失的欄位
						m := strings.Join(missFields, ",")

						details := `<執行失敗,欄位不完全,遺失欄位: ` + m + `>`

						// Socket Response
						if jsonBytes, err := json.Marshal(serverResponseStruct.LoginResponse{Command: command.Command, CommandType: command.CommandType, ResultCode: 1, Results: details, TransactionID: command.TransactionID}); err == nil {

							// 失敗:欄位不完全
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// break
						} else {
							// 錯誤logger
							// details := `-執行失敗,欄位不完全,但JSON轉換錯誤`
							details = `<JSON轉換錯誤>且` + details
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							// break
						}
						//return
					}

					// 判斷指令
					switch c := command.Command; c {

					case 1: // 平板 驗證碼登入 + 一線人員預設帳號登入

						whatKindCommandString := `登入`

						// 檢查欄位<帳號驗證功能>是否齊全
						if !cTool.checkFieldsCompletedAndResponseIfFail([]string{"userID", "userPassword", "deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備驗證密碼:拿ID+密碼去資料庫比對密碼，若正確則進行登入

						check := false
						accountPointer := &serverDataStruct.Account{}

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
						check, accountPointer = cTool.checkPasswordAndGetAccountPointer(command.UserID, command.UserPassword)

						// }

						// 驗證密碼成功:
						if check {

							details += `-驗證密碼成功`

							if accountPointer != nil {

								// 看專家帳號的驗證碼是否過期（資料庫有此帳號,且為專家帳號,才會用驗證信）
								// 僅檢查專家帳號，一線人員帳號，不會用驗證碼
								if 1 == accountPointer.IsExpert {
									details += `-此為專家帳號`
								}

								if 1 == accountPointer.IsFrontline {
									details += `-此為一線人員帳號`
								}

								// 看驗證碼是否過期
								isBefore := time.Now().Before(accountPointer.VerificationCodeValidPeriod) // 看是否還在期限內

								// m, _ := time.ParseDuration("10m") // 驗證碼有效時間
								// deadline := accountPointer.VerificationCodeValidPeriod.Add(m) // 此帳號驗證碼最後有效期限期限
								// fmt.Println("還在期限內？", isBefore)

								if !isBefore {
									// 已過期

									// details += `-驗證碼已過期,驗證失敗`
									details = `<驗證碼已過期,驗證失敗>` + details

									// Response：失敗
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// 警告logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
									break // 跳出
								}
							} else {

								// details += `-找不到帳號`
								details = `<找不到帳號>` + details

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, "資料庫找不到此帳號", command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出
							}

							// 取得裝置Pointer
							// devicePointer := getDevicePointer(command.DeviceID, command.DeviceBrand)

							// 改成此登入裝置的 DevicePointer
							devicePointer := cTool.getDevicePointerBySearchAndCreate(command.DeviceID, command.DeviceBrand)

							// 資料庫找不到此裝置
							if devicePointer == nil {

								// details += `-找不裝置`
								details = `<找不裝置>` + details

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出
							}

							// 進行裝置、帳號登入 (加入Map。包含處理裝置重複登入)
							if success, otherMessage := cTool.processLoginWithDuplicate(whatKindCommandString, clientPointer, command, devicePointer, accountPointer); success {
								// 成功

								// details += `-登入成功,回應客戶端`
								details = `<登入成功,回應客戶端>` + details

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 一般logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								deviceArray := cTool.getArrayPointer(devicePointer) // 包成array
								messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// 一般logger
								details += `-執行廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								// 失敗

								// details += `-登入失敗`
								details = `<登入失敗>` + details

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details+otherMessage, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details+otherMessage, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出
							}

						} else {
							// logger:帳密錯誤

							// details += `-驗證密碼失敗-無此帳號或密碼錯誤`
							details = `<驗證密碼失敗-無此帳號或密碼錯誤>` + details

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
							break // 跳出
						}

					case 15: // 平板 判斷帳號是否存在，若存在則寄出驗證信

						whatKindCommandString := `判斷帳號是否存在，若存在寄出驗證信`

						// 該有欄位外層已判斷

						// 不用登入就可使用

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 是否有此帳號
						haveAccount, accountPointer := cTool.checkAccountExistAndCreateAccountPointer(command.UserID)

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
								success, otherMessages = cTool.processSendVerificationCodeMail(accountPointer, whatKindCommandString, details, command, clientPointer)
							}

							if success {

								// details += `-驗證信已寄出`
								details = `<驗證信已寄出>` + details

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 一般logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {

								// details += `-驗證信寄出失敗,訊息:` + otherMessages
								details = `<驗證信寄出失敗,訊息:` + otherMessages + `>` + details

								// Response:失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							}

						} else {
							// details += `-找不到此帳號`
							details = `<找不到此帳號>` + details

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						}

					case 17: // 一線人員 QRcode登入

						whatKindCommandString := `QRcode登入`

						// 檢查<帳號驗證功能>欄位是否齊全
						if !cTool.checkFieldsCompletedAndResponseIfFail([]string{"userID", "deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// 一般logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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

						// 解密
						decryptionString := cTool.getDecryptionString(command.UserID)
						// token := jwts.ParseToken(command.UserID)

						fmt.Println("解密後結果為:", decryptionString)

						//若非nil 取出
						if decryptionString != "" {
							// 解密成功
							details += `-QR cdoe 解密成功`
						} else {
							// 解密出現錯誤或解完是空字串
							details = `<QR code 解密錯誤或結果為空字串>` + details

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 錯誤logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出

						}

						userid := decryptionString
						fmt.Println("解密後的userid=", userid)

						// 是否有此帳號
						check, accountPointer := cTool.checkAccountExistAndCreateAccountPointer(userid)

						// 找到帳號
						if check {

							details += `-找到帳號`

							// 找裝置Pointer
							// devicePointer := cTool.getDevicePointer(command.DeviceID, command.DeviceBrand)
							devicePointer := cTool.getDevicePointerBySearchAndCreate(command.DeviceID, command.DeviceBrand)

							// 找到裝置
							if devicePointer != nil {

								details += `-找到裝置`

								// 登入:裝置、帳號 (加入Map。包含處理裝置重複登入)
								if success, otherMeessage := cTool.processLoginWithDuplicate(whatKindCommandString, clientPointer, command, devicePointer, accountPointer); success {
									// 登入成功

									// details += `-登入成功`
									details = `<登入成功>` + details

									// Response:成功
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// 一般logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									// 準備廣播:包成Array:放入 Response Devices
									deviceArray := cTool.getArrayPointer(devicePointer) // 包成array
									messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

									// logger
									details += `-執行(區域)廣播,詳細訊息:` + messages
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								} else {
									// 登入失敗
									// details += `-登入失敗:` + otherMeessage
									details = `<登入失敗:` + otherMeessage + `>` + details

									// Response:失敗
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// 一般logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									break // 跳出

								}

							} else {
								// 找不到裝置
								// details += `-找不到裝置`
								details = `<找不到裝置>` + details

								// Response：失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								break // 跳出
							}

						} else {
							// logger:帳密錯誤
							// details += `-找不到帳號或密碼錯誤`
							details = `<找不到帳號或密碼錯誤>` + details

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// 警告logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出
						}

					case 2: // 專家帳號要取得跟專家自己同場域的所有眼鏡裝置清單

						whatKindCommandString := `專家帳號要取得跟專家自己同場域的所有眼鏡裝置清單`

						// 該有欄位外層已判斷

						// 是否已與Server建立連線
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 取得指定場域裝置清單
						if infoPointer, ok := clientInfoMap[clientPointer]; ok {

							// 檢查裝置
							if nil != infoPointer.DevicePointer {

								//取得跟專家場域一樣的、眼鏡裝置、並排除自己的所有裝置info
								infosInAreasExceptMineDevice, otherMessage := cTool.getDevicesWithInfoByAreaAndDeviceTypeExeptOneDevice(infoPointer.AccountPointer.Area, 1, infoPointer.DevicePointer) //待改

								// Response:成功
								// 此處json不直接轉成string,因為有 device Array型態，轉string不好轉
								if jsonBytes, err := json.Marshal(serverResponseStruct.InfosInTheSameAreaResponse{Command: 2, CommandType: 2, ResultCode: 0, Results: ``, TransactionID: command.TransactionID, Info: infosInAreasExceptMineDevice}); err == nil {

									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Response

									infoPointerString := cTool.getStringByInfoPointerArray(infosInAreasExceptMineDevice) // 將查詢結果轉成字串

									// 一般logger
									details += `-指令執行成功,取得清單為:` + infoPointerString + `,執行細節:` + otherMessage
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {

									// logger
									details += `-指令執行失敗,後端json轉換出錯,執行細節:` + otherMessage
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
									break // 跳出

								}
							} else {
								// 找不到裝置
								// details += `-找不到裝置`
								details += `<找不到裝置>` + details

								// Response:失敗
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// 警告logger
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								break // 跳出

							}
						} else {
							// 尚未建立連線
							// details += `-執行失敗，尚未建立連線`
							details = `<執行失敗，尚未建立連線>` + details

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break // 跳出
						}

					case 3: // 取得空房號

						whatKindCommandString := `取得空房號`

						// 該有欄位外層已判斷

						// 是否已登入
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, `whatKindCommandString`) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 增加房號
						roomID = roomID + 1

						// Response:成功
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+`, "roomID":%d}`, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID, roomID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details += `-指令執行成功,取得房號為` + strconv.Itoa(roomID)
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 4: // 求助

						whatKindCommandString := `求助`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 檢查欄位是否齊全
						if !cTool.checkFieldsCompletedAndResponseIfFail([]string{"pic", "roomID"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 檢核:房號未被取用過則失敗
						if command.RoomID > roomID {

							// details += `-執行失敗:房號未被取用過`
							details = `<執行失敗:房號未被取用過>` + details

							// Response:失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
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
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								//deviceArray := getArray(clientInfoMap[clientPointer].DevicePointer) // 包成array
								deviceArray := cTool.getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array
								messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// logger
								details += `-執行(區域)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, ``)
								break
							}

						} else {
							// Response:失敗
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, ``)
							break
						}

					case 5: // 回應求助

						whatKindCommandString := `回應求助`

						// 是否已登入
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 檢查欄位是否齊全
						if !cTool.checkFieldsCompletedAndResponseIfFail([]string{"deviceID", "deviceBrand"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 準備設定-求助者狀態

						// (求助者)裝置
						askerDevicePointer := cTool.getDevicePointer(command.DeviceID, command.DeviceBrand)

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
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									// 準備廣播:包成Array:放入 Response Devices
									deviceArray := cTool.getArrayPointer(giverDeivcePointer) // 包成Array:放入回應者device
									deviceArray = append(deviceArray, askerDevicePointer)    // 包成Array:放入求助者device

									// 進行區域廣播
									messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

									// logger
									details += `-執行(區域)廣播,詳細訊息:` + messages
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {
									details += `-(回應者)裝置不存在`
									cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
									break // 跳出
								}

							} else {
								details += `-(回應者)連線Info不存在`
								cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
								break // 跳出

							}

						} else {
							details += `-(求助者)裝置不存在`
							cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
							break // 跳出
						}

					case 6: // 變更	cam+mic 狀態

						whatKindCommandString := `變更攝影機+麥克風狀態`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 檢查欄位是否齊全
						if !cTool.checkFieldsCompletedAndResponseIfFail([]string{"cameraStatus", "micStatus"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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

								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								deviceArray := cTool.getArrayPointer(clientInfoMap[clientPointer].DevicePointer)

								messages := cTool.processBroadcastingDeviceChangeStatusInRoom(whatKindCommandString, command, clientPointer, deviceArray, details)

								// logger
								details += `-進行(房間)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								//找不到裝置
								details += `-找不到裝置`
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							// 找不到要求端連線info
							details += `-找不到要求端連線info`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 7: // 掛斷通話

						whatKindCommandString := `掛斷通話`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger:收到指令
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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
								otherDevicesPointer := cTool.getOtherDevicesInTheSameRoom(thisRoomID, clientPointer)

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
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									// 準備廣播:包成Array:放入 Response Devices
									// 要放入自己＋其他人
									deviceArray := cTool.getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array
									for _, e := range otherDevicesPointer {
										deviceArray = append(deviceArray, e)
									}

									messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

									// logger:進行廣播
									details += `-執行(區域)廣播,詳細訊息:` + messages
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {
									//找不到帳號
									details += `-找不到帳號`
									cTool.processResponseAccountNil(clientPointer, whatKindCommandString, command, details)
									break
								}

							} else {
								//找不到裝置
								details += `-找不到裝置`
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							//找不到要求端連線info
							details += `-找不到要求端連線info`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 8: // 登出

						whatKindCommandString := `登出`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間，登出就不用再重新計算了
						//commandTimeChannel <- time.Now()

						// logger:收到指令
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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
								_, message := cTool.setDevicePointerOffline(devicePointer)
								details += `-設置裝置為離線狀態` + message

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details += `-指令執行成功，連線已登出`
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								// deviceArray := getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array
								deviceArray := cTool.getArrayPointer(devicePointer) // 包成array

								// 進行廣播
								messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// 暫存即將斷線的資料
								// logger
								details += `-執行(區域)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 移除連線
								// 帳號包在clientInfoMap[clientPointer]裡面,會一併進行清空
								delete(clientInfoMap, clientPointer) //刪除
								disconnectHub(clientPointer)         //斷線

							} else {
								details += `-找不到裝置`
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
							}

						} else {
							details += `-找不到要求端連線Info`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
						}

					case 9: // 心跳包

						whatKindCommandString := `心跳包`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						// 成功:Response
						jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
						clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						// logger
						details += `-指令執行成功`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

					case 13: // 取得自己帳號資訊

						whatKindCommandString := `取得自己帳號資訊`

						// 該有欄位外層已判斷

						// 是否已與Server建立連線
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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
								if jsonBytes, err := json.Marshal(serverResponseStruct.MyAccountResponse{
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
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {
									details += `-後端json轉換出錯`

									// logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
								}
							} else {
								//帳戶為空
								details += `-找不到帳戶`
								cTool.processResponseAccountNil(clientPointer, whatKindCommandString, command, details)
							}

						} else {
							//info為空
							details += `-找不到要求端連線info`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
						}

					case 14: // 取得自己裝置資訊

						whatKindCommandString := `取得自己裝置資訊`

						// 該有欄位外層已判斷

						// 是否已經加到clientInfoMap中 表示已登入
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						if infoPointer, ok := clientInfoMap[clientPointer]; ok { //取info

							devicePointer := infoPointer.DevicePointer //取裝置

							if nil != devicePointer {

								details += `-找到裝置,裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								device := cTool.getDevicePointer(devicePointer.DeviceID, devicePointer.DeviceBrand) // 取得裝置清單-實體                                                                                     // 自己的裝置

								// Response:成功 (此處仍使用Marshal工具轉型，因考量有 物件{}形態，轉成string較為複雜。)
								if jsonBytes, err := json.Marshal(serverResponseStruct.MyDeviceResponse{Command: command.Command, CommandType: CommandTypeNumberOfAPIResponse, ResultCode: ResultCodeSuccess, Results: ``, TransactionID: command.TransactionID, Device: *device}); err == nil {
									//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

									// Response(場域、排除個人)
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger
									details += `-指令執行成功`
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								} else {

									// logger
									details += `-執行指令失敗，後端json轉換出錯。`
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								}
							} else {
								// 找不到裝置
								details += `-找不到裝置`
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							// 找不到info
							details += `-找不到要求端連線`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 16: // 取得現在同場域空閒專家人數

						whatKindCommandString := `取得現在同場域空閒專家人數`

						// 該有欄位外層已判斷

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

						if infoPointer, ok := clientInfoMap[clientPointer]; ok {
							details += `-找到要求端連線`

							devicePointer := infoPointer.DevicePointer //取裝置
							if nil != devicePointer {
								details += `-找到裝置,裝置ID=` + devicePointer.DeviceID + `,裝置Brand=` + devicePointer.DeviceBrand

								// 取得線上同場域閒置專家數
								onlinExperts := cTool.getOnlineIdleExpertsCountInArea(devicePointer.Area, whatKindCommandString, command, clientPointer)

								// Response:成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonStringExtend+`,"onlineExpertsIdle":%d}`, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID, onlinExperts))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

								// logger
								details += `-指令執行成功,取得現在同場域空閒專家人數=` + strconv.Itoa(onlinExperts)
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								//找不到裝置
								details += `-找不到裝置`
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}

						} else {
							//找不到info
							details += `-找不到要求端連線`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 18: // 眼鏡切換場域

						whatKindCommandString := `眼鏡切換場域`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 閒置中才可以切換場域（通話中、求助中無法切換場域）
						if !cTool.checkDeviceStatusIsIdleAndResponseIfFail(clientPointer, command, whatKindCommandString, "-檢查設備狀態為閒置才可切換場域") {
							break //跳出
						}

						// 眼鏡端才可以切換場域
						if !cTool.checkDeviceTypeIsGlassesAndResponseIfFail(clientPointer, command, whatKindCommandString, "-檢查設備為眼鏡端才可以切換場域") {
							break //跳出
						}

						// 檢查欄位是否齊全
						if !cTool.checkFieldsCompletedAndResponseIfFail([]string{"areaEncryptionString"}, clientPointer, command, whatKindCommandString) {
							break // 跳出case
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						details := `-收到指令`

						// logger
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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
							// details += `-執行指令失敗-字符轉換失敗-字符或數字轉換錯誤或解密錯誤`
							details = `<執行指令失敗-字符轉換失敗-字符或數字轉換錯誤或解密錯誤>` + details

							fmt.Println(details)

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break //跳出case
						}

						// 檢查是否有此場域代碼

						// if areaName, ok := areaNumberNameMap[newAreaNmuber]; ok {
						// 	details += `-找到此場域代碼,場域代碼=` + strconv.Itoa(newAreaNumber) + `,場域名稱=` + areaName
						// }

						// 找場域
						areaArray := cTool.getAreaById(newAreaNumber)
						// 暫存
						var oldArea []int        //舊場域代碼
						var oldAreaName []string //舊場域名

						var newAreaNumberArray []int  //新場域代碼
						var newAreaNameArray []string //新場域名

						if len(areaArray) > 0 {
							newAreaNumberArray = append(newAreaNumberArray, areaArray[0].Id) //封裝成array
							newAreaNameArray = append(newAreaNameArray, areaArray[0].Zhtw)   //封裝成array
							details += `-找到此場域代碼,場域代碼=` + strconv.Itoa(areaArray[0].Id) + `,場域名稱=` + areaArray[0].Zhtw
						} else {
							//失敗：沒有此場域區域
							// details += `-執行指令失敗-找不到此場域代碼與其對應名稱,場域代碼=` + strconv.Itoa(newAreaNumber)
							details = `<執行指令失敗-找不到此場域代碼與其對應名稱,場域代碼=` + strconv.Itoa(newAreaNumber) + `>` + details

							fmt.Println(details)

							// Response：失敗
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							// logger
							myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
							cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							break //跳出case
						}

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

									oldArea = devicePointer.Area         //暫存舊場域
									oldAreaName = devicePointer.AreaName //暫存舊場域名

									// 更新資料庫裝置的場域，並更新infoPointer中devicePointer的變數
									newDevicePointer := cTool.updateDeviceAreaIDInMongoDBAndGetOneDevicePointer(newAreaNumber, devicePointer)

									if nil != newDevicePointer {

										infoPointer.DevicePointer = newDevicePointer

										// Response:成功
										jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, ``, command.TransactionID))
										clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

										// logger
										// int [] string[] 轉換成string
										newAreaString := fmt.Sprint(infoPointer.DevicePointer.Area)
										newAreaNameString := fmt.Sprint(infoPointer.DevicePointer.AreaName)
										oldAreaString := fmt.Sprint(oldArea)
										oldAreaNameString := fmt.Sprint(oldAreaName)
										details += `-指令執行成功,已換成新場域,新場域代號=` + newAreaString + `,新場域名=` + newAreaNameString + `,舊場域代號=` + oldAreaString + `,舊場域名=` + oldAreaNameString
										myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
										cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

										// 準備廣播:包成Array:放入 Response Devices
										//deviceArray := getArray(clientInfoMap[clientPointer].DevicePointer) // 包成array
										deviceArray := cTool.getArrayPointer(clientInfoMap[clientPointer].DevicePointer) // 包成array

										// 廣播給舊場域的
										cTool.processBroadcastingDeviceChangeStatusInSomeArea(whatKindCommandString, command, clientPointer, deviceArray, oldArea, details)

										// 廣播給現在新場域的連線裝置
										messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

										// logger
										details += `-執行(區域)廣播,詳細訊息:` + messages
										myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
										cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
									} else {
										// details += `-指令執行失敗,執行更新後,卻找不到裝置`
										details = `<指令執行失敗,執行更新後,卻找不到裝置>` + details

										// Response：失敗
										jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
										clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

										myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
										cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

									}

								} else {
									//失敗：此裝置已經在這個場域，不進行切換
									// details += `-指令執行失敗,此裝置已經在這個場域(` + newAreaNameArray[0] + `)，不進行切換`
									details = `<指令執行失敗,此裝置已經在這個場域(` + newAreaNameArray[0] + `)，不進行切換>` + details
									fmt.Println(details)

									// Response：失敗
									jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
									clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

									// logger
									myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
									cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
									break
								}
							} else {
								//找不到裝置
								details += `-找不到裝置`
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break
							}
						} else {
							//找不到Info
							details += `-找不到要求端連線`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break
						}

					case 19: // 取消求助

						whatKindCommandString := `取消求助`

						// 是否已登入(TransactionID 外層已經檢查過)
						if !cTool.checkLogedInAndResponseIfFail(clientPointer, command, whatKindCommandString) {
							break //跳出
						}

						// 當送來指令，更新心跳包通道時間
						commandTimeChannel <- time.Now()

						// logger
						details := `-收到指令`
						myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
						cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

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
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

								// 準備廣播:包成Array:放入 Response Devices
								//deviceArray := getArray(clientInfoMap[clientPointer].DevicePointer) // 包成array
								deviceArray := cTool.getArrayPointer(devicePointer) // 包成array
								messages := cTool.processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString, command, clientPointer, deviceArray, details)

								// logger
								details += `-執行(區域)廣播,詳細訊息:` + messages
								myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom = cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
								cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

							} else {
								// 裝置為空
								details += `-找不到裝置`
								cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
								break

							}

						} else {
							// Info 為空
							details += `-找不到要求端連線`
							cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
							break

						}

					case 20:

						whatKindCommandString := `加密結果產生器`

						// 是否登入Admin帳號
						if `admin` == command.UserID && `adminadminadmin` == command.UserPassword {

							// 準備進行加密 封裝資料
							// userIDToken := jwts.TokenInfo{
							// 	Data: command.StringToEncryption,
							// 	//Data: "想加的文字",
							// }

							// // 加密結果
							// encryptionString := jwts.CreateToken(&userIDToken)

							encryptionString := cTool.getEncryptionString(command.StringToEncryption)
							fmt.Println(whatKindCommandString, ":加密後<", encryptionString, ">前面是加密結果")

							// fmt.Println(whatKindCommandString, ":加密後<", *encryptionString, ">前面是加密結果")
							//fmt.Println("加密後<", encryptionString, ">前面是加密結果")

							// Response 成功
							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, `成功加密結果為:`+encryptionString, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						} else {
							// Response:尚未登入admin
							details := `<尚未登入管理員>`

							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						}

						// case 12: // 加入房間 //未來要做多方通話再做

					case 21: //解密結果產生器
						whatKindCommandString := `解密結果產生器`
						// 是否登入Admin帳號
						if `admin` == command.UserID && `adminadminadmin` == command.UserPassword {

							decryptionString := cTool.getDecryptionString(command.StringToDecryption)
							// // 進行解密
							// token := jwts.ParseToken(command.StringToDecryption)
							// //token := jwts.ParseToken(*encryptionString)

							// fmt.Println(whatKindCommandString, "-解密後token：", token)

							// // 要取出的 data
							// var result string

							// // 取出內容 data
							// if token != nil {
							// 	result = token.Data
							// }

							fmt.Println(whatKindCommandString, "-解密後Data:", decryptionString)

							if "" == decryptionString {

								// Response 成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, `解密錯誤或結果為空字串`, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

							} else {

								// Response 成功
								jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeSuccess, `解密結果為:`+decryptionString, command.TransactionID))
								clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}
							}

						} else {
							// Response:尚未登入admin
							details := `<尚未登入管理員>`

							jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
							clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

						}
					}

				}

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
