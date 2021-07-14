package paths

import (
	"fmt"
	"os"
	"strings"

	"leapsy.com/packages/logings"
)

var (
	logger = logings.GetLogger() // 記錄器
)

// isPathNotExisted - 判斷路徑是否存在
/**
 * @param  string inputFileNameString  輸入檔名字串
 * @return bool  路徑是否存在
 */
func isFileNotExisted(inputFileNameString string) bool {

	_, osStatError := os.Stat(inputFileNameString) // 取得檔案資訊

	return os.IsNotExist(osStatError) // 回傳判斷是否為檔案不存在錯誤
}

// appendSlashIfNotEndWithOne - 若字串沒以"/"結尾則加"/"
/**
 * @param  string inputString  輸入字串
 * @return string outputString  輸出字串
 */
func appendSlashIfNotEndWithOne(inputString string) (outputString string) {

	if !strings.HasSuffix(inputString, `/`) { // 若輸入字串不以"/"結尾，則回傳輸入字串加"/"
		outputString = fmt.Sprintf(`%s/`, inputString) // 回傳輸入字串加"/"
	} else { // 若輸入字串以"/"結尾，則回傳輸入字串
		outputString = inputString // 回傳輸入字串
	}

	return // 回傳
}

// CreateIfPathNotExisted - 若路徑不存在則建立路徑
/**
 * @param  string inputPathString  輸入路徑字串
 */
func CreateIfPathNotExisted(inputPathString string) {

	pathString := appendSlashIfNotEndWithOne(inputPathString) // 若輸入路徑字串沒以"/"結尾則加"/"

	if isFileNotExisted(pathString) { // 若路徑不存在

		osMkdirAllError := os.MkdirAll(pathString, 0755) // 新增權限0755路徑檔案

		formatStringItemSlices := []string{`建立路徑 %s `} // 記錄器格式片段
		defaultArgs := []interface{}{pathString}       // 記錄器預設參數

		// 取得記錄器格式與參數
		formatString, args := logings.GetLogFuncFormatAndArguments(
			[]string{strings.Join(formatStringItemSlices, ``)},
			defaultArgs,
			osMkdirAllError,
		)

		if osMkdirAllError != nil { // 若新增路徑檔案錯誤，則記錄錯誤並逐層結束程式
			logger.Panicf(formatString, args...) // 記錄錯誤並逐層結束程式
		} else { // 若新增路徑檔案成功，則記錄資訊
			logger.Infof(formatString, args...) // 記錄資訊
		}
	}
}
