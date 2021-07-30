package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"leapsy.com/packages/configurations"
	"leapsy.com/packages/logings"
	"leapsy.com/packages/network"
	"leapsy.com/packages/networkHub"
)

var (
	logger = logings.GetLogger() // 記錄器
)

// main - 主程式
func main() {

	/** 此分支版本:繼續進行例外處理的分支 **/

	go startWebsocketServer() // 啟動Websocket伺服器

	//go networkHub.SetSecretByteArray(networkHub.GetNewSecretByteArray())
	// networkHub.Test() //測試心跳包用

	select {} // 阻止主程式結束
}

// startWebsocketServer - 啟動Websocket伺服器
func startWebsocketServer() {

	// 匯入更新所有裝置清單(之後待改成固定時間更新)
	// go networkHub.UpdateAllDevicesList()
	// go networkHub.UpdateAllAccountList()
	// go networkHub.UpdateAllAreaMap()

	// cTool := &networkHub.CommandTool{}
	// go cTool.UpdateAllDevicesList()
	// go cTool.UpdateAllAccountList()
	// go cTool.UpdateAllAreaMap()

	address := fmt.Sprintf(`%s:%d`,
		configurations.GetConfigValueOrPanic(`local`, `host`),
		configurations.GetConfigPositiveIntValueOrPanic(`local`, `port`),
	) // 預設主機

	network.SetAddressAlias(address, `Websocket伺服器`) // 設定預設主機別名

	enginePointer := gin.Default()

	// 設定url路徑與對應function
	enginePointer.GET(
		`/websocket`,
		getWebsocketHandler,
	)

	var enginePointerRunError error // 伺服器啟動錯誤

	go func() {
		enginePointerRunError = enginePointer.Run(address) // 啟動伺服器或回傳伺服器啟動錯誤
	}()

	<-time.After(time.Second * 3) // 等待伺服器啟動結果

	// 取得記錄器格式和參數
	formatString, args := logings.GetLogFuncFormatAndArguments(
		[]string{`%s %s 啟動`},
		network.GetAliasAddressPair(address),
		enginePointerRunError,
	)

	if nil != enginePointerRunError { // 若伺服器啟動錯誤
		logger.Panicf(formatString, args...) // 記錄錯誤並逐層結束程式
	} else { // 若伺服器啟動成功
		logger.Infof(formatString, args...) // 記錄資訊
	}

}

// getWebsocketHandler - 處理websocket
/**
 * @param  *gin.Context ginContextPointer  gin Context 指標
 */
func getWebsocketHandler(ginContextPointer *gin.Context) {

	//從socket升級成webSocket
	connection, _, _, wsUpgradeHTTPError := ws.UpgradeHTTP(ginContextPointer.Request, ginContextPointer.Writer)

	address := fmt.Sprintf(`%s:%d`,
		configurations.GetConfigValueOrPanic(`local`, `host`),
		configurations.GetConfigPositiveIntValueOrPanic(`local`, `port`),
	) // 預設主機

	defaultArgs := network.GetAliasAddressPair(address) // 記錄器預設參數

	// 取得記錄器格式字串與參數
	formatString, args := logings.GetLogFuncFormatAndArguments(
		[]string{`%s %s 連接 %s `},
		append(defaultArgs, connection.RemoteAddr().String()),
		wsUpgradeHTTPError,
	)

	if nil != wsUpgradeHTTPError { // 若伺服器啟動錯誤
		logger.Panicf(formatString, args...) // 記錄錯誤並逐層結束程式
	} else { // 若伺服器啟動成功
		go logger.Infof(formatString, args...) // 記錄資訊

		go networkHub.HandleNewConnection(&connection)
		// deleteInvalidAndOutputValidFiles(&connection)
	}

}

// deleteInvalidAndOutputValidFiles - 刪除失效檔案並輸出有效檔案
/**
 * @param  *net.Conn connectionPointer  連線指標
 */
// func deleteInvalidAndOutputValidFiles(connectionPointer *net.Conn) {
//
// 	path := configurations.GetConfigValueOrPanic(`local`, `path`) // 取得預設路徑
//
// 	files, ioutilReadDirError := ioutil.ReadDir(path) // 讀取資料夾
//
// 	connection := *connectionPointer // 連線
//
// 	formatSlices := []string{`%s %s 回應 %s 請求讀取路徑 %s `} // 格式片段
//
// 	// 記錄器預設參數
// 	defaultArgs := append(
// 		network.GetAliasAddressPair(connection.LocalAddr().String()),
// 		connection.RemoteAddr().String(),
// 		path,
// 	)
//
// 	// 取得記錄器格式字串與參數
// 	formatString, args := logings.GetLogFuncFormatAndArguments(
// 		formatSlices,
// 		defaultArgs,
// 		ioutilReadDirError,
// 	)
//
// 	if nil != ioutilReadDirError { // 若讀取資料夾錯誤，則記錄錯誤
// 		logger.Errorf(formatString, args...) // 記錄錯誤
// 		return                               // 回傳
// 	}
//
// 	go logger.Infof(formatString, args...) // 記錄資訊
//
// 	currentTime := time.Now()                  // 目前時間
// 	fileValidMaxDuration := time.Hour * 24 * 7 // 檔案最大有效時間
//
// 	for _, file := range files { // 針對每一個檔案
//
// 		if !file.IsDir() { // 若檔案不是資料夾，則輸出檔名和檔案內容
// 			fileName := file.Name()
// 			pathFileName := path + fileName
//
// 			if currentTime.Sub(file.ModTime()) > fileValidMaxDuration {
//
// 				go func() {
//
// 					osRemoveError := os.Remove(pathFileName)
//
// 					// 取得記錄器格式字串與參數
// 					formatString, args = logings.GetLogFuncFormatAndArguments(
// 						[]string{`%s %s 為 %s 準備資料: 刪除 %s 下存放超過 %v 的檔案 %s `},
// 						append(defaultArgs, fileValidMaxDuration, fileName),
// 						osRemoveError,
// 					)
//
// 					if nil != osRemoveError { // 若刪除檔案失敗
// 						logger.Warnf(formatString, args...) // 紀錄警告
// 					} else {
// 						logger.Infof(formatString, args...) // 紀錄資訊
// 					}
//
// 				}()
//
// 			} else {
//
// 				// networkHub.OutputStringToConnection(connectionPointer, fileName) // 輸出檔名
// 				// fileContentBytes, ioutilReadFileError := ioutil.ReadFile(pathFileName)
// 				//
// 				// // 取得記錄器格式字串與參數
// 				// formatString, args = logings.GetLogFuncFormatAndArguments(
// 				// 	[]string{`%s %s 為 %s 準備資料: 讀取 %s 下檔案 %s `},
// 				// 	append(defaultArgs, fileName),
// 				// 	ioutilReadFileError,
// 				// )
// 				//
// 				// if nil != ioutilReadFileError { // 若讀取檔案錯誤
// 				// 	logger.Errorf(formatString, args...) // 記錄錯誤
// 				// 	return                               // 回傳
// 				// }
// 				//
// 				// go logger.Infof(formatString, args...)                                  // 紀錄資訊
// 				// networkHub.OutputBytesToConnection(connectionPointer, fileContentBytes) // 輸出檔案內容
//
// 			}
//
// 		}
//
// 	} // end for 所有檔案
//
// 	networkHub.OutputStringToConnection(connectionPointer, ``)      // 輸出空字串表示結束
// 	networkHub.OutputBytesToConnection(connectionPointer, []byte{}) // 輸出空陣列表示結束
//
// }
