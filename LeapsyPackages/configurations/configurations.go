package configurations

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"

	"leapsy.com/packages/logings"
)

const (
	configFileNameConstString string = `config.ini` // 設定檔名
)

var (
	logger = logings.GetLogger() // 記錄器
)

var configMap map[string]map[string]string // 設定資料

// init - 初始函式
func init() {
	initializeConfigMapOrPanic() // 初始化設定資料或逐層結束程式
}

// initializeConfigMapOrPanic - 初始化設定資料或逐層結束程式
func initializeConfigMapOrPanic() {

	if nil == configMap { // 若沒有設定資料
		loadConfigFileOrPanic() // 載入設定檔或逐層結束程式
	}

}

type emptyStruct struct{} // 空結構

// loadConfigFileOrPanic - 載入設定檔或逐層結束程式
func loadConfigFileOrPanic() {

	configFile, iniLoadError := ini.Load(configFileNameConstString) // 載入設定檔

	formatStringItemSlices := []string{`載入設定檔 %s `}         // 記錄器格式片段
	defaultArgs := []interface{}{configFileNameConstString} // 記錄器預設參數

	// 取得記錄器格式字串與參數
	formatString, args := logings.GetLogFuncFormatAndArguments(
		[]string{strings.Join(formatStringItemSlices, ``)},
		defaultArgs,
		iniLoadError,
	)

	if nil != iniLoadError { // 若載入設定檔錯誤，則記錄錯誤
		logger.Panicf(formatString, args...) // 記錄錯誤
	} else { // 若載入設定檔成功，則儲存設定資料
		go logger.Infof(formatString, args...) // 記錄資訊

		configMap = make(map[string]map[string]string) // 位設定資料建立空間

		for _, section := range configFile.Sections() { // 針對每一個設定檔區塊

			configMap[section.Name()] = make(map[string]string) // 為設定檔區塊建立空間

			for _, key := range section.KeyStrings() { // 針對設定檔區塊下每一個關鍵字
				configMap[section.Name()][key] = section.Key(key).String() // 設定設定檔區塊下關鍵字對應的值
			}

		}

	}

}

// GetConfigValueOrPanic - 取得設定值否則結束程式
/**
 * @param  string sectionName  區塊名
 * @param  string key  關鍵字
 * @return string 設定資料區塊下關鍵字對應的值
 */
func GetConfigValueOrPanic(sectionName, key string) string {
	configValue, ok := configMap[sectionName][key] // 取得設定檔區塊下關鍵字對應的值

	if !ok { // 若取得設定檔區塊下關鍵字對應的值失敗

		formatStringItemSlices := []string{`設定檔 %s 設定`}         // 記錄器格式片段
		defaultArgs := []interface{}{configFileNameConstString} // 記錄器預設參數

		// 取得記錄器格式字串與參數
		formatString, args := logings.GetLogFuncFormatAndArguments(
			[]string{strings.Join(formatStringItemSlices, ``)},
			defaultArgs,
			fmt.Errorf(fmt.Sprintf(`[ %s ] %s`, sectionName, key)),
		)

		logger.Panicf(formatString, args...) // 記錄錯誤並逐層結束程式
	}

	return configValue // 回傳取得的設定檔區塊下關鍵字對應的值
}

// GetConfigPositiveIntValueOrPanic - 取得正整數設定值否則結束程式
/**
 * @param  string sectionName  區塊名
 * @param  string key  關鍵字
 * @return string 設定資料區塊下關鍵字對應的正整數值
 */
func GetConfigPositiveIntValueOrPanic(sectionName string, key string) int {

	value, _ := strconv.Atoi(GetConfigValueOrPanic(sectionName, key)) // 取得設定檔區塊下關鍵字對應的整數值

	if value <= 0 { // 若取得設定檔區塊下關鍵字對應的整數值非正整數
		formatStringItemSlices := []string{`設定檔 %s 設定`}         // 記錄器格式片段
		defaultArgs := []interface{}{configFileNameConstString} // 記錄器預設參數

		// 取得記錄器格式字串與參數
		formatString, args := logings.GetLogFuncFormatAndArguments(
			[]string{strings.Join(formatStringItemSlices, ``)},
			defaultArgs,
			fmt.Errorf(fmt.Sprintf(`[ %s ] %s 應為正整數`, sectionName, key)),
		)

		logger.Panicf(formatString, args...) // 記錄錯誤並逐層結束程式
	}

	return value // 回傳取得的設定檔區塊下關鍵字對應的正整數值
}
