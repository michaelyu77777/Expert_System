package networkHub

import "github.com/gobwas/ws"

// websocketData - websocket傳遞資料
type websocketData struct {
	wsOpCode  ws.OpCode // 操作碼
	dataBytes []byte    // 資料位元組
}
