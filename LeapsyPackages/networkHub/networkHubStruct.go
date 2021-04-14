package networkHub

import (
	"errors"
	"sync"

	"../logings"
	"../network"
	"github.com/gobwas/ws"
)

// networkHub - 網路中心
type networkHub struct {
	broadcastChannel chan websocketData // 廣播通道

	clientsMutexPointer *sync.RWMutex    // 讀寫鎖指標                                                              // 讀寫鎖//
	clients             map[*client]bool // 客戶端

	connectChannel chan *client // 客戶端連接通道

	disconnectChannel chan *client // 客戶中斷連接通道
}

// initialize - 初始化
func (networkHubPointer *networkHub) initialize() {

	if nil == networkHubPointer.clientsMutexPointer { // 若沒讀寫鎖
		var clientsMutex sync.RWMutex                         // 讀寫鎖
		networkHubPointer.clientsMutexPointer = &clientsMutex // 儲存
	}

}

// setClientsOfKey - 設定客戶端值
/**
 * @param  *client key  關鍵字
 * @param  bool value  值
 */
func (networkHubPointer *networkHub) setClientsValueOfKey(key *client, value bool) {

	if nil != key { // 若指標不為空
		networkHubPointer.initialize()                 // 初始化
		networkHubPointer.clientsMutexPointer.Lock()   // 鎖寫
		networkHubPointer.clients[key] = value         // 客戶端
		networkHubPointer.clientsMutexPointer.Unlock() // 解鎖寫
	}

}

// getClients - 取得客戶端
/**
 * @param  map[*client]bool returnClients  客戶端
 */
func (networkHubPointer *networkHub) getClients() (returnClients map[*client]bool) {
	networkHubPointer.initialize()                  // 初始化
	networkHubPointer.clientsMutexPointer.RLock()   // 鎖讀
	returnClients = networkHubPointer.clients       // 客戶端
	networkHubPointer.clientsMutexPointer.RUnlock() // 解鎖讀

	return // 回傳
}

// getClientsValueAndOKOfKey - 取得客戶端值與ok
/**
 * @return  bool returnValue  值
 * @return  bool returnOK  OK
 */
func (networkHubPointer *networkHub) getClientsValueAndOKOfKey(key *client) (returnValue, returnOK bool) {

	if nil != key { // 若指標不為空

		networkHubPointer.initialize()                         // 初始化
		networkHubPointer.clientsMutexPointer.Lock()           // 鎖寫
		returnValue, returnOK = networkHubPointer.clients[key] // 客戶端
		networkHubPointer.clientsMutexPointer.Unlock()         // 解鎖寫

	}

	return
}

// deleteClientsKey - 刪除客戶端
/**
 * @param  *client key  關鍵字
 */
func (networkHubPointer *networkHub) deleteClientsKey(key *client) {

	if nil != key { // 若指標不為空

		networkHubPointer.initialize()                 // 初始化
		networkHubPointer.clientsMutexPointer.Lock()   // 鎖寫\
		delete(networkHubPointer.clients, key)         // 刪除客戶端
		networkHubPointer.clientsMutexPointer.Unlock() // 解鎖寫

	}

	return
}

var (
	networkHubPointer = newNetworkHub() // 創建網路中心
)

// init - 初始函式
func init() {
	go startNetworkHub() // 啟動網路中心
}

// newNetworkHub 建立新網路中心
func newNetworkHub() *networkHub {

	// 回傳建立的網路中心指標
	return &networkHub{
		broadcastChannel:  make(chan websocketData, channelSize),
		clients:           make(map[*client]bool),
		connectChannel:    make(chan *client, channelSize),
		disconnectChannel: make(chan *client, channelSize),
	}

}

