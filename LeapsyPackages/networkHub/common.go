package networkHub

import (
	"../configurations"
	"../logings"
)

var (
	logger      = logings.GetLogger()                                                      // 記錄器
	channelSize = configurations.GetConfigPositiveIntValueOrPanic(`local`, `channel-size`) // 取得預設通道大小                                                              // 讀寫鎖
)
