package network

import (
	"net"
	"regexp"
	"strings"
	"sync"

	"leapsy.com/packages/logings"
)

var (
	logger = logings.GetLogger() // 記錄器

	addressToAliasMap = make(map[string]string) // 網址對應別名
	readWriteLock     = new(sync.RWMutex)       // 讀寫鎖
)

// LookupHostString - 查找主機
/**
 * @param  string hostString  主機
 * @return []string results 查找主機結果
 */
func LookupHostString(hostString string) (results []string) {

	// 主機正規表示式
	hostStringRegularExpression :=
		`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9]):\d{1,5}$`

	if regexp.MustCompile(hostStringRegularExpression).MatchString(hostString) { // 若主機符合格式

		hostSlices := strings.Split(hostString, `:`) // 將主機名稱切開

		formatStringSlices := []string{`查找主機 %s `} // 記錄器格式片段
		defaultArgs := []interface{}{hostString}   // 記錄器預設參數

		addrs, netLookupHostError := net.LookupHost(hostSlices[0]) // 查找主機

		// 取得記錄器格式和參數
		formatString, args := logings.GetLogFuncFormatAndArguments(
			formatStringSlices,
			defaultArgs,
			netLookupHostError,
		)

		if nil != netLookupHostError { // 若查找主機錯誤
			logger.Errorf(formatString, args...) // 記錄錯誤
			return                               // 回傳
		}

		go logger.Infof(formatString, args...) // 記錄資訊

		if nil != netLookupHostError { // 若查找主機錯誤
			logger.Errorf(formatString, args...) // 記錄錯誤
			return                               // 回傳
		}

		for _, addr := range addrs { // 針對每一結果

			updatedHostString := addr + `:` + hostSlices[1] // 結果加埠

			if _, ok := addressToAliasMap[updatedHostString]; ok { // 若結果加埠存在別名
				results = append(results, updatedHostString) // 將結果加埠加入回傳
			}

		}

		// 取得記錄器格式和參數
		formatString, args = logings.GetLogFuncFormatAndArguments(
			[]string{`取得主機 %s 資料 %v `},
			append(defaultArgs, results),
			nil,
		)

		logger.Infof(formatString, args...) // 記錄資訊

	} else { // 若主機不符合格式，則記錄資訊
		logger.Infof(`主機不符合格式"[hostname]:[port]": %s`, hostString)
	}

	return // 回傳
}

// GetAddressAlias - 取得位址別名
/**
 * @param  string addressString 位址字串
 * @return string 位址別名
 */
func GetAddressAlias(addressString string) string {
	readWriteLock.RLock()                   // 讀鎖
	defer readWriteLock.RUnlock()           // 記得解開讀鎖
	return addressToAliasMap[addressString] // 回傳位址別名
}

// SetAddressAlias - 設定位址別名
/**
 * @param  string addressString 位址字串
 * @param  string aliasString 別名字串
 */
func SetAddressAlias(addressString, aliasString string) {
	readWriteLock.Lock()                           // 寫鎖
	defer readWriteLock.Unlock()                   // 記得解開寫鎖
	addressToAliasMap[addressString] = aliasString // 設定位址別名
}

// GetAliasAddressPair - 取得(別名,位址)對
/**
 * @param  string addressString 位址字串
 * @return  []interface{} (別名,位址)對
 */
func GetAliasAddressPair(addressString string) []interface{} {
	return []interface{}{GetAddressAlias(addressString), addressString} // 回傳(別名,位址)對
}