// startNetworkHub - 啟動網路中心
func startNetworkHub() {

	if nil == networkHubPointer { // 若指標為空
		networkHubPointer = newNetworkHub() // 新建一個網路中心
	}

	for { // 循環處理通道接收

		select {

		case websocketData, ok := <-networkHubPointer.broadcastChannel: // 廣播通道收到資料

			wsOpCode := websocketData.wsOpCode
			dataBytes := websocketData.dataBytes

			formatSlice := `廣播通道接收`
			formatSlices := []string{formatSlice + `的資料為 %v `} // 紀錄器格式
			defaultArgs := []interface{}{dataBytes}            // 紀錄器預設參數

			if ws.OpText == wsOpCode {
				formatSlices = []string{formatSlice + `的字串為 %s `} // 紀錄器格式
				defaultArgs = []interface{}{string(dataBytes)}    // 紀錄器預設參數
			}

			var okError error // 接收錯誤

			if !ok { // 若廣播通道收到資料失敗，則建立接收錯誤
				okError = errors.New(`結束程式`) // 建立接收錯誤
			}

			// 取得紀錄器格式字串與參數
			formatString, args := logings.GetLogFuncFormatAndArguments(
				formatSlices,
				defaultArgs,
				okError,
			)

			if !ok { // 若客戶中斷連接通道接收客戶端失敗，則記錄錯誤
				logger.Errorf(formatString, args...) // 記錄錯誤
				return                               // 回傳
			}

			go logger.Infof(formatString, args...) // 紀錄資訊

			for client := range networkHubPointer.getClients() { // 針對每一個客戶端

				select {

				case client.outputChannel <- websocketData: // 傳資料給客戶端的輸出通道

				default:
					networkHubPointer.disconnectChannel <- client // 傳客戶端給客戶中斷連線通道
				} // end select

			} // end for

		case clientPointer, ok := <-networkHubPointer.connectChannel: // 若客戶端通道收到客戶輸入端

			connectionPointer := clientPointer.getConnectionPointer() // 連線指標

			if nil != connectionPointer { // 若連線指標不為空

				formatSlices := []string{`客戶通道 接收 客戶端 %v `}                         // 紀錄器格式
				defaultArgs := []interface{}{*clientPointer.getConnectionPointer()} // 紀錄器預設參數

				var okError error // 接收錯誤

				if !ok { // 若客戶輸入端通道收到客戶輸入端失敗，則建立接收錯誤
					okError = errors.New(`結束程式`) // 建立接收錯誤
				}

				// 取得紀錄器格式字串與參數
				formatString, args := logings.GetLogFuncFormatAndArguments(
					formatSlices,
					defaultArgs,
					okError,
				)

				if !ok { // 若客戶端通道收到客戶端失敗
					logger.Errorf(formatString, args...) // 記錄錯誤
					return                               // 回傳
				}

				logger.Infof(formatString, args...)

				networkHubPointer.setClientsValueOfKey(clientPointer, true) // 儲存客戶端

			}

		case clientPointer, ok := <-networkHubPointer.disconnectChannel: // 若客戶中斷連接通道收到客戶端

			connectionPointer := clientPointer.getConnectionPointer() // 連線指標

			if nil != connectionPointer { // 若連線指標不為空

				formatSlices := []string{`客戶中斷連接通道 接收 客戶端 %s `}                     // 紀錄器格式
				defaultArgs := []interface{}{*clientPointer.getConnectionPointer()} // 紀錄器預設參數

				var okError error // 接收錯誤

				if !ok { // 若客戶中斷連接通道接收客戶端失敗，則建立接收錯誤
					okError = errors.New(`結束程式`) // 建立接收錯誤
				}

				// 取得紀錄器格式字串與參數
				formatString, args := logings.GetLogFuncFormatAndArguments(
					formatSlices,
					defaultArgs,
					okError,
				)

				connectionPointer := clientPointer.getConnectionPointer() // 連線指標

				if !ok { // 若客戶中斷連接通道接收客戶端失敗，則記錄錯誤
					logger.Errorf(formatString, args...) // 記錄錯誤
					return                               // 回傳
				}

				connection := *connectionPointer // 客戶端的連線

				// 取得紀錄器格式字串與參數
				formatString, args = logings.GetLogFuncFormatAndArguments(
					[]string{`%s %s 中斷與客戶端 %s 的連線`},
					append(network.GetAliasAddressPair(connection.LocalAddr().String()), connection.RemoteAddr().String()),
					nil,
				)

				_, clientsOK := networkHubPointer.getClientsValueAndOKOfKey(clientPointer)

				if clientsOK { // 若客戶端存在

					networkHubPointer.deleteClientsKey(clientPointer) // 刪除客戶端
					clientPointer.setConnectionPointer(nil)           // 連線指標為空

					connection.Close() // 中斷客戶端連線

					go logger.Infof(formatString, args...) // 紀錄資訊

				}

			}

		} // end select

	} // end for

}

// broadcastHubWebsocketData - 廣播websocket資料
/**
 * @param  websocketData websocketData  websocket資料
 */
func broadcastHubWebsocketData(websocketData websocketData) {

	if nil != networkHubPointer { // 若指標不為空
		networkHubPointer.broadcastChannel <- websocketData // 傳資料給網路中心廣播通道
	}

}

// BroadcastHubString - 廣播字串
/**
 * @param  string inputString  輸入字串
 */
func BroadcastHubString(inputString string) {

	if nil != networkHubPointer { // 若指標不為空
		// 傳websocket資料給網路中心廣播通道
		networkHubPointer.broadcastChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: []byte(inputString)}
	}

}

// connectHub - 連接客戶端到網路中心
/**
 * @param  *client clientPointer  客戶端指標
 */
func connectHub(clientPointer *client) {

	if nil != clientPointer && nil != clientPointer.getConnectionPointer() && nil != networkHubPointer { // 若指標不為空
		networkHubPointer.connectChannel <- clientPointer // 傳客戶端給客戶輸入端通道
	}

}

// disconnectHub - 中斷客戶端與網路中心連線
/**
 * @param  *client clientPointer  客戶端指標
 */
func disconnectHub(clientPointer *client) {

	if nil != clientPointer && nil != networkHubPointer { // 若指標不為空
		networkHubPointer.disconnectChannel <- clientPointer // 傳客戶端給客戶中斷連接通道
	}

}
