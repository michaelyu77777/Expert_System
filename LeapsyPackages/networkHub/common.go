package networkHub

import (
	"leapsy.com/packages/configurations"
	"leapsy.com/packages/logings"
)

var (
	logger      = logings.GetLogger()                                                      // 記錄器
	channelSize = configurations.GetConfigPositiveIntValueOrPanic(`local`, `channel-size`) // 取得預設通道大小                                                              // 讀寫鎖
)
