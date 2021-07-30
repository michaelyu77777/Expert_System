package networkHub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gobwas/ws"
	gomail "gopkg.in/gomail.v2"
	"leapsy.com/packages/model"
	"leapsy.com/packages/serverDataStruct"
	"leapsy.com/packages/serverResponseStruct"

	// "leapsy.com/packages/model"

	intersect "leapsy.com/packages/tools"
)

type CommandTool struct {
}

const (

	// 代碼-指令
	CommandNumberOfLogout             = 8  //登出
	CommandNumberOfBroadcastingInArea = 10 //區域廣播
	CommandNumberOfBroadcastingInRoom = 11 //房間廣播

	// 代碼-指令類型
	CommandTypeNumberOfAPI         = 1 // 客戶端-->Server
	CommandTypeNumberOfAPIResponse = 2 // Server-->客戶端
	CommandTypeNumberOfBroadcast   = 3 // Server廣播
	CommandTypeNumberOfHeartbeat   = 4 // 心跳包

	// 代碼-結果
	ResultCodeSuccess = 0 // 成功
	ResultCodeFail    = 1 // 失敗
)

// 匯入所有帳號到<帳號清單>中
// func (cTool *CommandTool) importAllAccountList() {

// 	// 取得資料庫帳號
// 	// accountsMongoDB := []model.Account{}
// 	// accountsMongoDB = mongoDB.FindAllAccounts()
// 	// for i, e := range accountsMongoDB {
// 	// 	fmt.Printf(`從資料庫取得 帳戶[ %d ] %+v `+"\n", i, e)
// 	// }

// 	picExpertA := cTool.getAccountPicString(accountPicPath + "picExpertA.txt")
// 	picExpertB := cTool.getAccountPicString(accountPicPath + "picExpertB.txt")
// 	picFrontline := cTool.getAccountPicString(accountPicPath + "picFrontline.txt")
// 	picDefault := cTool.getAccountPicString(accountPicPath + "picDefault.txt")

// 	//專家帳號 場域A
// 	accountExpertA := serverDataStruct.Account{
// 		UserID:               "expertA@leapsyworld.com",
// 		UserPassword:         "expertA@leapsyworld.com",
// 		UserName:             "專家-Adora",
// 		IsExpert:             1,
// 		IsFrontline:          2,
// 		Area:                 []int{1},
// 		AreaName:             []string{"場域A"},
// 		Pic:                  picExpertA,
// 		verificationCodeTime: time.Now().AddDate(1000, 0, 0), // 驗證碼永久有效時間1000年
// 	}
// 	//專家帳號 場域B
// 	accountExpertB := serverDataStruct.Account{
// 		UserID:               "expertB@leapsyworld.com",
// 		UserPassword:         "expertB@leapsyworld.com",
// 		UserName:             "專家-Belle",
// 		IsExpert:             1,
// 		IsFrontline:          2,
// 		Area:                 []int{2},
// 		AreaName:             []string{"場域B"},
// 		Pic:                  picExpertB,
// 		verificationCodeTime: time.Now().AddDate(1000, 0, 0), // 驗證碼永久有效時間1000年
// 	}

// 	//專家帳號 場域AB
// 	accountExpertAB := serverDataStruct.Account{
// 		UserID:               "expertAB@leapsyworld.com",
// 		UserPassword:         "expertAB@leapsyworld.com",
// 		UserName:             "專家-Abel",
// 		IsExpert:             1,
// 		IsFrontline:          2,
// 		Area:                 []int{1, 2},
// 		AreaName:             []string{"場域A", "場域B"},
// 		Pic:                  picExpertB,
// 		verificationCodeTime: time.Now().AddDate(1000, 0, 0), // 驗證碼永久有效時間1000年
// 	}

// 	//專家帳號 場域AB
// 	accountExpertPogo := serverDataStruct.Account{
// 		UserID:       "pogolin@leapsyworld.com",
// 		UserPassword: "pogolin@leapsyworld.com",
// 		UserName:     "專家-Pogo",
// 		IsExpert:     1,
// 		IsFrontline:  2,
// 		Area:         []int{1, 2},
// 		AreaName:     []string{"場域A", "場域B"},
// 		Pic:          picExpertB,
// 	}

// 	//專家帳號 場域AB
// 	accountExpertMichael := serverDataStruct.Account{
// 		UserID:       "michaelyu77777@gmail.com",
// 		UserPassword: "michaelyu77777@gmail.com",
// 		UserName:     "專家-Michael",
// 		IsExpert:     1,
// 		IsFrontline:  2,
// 		Area:         []int{1, 2},
// 		AreaName:     []string{"場域A", "場域B"},
// 		Pic:          picExpertB,
// 	}

// 	//一線人員帳號 匿名帳號
// 	defaultAccount := serverDataStruct.Account{
// 		UserID:       "default",
// 		UserPassword: "default",
// 		UserName:     "預設帳號",
// 		IsExpert:     2,
// 		IsFrontline:  1,
// 		Area:         []int{},
// 		AreaName:     []string{},
// 		Pic:          picDefault,
// 	}

// 	//一線人員帳號
// 	accountFrontLine := serverDataStruct.Account{
// 		UserID:       "frontLine@leapsyworld.com",
// 		UserPassword: "frontLine@leapsyworld.com",
// 		UserName:     "一線人員帳號",
// 		IsExpert:     2,
// 		IsFrontline:  1,
// 		Area:         []int{},
// 		AreaName:     []string{},
// 		Pic:          picFrontline,
// 	}

// 	//一線人員帳號2
// 	accountFrontLine2 := serverDataStruct.Account{
// 		UserID:       "frontLine2@leapsyworld.com",
// 		UserPassword: "frontLine2@leapsyworld.com",
// 		UserName:     "一線人員帳號",
// 		IsExpert:     2,
// 		IsFrontline:  1,
// 		Area:         []int{},
// 		AreaName:     []string{},
// 		Pic:          picFrontline,
// 	}

// 	allAccountPointerList = append(allAccountPointerList, &accountExpertA)
// 	allAccountPointerList = append(allAccountPointerList, &accountExpertB)
// 	allAccountPointerList = append(allAccountPointerList, &accountExpertAB)
// 	allAccountPointerList = append(allAccountPointerList, &accountExpertPogo)
// 	allAccountPointerList = append(allAccountPointerList, &accountExpertMichael)
// 	allAccountPointerList = append(allAccountPointerList, &defaultAccount)
// 	allAccountPointerList = append(allAccountPointerList, &accountFrontLine)
// 	allAccountPointerList = append(allAccountPointerList, &accountFrontLine2)
// }

// // 匯入所有裝置到<裝置清單>中
// func (cTool *CommandTool) importAllDevicesList() {

// 	// 待補:真的匯入資料庫所有裝置清單

// 	// devicesMongoDB := []model.Device{}
// 	// devicesMongoDB = mongoDB.FindAllDevices()
// 	// for i, e := range devicesMongoDB {
// 	// 	fmt.Printf(`從資料庫取得所有 裝置[ %d ] %+v `+"\n", i, e)
// 	// }

// 	// search001 := []model.Device{}
// 	// search001 = mongoDB.FindDevicesByDeviceIDAndDeviceBrand("001", "001")
// 	// for i, e := range search001 {
// 	// 	fmt.Printf(`從資料庫取得search001 裝置[ %d ] %+v `+"\n", i, e)
// 	// }

// 	search002 := []model.Device{}
// 	search002 = mongoDB.FindDevicesByDeviceIDAndDeviceBrand("002", "002")
// 	for i, e := range search002 {
// 		fmt.Printf(`從資料庫取得search002 裝置[ %d ] %+v `+"\n", i, e)
// 	}

// 	searchDeviceArea := []model.DeviceArea{}
// 	searchDeviceArea = mongoDB.FindDeviceAreasById(2)
// 	for i, e := range searchDeviceArea {
// 		fmt.Printf(`從資料庫取得searchDeviceArea 裝置[ %d ] %+v `+"\n", i, e)
// 	}

// 	searchDeviceType := []model.DeviceArea{}
// 	searchDeviceType = mongoDB.FindDeviceTypesById(2)
// 	for i, e := range searchDeviceType {
// 		fmt.Printf(`從資料庫取得searchDeviceType 裝置[ %d ] %+v `+"\n", i, e)
// 	}

// 	// 新增假資料：眼鏡假資料-場域A 眼鏡Model
// 	modelGlassesA := serverDataStruct.Device{
// 		DeviceID:     "",
// 		DeviceBrand:  "",
// 		DeviceType:   1,        //眼鏡
// 		Area:         []int{1}, // 依據裝置ID+Brand，從資料庫查詢
// 		AreaName:     []string{"場域A"},
// 		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
// 		Pic:          "",           // <求助>時才會從客戶端得到
// 		OnlineStatus: 2,            // 離線
// 		DeviceStatus: 0,            // 未設定
// 		MicStatus:    0,            // 未設定
// 		CameraStatus: 0,            // 未設定
// 		RoomID:       0,            // 無房間
// 	}

// 	// 新增假資料：場域B 眼鏡Model
// 	modelGlassesB := serverDataStruct.Device{
// 		DeviceID:     "",
// 		DeviceBrand:  "",
// 		DeviceType:   1,        //眼鏡
// 		Area:         []int{2}, // 依據裝置ID+Brand，從資料庫查詢
// 		AreaName:     []string{"場域B"},
// 		DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
// 		Pic:          "",           // <求助>時才會從客戶端得到
// 		OnlineStatus: 2,            // 離線
// 		DeviceStatus: 0,            // 未設定
// 		MicStatus:    0,            // 未設定
// 		CameraStatus: 0,            // 未設定
// 		RoomID:       0,            // 無房間
// 	}

// 	// 假資料：平板Model（沒有場域之分）
// 	modelTab := serverDataStruct.Device{
// 		DeviceID:     "",
// 		DeviceBrand:  "",
// 		DeviceType:   2,       // 平版
// 		Area:         []int{}, // 依據裝置ID+Brand，從資料庫查詢
// 		AreaName:     []string{},
// 		DeviceName:   "平板裝置", // 依據裝置ID+Brand，從資料庫查詢
// 		Pic:          "",     // <求助>時才會從客戶端得到
// 		OnlineStatus: 2,      // 離線
// 		DeviceStatus: 0,      // 未設定
// 		MicStatus:    0,      // 未設定
// 		CameraStatus: 0,      // 未設定
// 		RoomID:       0,      // 無房間
// 	}

// 	// 新增假資料：場域A 平版Model
// 	// modelTabA := serverDataStruct.Device{
// 	// 	DeviceID:     "",
// 	// 	DeviceBrand:  "",
// 	// 	DeviceType:   2,        // 平版
// 	// 	Area:         []int{1}, // 依據裝置ID+Brand，從資料庫查詢
// 	// 	AreaName:     []string{"場域A"},
// 	// 	DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
// 	// 	Pic:          "",           // <求助>時才會從客戶端得到
// 	// 	OnlineStatus: 2,            // 離線
// 	// 	DeviceStatus: 0,            // 未設定
// 	// 	MicStatus:    0,            // 未設定
// 	// 	CameraStatus: 0,            // 未設定
// 	// 	RoomID:       0,            // 無房間
// 	// }

// 	// // 新增假資料：場域B 平版Model
// 	// modelTabB := serverDataStruct.Device{
// 	// 	DeviceID:     "",
// 	// 	DeviceBrand:  "",
// 	// 	DeviceType:   2,        // 平版
// 	// 	Area:         []int{2}, // 依據裝置ID+Brand，從資料庫查詢
// 	// 	AreaName:     []string{"場域B"},
// 	// 	DeviceName:   "DeviceName", // 依據裝置ID+Brand，從資料庫查詢
// 	// 	Pic:          "",           // <求助>時才會從客戶端得到
// 	// 	OnlineStatus: 2,            // 離線
// 	// 	DeviceStatus: 0,            // 未設定
// 	// 	MicStatus:    0,            // 未設定
// 	// 	CameraStatus: 0,            // 未設定
// 	// 	RoomID:       0,            // 無房間
// 	// }

// 	// 新增假資料：場域A 眼鏡
// 	var glassesPointerA [5]*serverDataStruct.Device
// 	for i, devicePointer := range glassesPointerA {
// 		device := modelGlassesA
// 		devicePointer = &device
// 		devicePointer.DeviceID = "00" + strconv.Itoa(i+1)
// 		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+1)
// 		glassesPointerA[i] = devicePointer
// 	}
// 	fmt.Printf("假資料眼鏡A=%+v\n", glassesPointerA)

// 	// 場域B 眼鏡
// 	var glassesPointerB [5]*serverDataStruct.Device
// 	for i, devicePointer := range glassesPointerB {
// 		device := modelGlassesB
// 		devicePointer = &device
// 		devicePointer.DeviceID = "00" + strconv.Itoa(i+6)
// 		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+6)
// 		glassesPointerB[i] = devicePointer
// 	}
// 	fmt.Printf("假資料眼鏡B=%+v\n", glassesPointerB)

// 	// 平版-0011
// 	var tabsPointerA [1]*serverDataStruct.Device
// 	for i, devicePointer := range tabsPointerA {
// 		device := modelTab
// 		devicePointer = &device
// 		devicePointer.DeviceID = "00" + strconv.Itoa(i+11)
// 		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+11)
// 		devicePointer.DeviceName = "平板00" + strconv.Itoa(i+11)
// 		tabsPointerA[i] = devicePointer
// 	}
// 	fmt.Printf("假資料平版-0011=%+v\n", tabsPointerA)

// 	// 平版-0012
// 	var tabsPointerB [1]*serverDataStruct.Device
// 	for i, devicePointer := range tabsPointerB {
// 		device := modelTab
// 		devicePointer = &device
// 		devicePointer.DeviceID = "00" + strconv.Itoa(i+12)
// 		devicePointer.DeviceBrand = "00" + strconv.Itoa(i+12)
// 		devicePointer.DeviceName = "平板00" + strconv.Itoa(i+12)
// 		tabsPointerB[i] = devicePointer
// 	}
// 	fmt.Printf("假資料平版-0012=%+v\n", tabsPointerB)

// 	// 加入眼鏡A
// 	for _, e := range glassesPointerA {
// 		allDevicePointerList = append(allDevicePointerList, e)
// 	}
// 	// 加入眼鏡B
// 	for _, e := range glassesPointerB {
// 		allDevicePointerList = append(allDevicePointerList, e)
// 	}
// 	// 加入平版A
// 	for _, e := range tabsPointerA {
// 		allDevicePointerList = append(allDevicePointerList, e)
// 	}
// 	// 加入平版B
// 	for _, e := range tabsPointerB {
// 		allDevicePointerList = append(allDevicePointerList, e)
// 	}

// 	// fmt.Printf("\n取得所有裝置%+v\n", AllDeviceList)
// 	for _, e := range allDevicePointerList {
// 		fmt.Printf("資料庫所有裝置清單:%+v\n", e)
// 	}

// }

// // 匯入所有場域對應名稱
// func (cTool *CommandTool) importAllAreasNameToMap() {
// 	areaNumberNameMap[1] = "場域A"
// 	areaNumberNameMap[2] = "場域B"
// }

// 登入(並處理重複登入問題)
/**
 * @param string whatKindCommandString 呼叫此函數的指令名稱
 * @param *client clientPointer 連線指標
 * @param serverResponseStruct.serverResponseStruct.Command command 客戶端傳來的指令
 * @param *serverDataStruct.Device newDevicePointer 登入之新裝置指標
 * @param *serverDataStruct.Account newAccountPointer 欲登入之新帳戶指標
 * @return bool isSuccess 回傳是否成功
 * @return string otherMessage 回傳詳細訊息
 */
func (cTool *CommandTool) processLoginWithDuplicate(whatKindCommandString string, clientPointer *client, command serverResponseStruct.Command, newDevicePointer *serverDataStruct.Device, newAccountPointer *serverDataStruct.Account) (isSuccess bool, messages string) {

	// 建立Info
	newInfoPointer := serverDataStruct.Info{
		AccountPointer: newAccountPointer,
		DevicePointer:  newDevicePointer,
	}

	// 登入步驟: 四種況狀判斷：相同連線、相異連線、不同裝置、相同裝置 重複登入之處理
	if infoPointer, ok := clientInfoMap[clientPointer]; ok {
		//相同連線
		messages += "-相同連線"

		//檢查裝置
		devicePointer := infoPointer.DevicePointer

		if devicePointer != nil {

			if (devicePointer.DeviceID == command.DeviceID) && (devicePointer.DeviceBrand == command.DeviceBrand) {
				// 裝置相同：同裝置重複登入
				messages += "-相同裝置"

				// 設定info
				infoPointer = &newInfoPointer

				devicePointer.OnlineStatus = 1 // 狀態為上線
				devicePointer.DeviceStatus = 1 // 裝置變閒置

			} else {
				// 裝置不同（現實中不會出現，只有測試才會出現）
				messages += "-不同裝置"

				// 將舊的裝置離線，並重設舊的裝置狀態
				if isSuccess, errMsg := cTool.resetDevicePointerStatus(devicePointer); !isSuccess {

					messages += "-重設舊裝置狀態失敗"

					//重設失敗
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
					cTool.processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
					return false, (messages + errMsg)
				}

				// 設定新info
				// 檢查
				if nil != newInfoPointer.DevicePointer {

					newInfoPointer.DevicePointer.OnlineStatus = 1 // 新的裝置＝上線
					newInfoPointer.DevicePointer.DeviceStatus = 1 // 裝置變閒置

					// 重要！將new info 指回clientInfoMap
					clientInfoMap[clientPointer] = &newInfoPointer

					//不需要斷線

				} else {
					//找不到新裝置
					messages += "-找不到新裝置"
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
					cTool.processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
					return false, messages
				}
			}
		} else {
			//失敗:找不到舊裝置
			messages += "-找不到舊裝置"
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
			return false, messages
		}

	} else {
		// 不同連線
		messages += `-發現不同連線`

		// 去找Map中有沒有已經存在相同裝置
		isExist, oldClient := cTool.isDeviceExistInClientInfoMap(newDevicePointer)

		if isExist {
			messages += `-相同裝置`
			// 裝置相同：（現實中，只有實體裝置重複ID才會實現）

			// 舊的連線：斷線＋從MAP移除＋Response舊的連線即將斷線
			if success, otherMessage := cTool.processDisconnectAndResponse(oldClient); success {
				otherMessage += messages + `-將重複裝置斷線`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, otherMessage, command, clientPointer) //所有值複製一份做logger
				cTool.processLoggerWarnf(whatKindCommandString, otherMessage, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
			} else {
				// 失敗:無此狀況
				return false, ""
			}

			// 新的連線，加入到Map，並且對應到新的裝置與帳號
			clientInfoMap[clientPointer] = &newInfoPointer
			// fmt.Printf("找到重複的連線，從Map中刪除，將此Socket斷線。\n")

			//檢查裝置
			devicePointer := clientInfoMap[clientPointer].DevicePointer

			if devicePointer != nil {
				devicePointer.OnlineStatus = 1 // 狀態為上線
				devicePointer.DeviceStatus = 1 // 裝置變閒置

			} else {
				//裝置為空

				// 警告logger
				messages += `-找不到此裝置`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
				cTool.processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				return false, messages

			}

		} else {
			//裝置不同：正常新增一的新裝置
			messages += `-不同裝置`

			clientInfoMap[clientPointer] = &newInfoPointer

			//檢查裝置
			devicePointer := clientInfoMap[clientPointer].DevicePointer
			if devicePointer != nil {

				devicePointer.OnlineStatus = 1 // 裝置狀態＝線上
				devicePointer.DeviceStatus = 1 // 裝置變閒置

			} else {
				//裝置為空

				// 警告logger
				messages += `-找不到此裝置`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, messages, command, clientPointer) //所有值複製一份做logger
				cTool.processLoggerWarnf(whatKindCommandString, messages, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				return false, messages
			}
		}
	}
	return true, messages
}

// 重設裝置狀態為預設狀態
/**
 * @param devicePointer *serverDataStruct.Device 裝置指標(想要重設的裝置)
 * @return isSuccess bool 回傳是否成功
 * @return messages string 回傳詳細訊息
 */
func (cTool *CommandTool) resetDevicePointerStatus(devicePointer *serverDataStruct.Device) (isSuccess bool, messages string) {

	// 檢查裝置指標
	if nil != devicePointer {
		//成功
		devicePointer.OnlineStatus = 2
		devicePointer.DeviceStatus = 0
		devicePointer.CameraStatus = 0
		devicePointer.MicStatus = 0
		devicePointer.Pic = ""
		devicePointer.RoomID = 0
		return true, ``
	} else {
		//若找不到裝置指標
		return false, `-找不到裝置`
	}

}

// 設置裝置為離線
/**
 * @param devicePointer *serverDataStruct.Device 裝置指標(想要設置離線的裝置)
 * @return isSuccess bool 回傳是否成功
 * @return messages string 回傳詳細訊息
 */
func (cTool *CommandTool) setDevicePointerOffline(devicePointer *serverDataStruct.Device) (isSuccess bool, messages string) {

	// 檢查裝置指標
	if nil != devicePointer {
		//成功
		devicePointer.OnlineStatus = 2
		devicePointer.DeviceStatus = 0
		devicePointer.CameraStatus = 0
		devicePointer.MicStatus = 0
		devicePointer.Pic = ""
		devicePointer.RoomID = 0
		return true, ``
	} else {
		//若找不到裝置指標
		return false, `-找不到裝置`
	}
}

// 將某連線斷線，並Response此連線
/**
 * @param clientPointer *client 連線指標(想要斷線的連線)
 * @return isSuccess bool 回傳是否成功
 * @return messages string 回傳詳細訊息
 */
func (cTool *CommandTool) processDisconnectAndResponse(clientPointer *client) (isSuccess bool, messages string) {

	// 告知舊的連線，即將斷線
	// (讓logger可以進行平行處理，怕尚未執行到，就先刪掉了連線與裝置，就無法印出了)

	// Response:被斷線的連線:有裝置重複登入，已斷線
	details := `已斷線，有其他相同裝置ID登入伺服器`
	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, CommandNumberOfLogout, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, ""))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} // Response

	// logger:此斷線裝置的訊息
	details += `-此帳號與裝置為被斷線的連線裝置`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(``, details, serverResponseStruct.Command{}, clientPointer) //所有值複製一份做logger
	cTool.processLoggerWarnf(``, details, serverResponseStruct.Command{}, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	fmt.Println("測試:已經進行斷線Response")

	// 舊的連線，從Map移除
	delete(clientInfoMap, clientPointer) // 此連線從Map刪除

	// 舊的連線，進行斷線
	disconnectHub(clientPointer) // 此連線斷線

	return true, `已將指定連線斷線`
}

// 查詢在線清單中(ClientInfoMap)，是否已經有相同裝置存在
/**
 * @param myDevicePointer *serverDataStruct.Device 裝置指標(想找的裝置)
 * @return bool 回傳結果(存在/不存在)
 * @return *client 回傳找到存在的連線指標(找不到則回傳nil)
 */
func (cTool *CommandTool) isDeviceExistInClientInfoMap(myDevicePointer *serverDataStruct.Device) (bool, *client) {

	for clientPointer, infoPointer := range clientInfoMap {
		// 若找到相同裝置，回傳連線

		// 檢查傳入的裝置
		if myDevicePointer != nil {

			//檢查info
			if infoPointer != nil {

				//檢查裝置
				if infoPointer.DevicePointer != nil {

					if infoPointer.DevicePointer.DeviceID == myDevicePointer.DeviceID &&
						infoPointer.DevicePointer.DeviceBrand == myDevicePointer.DeviceBrand {
						return true, clientPointer
					}

				} else {
					//裝置為空,不額外處理
					return false, nil
				}
			} else {
				//info 為空,不額外處理
				return false, nil
			}
		} else {
			//傳入的裝置為空,不額外處理
			return false, nil
		}
	}
	return false, nil
}

// 根據裝置找到連線，關閉連線(排除自己這條連線)
// func findClientByDeviceAndCloseSocket(device *serverDataStruct.Device, excluder *client) {

// 	for client, e := range clientInfoMap {
// 		// 若找到相同裝置，關閉此client連線
// 		if e.DevicePointer.DeviceID == device.DeviceID && e.DevicePointer.DeviceBrand == device.DeviceBrand && client != excluder {
// 			delete(clientInfoMap, client) // 此連線從Map刪除
// 			disconnectHub(client)         // 此連線斷線
// 			fmt.Printf("找到重複的連線，從Map中刪除，將此Socket斷線。\n")
// 		}
// 	}

// }

// 取得帳號，透過UserID(加密的時候使用)
/**
 * @param userID string 想查找的UserID
 * @return accountPointer *serverDataStruct.Account 回傳找到的帳號指標
 */
// func (cTool *CommandTool) getAccountByUserID(userID string) (accountPointer *serverDataStruct.Account) {

// 	for _, accountPointer := range allAccountPointerList {
// 		// 檢查帳號
// 		if nil != accountPointer {
// 			// 若找到，直接返回帳號指標
// 			if userID == accountPointer.UserID {
// 				return accountPointer
// 			}
// 		}
// 	}

// 	// 沒找到帳號，直接回一個空的
// 	accountPointer = &Account{}
// 	return
// }

// 確認密碼是否正確
/**
 * @param userID string 帳號
 * @param userPassword string 密碼
 * @return bool 回傳是否正確
 * @return *serverDataStruct.Account 回傳找到的帳號資料
 */
func (cTool *CommandTool) checkPasswordAndGetAccountPointer(userID string, userPassword string) (bool, *serverDataStruct.Account) {

	// 新版:資料庫版本
	result := []model.Account{}
	result = mongoDB.FindAccountByUserID(userID)
	for i, e := range result {
		fmt.Printf(`查到 - Account[%d] %+v `+"\n", i, e)
	}

	// 結果不為空
	// if result != nil {
	if len(result) > 0 {

		accountPointer := &serverDataStruct.Account{
			UserID:       result[0].UserID,
			UserPassword: result[0].UserPassword,
			UserName:     result[0].UserName,
			IsExpert:     result[0].IsExpert,
			IsFrontline:  result[0].IsFrontline,
			Area:         result[0].Area,
			// AreaName:     , 等待設定
			Pic: result[0].Pic,
			// verificationCodeTime: result[0].VerificationCodeTime,
			VerificationCodeTime: result[0].VerificationCodeTime,
		}

		//若為demo模式,且為測試帳號直接通過
		if (1 == expertdemoMode) &&
			("expertA@leapsyworld.com" == userID ||
				"expertB@leapsyworld.com" == userID ||
				"expertAB@leapsyworld.com" == userID) {
			return true, accountPointer
		} else {
			//非測試帳號 驗證密碼
			if userPassword == accountPointer.UserPassword {
				return true, accountPointer
			} else {
				return false, nil
			}
		}

	}

	return false, nil

	// 舊版:軟體版
	// for _, accountPointer := range allAccountPointerList {

	// 	// 帳號不為空
	// 	if accountPointer != nil {
	// 		if userID == accountPointer.UserID {

	// 			//若為demo模式,且為測試帳號直接通過
	// 			if (1 == expertdemoMode) &&
	// 				("expertA@leapsyworld.com" == userID || "expertB@leapsyworld.com" == userID || "expertAB@leapsyworld.com" == userID) {
	// 				return true, accountPointer
	// 			} else {
	// 				//非測試帳號 驗證密碼
	// 				if userPassword == accountPointer.UserPassword {
	// 					return true, accountPointer
	// 				} else {
	// 					return false, nil
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// return false, nil
}

// 判斷某連線是否已登入指令(若沒有登入，則直接RESPONSE給客戶端，並說明因尚未登入執行失敗)
/**
 * @param clientPointer *client 連線
 * @param command serverResponseStruct.Command 客戶端傳來的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @return isLogedIn bool 回傳是否已登入
 */
func (cTool *CommandTool) checkLogedInAndResponseIfFail(clientPointer *client, command serverResponseStruct.Command, whatKindCommandString string) (isLogedIn bool) {

	// 若登入過
	if _, ok := clientInfoMap[clientPointer]; ok {

		isLogedIn = true
		return

	} else {
		// 尚未登入
		isLogedIn = false
		details := `-執行失敗，連線尚未登入`

		// 失敗:Response
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

		// 警告logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return
	}
}

// 判斷某連線裝置是否為閒置(判斷依據:eviceStatus)
/**
 * @param client *client 連線
 * @param command serverResponseStruct.Command 客戶端傳來的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @param details string 之前已經處理的細節
 * @return bool 回傳是否為閒置
 */
func (cTool *CommandTool) checkDeviceStatusIsIdleAndResponseIfFail(client *client, command serverResponseStruct.Command, whatKindCommandString string, details string) bool {

	// 若連線存在
	if e, ok := clientInfoMap[client]; ok {
		details += `-找到連線`

		//檢查裝置
		devicePointer := e.DevicePointer
		if devicePointer != nil {
			details += `-找到裝置`

			// 若為閒置
			if devicePointer.DeviceStatus == 1 {
				// 狀態為閒置
				return true
			} else {
				// 狀態非閒置
				details += `-裝置狀態非閒置`

				// 失敗:Response
				jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
				client.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

				// logger
				details += `-執行指令失敗`
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, client) //所有值複製一份做logger
				cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return false
			}
		} else {
			// 裝置不存在
			// 失敗:Response
			details += `-找不到裝置`
			cTool.processResponseDeviceNil(client, whatKindCommandString, command, details)
			return false

		}

	} else {
		// 此連線不存在
		details += `-找不到連線`
		cTool.processResponseInfoNil(client, whatKindCommandString, command, details)
		return false

	}
}

// 判斷此裝置是否為眼鏡端(若不是的話，則直接RESPONSE給客戶端，並說明無法切換場域)
/**
 * @param clientPointer *client 連線指標
 * @param command serverResponseStruct.Command 客戶端傳來的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @param details string 之前已經處理的細節
 * @return 是否為眼鏡端
 */
func (cTool *CommandTool) checkDeviceTypeIsGlassesAndResponseIfFail(clientPointer *client, command serverResponseStruct.Command, whatKindCommandString string, details string) bool {

	// 取連線
	if infoPointer, ok := clientInfoMap[clientPointer]; ok {
		details += "-找到連線"

		// 取裝置
		devicePointer := infoPointer.DevicePointer
		if devicePointer != nil {
			details += "-找到裝置"

			// 若為眼鏡端
			if devicePointer.DeviceType == 1 {
				// 成功
				return true
			} else {
				// 失敗Response:非眼鏡端
				details += "-非眼鏡端無法切換區域"

				// 失敗:Response
				jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
				clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes} //Socket Response

				// logger
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
				cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return false
			}
		} else {
			// 失敗Response:找不到裝置
			details += "-找不到裝置"
			cTool.processResponseDeviceNil(clientPointer, whatKindCommandString, command, details)
			return false
		}

	} else {
		//失敗:連線不存在
		details += "-找不到連線"
		cTool.processResponseInfoNil(clientPointer, whatKindCommandString, command, details)
		return false
	}
}

// 判斷某連線是否已登入(用連線存不存在決定)
/**
 * @param client *client 連線指標
 * @return bool 回傳結果
 */
func (cTool *CommandTool) checkLogedInByClient(client *client) bool {

	logedIn := false

	if _, ok := clientInfoMap[client]; ok {
		logedIn = true
	}

	return logedIn

}

// // 從清單移除某裝置
// func removeDeviceFromList(slice []*serverDataStruct.Device, s int) []*serverDataStruct.Device {
// 	return append(slice[:s], slice[s+1:]...) //回傳移除後的array
// }

// // 取得所有裝置清單 By clientInfoMap（For Logger）
// func getOnlineDevicesByClientInfoMap() []serverDataStruct.Device {

// 	deviceArray := []serverDataStruct.Device{}

// 	for _, info := range clientInfoMap {
// 		device := info.DevicePointer
// 		deviceArray = append(deviceArray, *device)
// 	}

// 	return deviceArray
// }

// 取得所有匯入的裝置清單COPY副本(For Logger)
func (cTool *CommandTool) getAllDeviceByList() []serverDataStruct.Device {
	deviceArray := []serverDataStruct.Device{}

	// 改成透過clientInfoMap來取得
	for _, infoPointer := range clientInfoMap {
		devicePointer := infoPointer.DevicePointer
		if nil != devicePointer {
			deviceArray = append(deviceArray, *devicePointer)
		}
	}

	// for _, d := range allDevicePointerList {
	// 	deviceArray = append(deviceArray, *d)
	// }

	return deviceArray
}

func (cTool *CommandTool) getDevicePointerBySearchAndCreate(deviceID string, deviceBrand string) (devicePointer *serverDataStruct.Device) {

	fmt.Printf("建立DevicePointer,deviceID= %s, deviceBrand=%s。", deviceID, deviceBrand)

	result := []model.Device{}
	result = mongoDB.FindDevicesByDeviceIDAndDeviceBrand(deviceID, deviceBrand)

	//有找到
	if len(result) > 0 {

		devicePointer = &serverDataStruct.Device{
			DeviceID:     result[0].DeviceID,
			DeviceBrand:  result[0].DeviceBrand,
			DeviceType:   result[0].DeviceType,
			Area:         result[0].Area,
			AreaName:     cTool.getAreaNameArrayByAreaIDAarray(result[0].Area),
			DeviceName:   result[0].DeviceName,
			Pic:          "", // 等待客戶端設定
			OnlineStatus: 2,
			DeviceStatus: 0,
			CameraStatus: 0,
			MicStatus:    0,
			RoomID:       0,
		}

		return
	} else {
		return nil
	}

}

// 取得相對應的 AreadName Array
func (cTool *CommandTool) getAreaNameArrayByAreaIDAarray(idArray []int) (nameArray []string) {

	fmt.Printf("取得相對應 AreadName Array,idArray=%+v。", idArray)

	result := []model.Area{}

	for _, id := range idArray {
		result = mongoDB.FindAreaById(id)
		if len(result) > 0 {
			nameArray = append(nameArray, result[0].Zhtw)
		}
	}

	return

}

// 取得某裝置指標
/**
 * @param deviceID string 裝置ID
 * @param deviceBrand string 裝置Brand
 * @return result *serverDataStruct.Device 回傳裝置指標
 */
func (cTool *CommandTool) getDevicePointer(deviceID string, deviceBrand string) (result *serverDataStruct.Device) {

	// 改成透過clientInfoMap來找了
	for _, infoPointer := range clientInfoMap {
		if nil != infoPointer {
			devicePointer := infoPointer.DevicePointer
			if nil != devicePointer {
				if devicePointer.DeviceID == deviceID {
					if devicePointer.DeviceBrand == deviceBrand {
						result = devicePointer
						// fmt.Println("找到裝置", result)
						// fmt.Println("所有裝置清單", allDevicePointerList)
					}
				}
			}

		}
	}

	// 若找到則返回
	// for _, devicePointer := range allDevicePointerList {
	// 	if nil != devicePointer {
	// 		if devicePointer.DeviceID == deviceID {
	// 			if devicePointer.DeviceBrand == deviceBrand {
	// 				result = devicePointer
	// 				// fmt.Println("找到裝置", result)
	// 				// fmt.Println("所有裝置清單", allDevicePointerList)
	// 			}
	// 		}
	// 	}
	// }

	return // 回傳
}

// // 取得裝置:同區域＋同類型＋去掉某一裝置（自己）
// func getDevicesByAreaAndDeviceTypeExeptOneDevice(area []int, deviceType int, device *serverDataStruct.Device) []*serverDataStruct.Device {

// 	result := []*serverDataStruct.Device{}

// 	// 若找到則返回
// 	for _, e := range allDevicePointerList {
// 		intersection := intersect.Hash(e.Area, area) //取交集array

// 		// 同區域＋同類型＋去掉某一裝置（自己）
// 		if len(intersection) > 0 && e.DeviceType == deviceType && e != device {

// 			result = append(result, e)

// 		}
// 	}

// 	fmt.Printf("找到指定場域＋指定類型＋排除自己的所有裝置:%+v \n", result)
// 	return result // 回傳
// }

// 針對<專家＋平版端>，組合出所有連線指標InfoPointer(裝置+帳號組合):符合與專家同場域(某場域)＋裝置為眼鏡(某裝置類型)＋去掉自己(某一裝置)
/**
 * @param myArea []int 想查的場域
 * @param someDeviceType int 想查的裝置類型(眼鏡/平板)
 * @param myDevice *serverDataStruct.Device 排除廣播的自己裝置(眼鏡/平板)
 * @return resultInfoPointers []*Info 回傳組合IngoPointer的結果
 * @return otherMeessage string 回傳詳細狀況
 */
func (cTool *CommandTool) getDevicesWithInfoByAreaAndDeviceTypeExeptOneDevice(myArea []int, someDeviceType int, myDevice *serverDataStruct.Device) (resultInfoPointers []*serverDataStruct.Info, otherMeessage string) {

	fmt.Printf("測試：目標要找 myArea =%+v", myArea)

	// 改成透過clientInfoMap去找
	for _, infoPointer := range clientInfoMap {
		devicePointer := infoPointer.DevicePointer
		if nil != devicePointer {
			// 找到裝置
			otherMeessage += "-找到裝置"

			//intersection := []int{1, 2}

			intersection := intersect.Hash(devicePointer.Area, myArea)
			// intersection := intersect.Hash(devicePointer.Area, myArea) //場域交集array
			fmt.Printf("\n\n 找交集 intersection =%+v, device=%s", intersection, devicePointer.DeviceID)

			// 有相同場域 + 同類型 + 去掉某一裝置（自己）
			if (len(intersection) > 0) && (devicePointer.DeviceType == someDeviceType) && (devicePointer != myDevice) {

				otherMeessage += "-找到相同場域"
				fmt.Printf("\n\n 找到交集 intersection =%+v, device=%s", intersection, devicePointer.DeviceID)

				// 準備進行同場域info包裝，針對空Account進行處理

				// 裝置在線，取出info
				if 1 == devicePointer.OnlineStatus {
					infoPointer := cTool.getInfoByOnlineDevice(devicePointer)

					//若有找到則加入結果清單
					if nil != infoPointer {
						resultInfoPointers = append(resultInfoPointers, infoPointer) // 加到結果
						fmt.Printf("\n\n 找到在線裝置=%+v,帳號=%+v", devicePointer, infoPointer.AccountPointer)
					}

				} else {
					//裝置離線，給空的Account
					emptyAccountPointer := &serverDataStruct.Account{}
					infoPointer := &serverDataStruct.Info{AccountPointer: emptyAccountPointer, DevicePointer: devicePointer}
					resultInfoPointers = append(resultInfoPointers, infoPointer) // 加到結果
					fmt.Printf("\n\n 找到離線裝置=%+v,帳號=%+v", devicePointer, infoPointer.AccountPointer)
				}
			}
		} else {
			// 裝置為空 不做事
			otherMeessage += "-裝置為空"
		}

	}

	// 若找到則返回
	// for _, devicePointer := range allDevicePointerList {

	// 	if nil != devicePointer {
	// 		// 找到裝置
	// 		otherMeessage += "-找到裝置"

	// 		//intersection := []int{1, 2}

	// 		intersection := intersect.Hash(devicePointer.Area, myArea)
	// 		// intersection := intersect.Hash(devicePointer.Area, myArea) //場域交集array
	// 		fmt.Printf("\n\n 找交集 intersection =%+v, device=%s", intersection, devicePointer.DeviceID)

	// 		// 有相同場域 + 同類型 + 去掉某一裝置（自己）
	// 		if (len(intersection) > 0) && (devicePointer.DeviceType == someDeviceType) && (devicePointer != myDevice) {

	// 			otherMeessage += "-找到相同場域"
	// 			fmt.Printf("\n\n 找到交集 intersection =%+v, device=%s", intersection, devicePointer.DeviceID)

	// 			// 準備進行同場域info包裝，針對空Account進行處理

	// 			// 裝置在線，取出info
	// 			if 1 == devicePointer.OnlineStatus {
	// 				infoPointer := cTool.getInfoByOnlineDevice(devicePointer)

	// 				//若有找到則加入結果清單
	// 				if nil != infoPointer {
	// 					resultInfoPointers = append(resultInfoPointers, infoPointer) // 加到結果
	// 					fmt.Printf("\n\n 找到在線裝置=%+v,帳號=%+v", devicePointer, infoPointer.AccountPointer)
	// 				}

	// 			} else {
	// 				//裝置離線，給空的Account
	// 				emptyAccountPointer := &serverDataStruct.Account{}
	// 				infoPointer := &Info{AccountPointer: emptyAccountPointer, DevicePointer: devicePointer}
	// 				resultInfoPointers = append(resultInfoPointers, infoPointer) // 加到結果
	// 				fmt.Printf("\n\n 找到離線裝置=%+v,帳號=%+v", devicePointer, infoPointer.AccountPointer)
	// 			}
	// 		}
	// 	} else {
	// 		// 裝置為空 不做事
	// 		otherMeessage += "-裝置為空"
	// 	}

	// }

	fmt.Printf("找到指定場域＋指定裝置類型＋排除自己的所有裝置，結果為:%+v \n", resultInfoPointers)
	return // 回傳
}

// 取得在線裝置所對應的Info
/**
 * @param devicePointer *serverDataStruct.Device 某在線裝置指標
 * @return *Info 回傳對應的連線指標
 */
func (cTool *CommandTool) getInfoByOnlineDevice(devicePointer *serverDataStruct.Device) *serverDataStruct.Info {

	for _, infoPointer := range clientInfoMap {

		// 若找到裝置對應的info
		if devicePointer == infoPointer.DevicePointer {
			return infoPointer
		}
	}

	//找不到就回傳NIL
	return nil

	//找不到就回傳空的
	//return &Info{}
}

// // 排除某連線進行廣播 (excluder 被排除的client)
// func broadcastExceptOne(excluder *client, websocketData websocketData) {

// 	//Response to all
// 	for clientPointer, _ := range clientInfoMap {

// 		// 僅排除一個連線
// 		if nil != clientPointer && clientPointer != excluder {

// 			clientPointer.outputChannel <- websocketData //Socket Response

// 		}
// 	}
// }

// // 針對指定群組進行廣播，排除某連線(自己)
// func broadcastByGroup(clientPinters []*client, websocketData websocketData, excluder *client) {

// 	for _, clientPointer := range clientPinters {

// 		//排除自己
// 		if nil != clientPointer && clientPointer != excluder {
// 			clientPointer.outputChannel <- websocketData //Socket Respone
// 		}
// 	}
// }

// 針對某場域(Area)進行廣播，排除某連線(自己)
/**
 * @param area []int 想廣播的區域代碼
 * @param websocketData websocketData 想廣播的內容
 * @param whatKindCommandString string 是哪個指令呼叫此函式
 * @param command Command 客戶端的指令
 * @param excluder *client 排除廣播的客戶端連線指標(通常是自己)
 * @param details string 詳細資訊
 */
func (cTool *CommandTool) broadcastByArea(area []int, websocketData websocketData, whatKindCommandString string, command serverResponseStruct.Command, excluder *client, details string) {

	for clientPointer, infoPointer := range clientInfoMap {

		// 檢查nil
		if nil != infoPointer.DevicePointer {

			// 找相同的場域
			myArea := cTool.getMyAreaByClientPointer(whatKindCommandString, command, clientPointer, details) //取每個clientPointer的場域
			intersection := intersect.Hash(myArea, area)                                                     //取交集array

			fmt.Println("測試中：同場域有哪些場域：intersection=", intersection, "裝置ID＝", infoPointer.DevicePointer.DeviceID)

			// 若有找到
			if len(intersection) > 0 {

				fmt.Println("測試中：找到有交集的裝置為：intersection=", intersection, "裝置ID＝", infoPointer.DevicePointer.DeviceID)

				//if clientInfoMap[client].DevicePointer.Area == area {
				if clientPointer != excluder { //排除自己
					fmt.Println("測試中：正在廣播：intersection=", intersection)

					// 廣播
					otherMessage := "-對某裝置進行廣播-裝置ID=" + infoPointer.DevicePointer.DeviceID
					clientPointer.outputChannel <- websocketData //Socket Response

					// 一般logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details+otherMessage, command, clientPointer) //所有值複製一份做logger
					cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				}
			}
		}
	}
}

// 針對某房間(RoomID)進行廣播，排除某連線(自己)
/**
 * @param roomID int 欲廣播的房間號
 * @param websocketData websocketData 欲廣播的內容
 * @param excluder *client 欲排除的連線指標(通常是自己)
 */
func (cTool *CommandTool) broadcastByRoomID(roomID int, websocketData websocketData, excluder *client) {

	for clientPointer, infoPointer := range clientInfoMap {
		// 檢查nil
		if nil != infoPointer.DevicePointer {
			// 找到相同房間的連線
			if infoPointer.DevicePointer.RoomID == roomID {

				if clientPointer != excluder { //排除自己

					// 廣播
					clientPointer.outputChannel <- websocketData //Socket Response
				}
			}
		}
	}
}

// 取得某clientPointer的場域：(眼鏡端：取眼鏡場域，平版端：取專家場域)
/**
 * @param whatKindCommandString string 是哪個指令呼叫此函式
 * @param command serverResponseStruct.Command 客戶端的指令
 * @param clientPointer *client 連線指標
 * @param details string 之前已處理的詳細訊息
 * @return area []int 回傳查詢到的場域代碼
 */
func (cTool *CommandTool) getMyAreaByClientPointer(whatKindCommandString string, command serverResponseStruct.Command, clientPointer *client, details string) (area []int) {

	if infoPointer, ok := clientInfoMap[clientPointer]; ok {

		devicePointer := infoPointer.DevicePointer

		if nil != devicePointer {
			// 若自己為眼鏡端
			details += "-找到連線"
			if devicePointer.DeviceType == 1 {
				// 取眼鏡場域
				details += "-取眼鏡場端場域"
				area = devicePointer.Area
				fmt.Println("測試：我是眼鏡端取裝置場域＝", area)
				return

				// 若自己為平板端
			} else if devicePointer.DeviceType == 2 {
				//取專家帳號場域
				details += "-取專家場域"
				accountPointer := infoPointer.AccountPointer
				if nil != accountPointer {
					details += "-找到專家帳號"
					area = accountPointer.Area
					fmt.Println("測試：我是專家端取專家場域", area)
					return
				} else {
					//找不到帳號
					details += "-找不到帳號"

					// 警告logger
					myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
					cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
				}
			}
		} else {
			//找不到裝置
			details += "-找不到裝置"

			// 警告logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		}
	} else {
		//找不到連線
		details += "-找不到連線"

		// 警告logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	}

	return []int{}
}

// 將裝置指標 包裝成 裝置實體array
/**
 * @param 裝置指標
 * @return 裝置實體array
 */
func (cTool *CommandTool) getArray(device *serverDataStruct.Device) []serverDataStruct.Device {
	var array = []serverDataStruct.Device{}
	array = append(array, *device)
	return array
}

// 將裝置指標 包裝成 裝置指標array
/**
 * @param 裝置指標
 * @return 裝置指標array
 */
func (cTool *CommandTool) getArrayPointer(device *serverDataStruct.Device) []*serverDataStruct.Device {
	var array = []*serverDataStruct.Device{}
	array = append(array, device)
	return array
}

// 檢查Command的指定欄位是否齊全(command 非指標 不用檢查Nil問題)
/**
 * @param command serverResponseStruct.Command 客戶端的指令
 * @param fields []string 檢查的欄位名稱array
 * @return ok bool 回傳是否齊全
 * @return missFields []string 回傳遺漏的欄位名稱
 */
func (cTool *CommandTool) checkCommandFields(command serverResponseStruct.Command, fields []string) (ok bool, missFields []string) {

	missFields = []string{} // 遺失的欄位
	ok = true               // 是否齊全

	for _, e := range fields {

		switch field := e; field {

		case "command":
			if command.Command == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "commandType":
			if command.CommandType == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "transactionID":
			if command.TransactionID == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "userID":
			if command.UserID == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "userPassword":
			if command.UserPassword == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceID":
			if command.DeviceID == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceBrand":
			if command.DeviceBrand == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceType":
			if command.DeviceType == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "area":
			if command.Area == nil {
				missFields = append(missFields, field)
				ok = false
			}

		case "areaEncryptionString":
			if command.AreaEncryptionString == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "areaName":
			if command.AreaName == nil {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceName":
			if command.DeviceName == nil {
				missFields = append(missFields, field)
				ok = false
			}

		case "pic":
			if command.Pic == "" {
				missFields = append(missFields, field)
				ok = false
			}

		case "onlineStatus":
			if command.OnlineStatus == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "deviceStatus":
			if command.DeviceStatus == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "cameraStatus":
			if command.CameraStatus == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "micStatus":
			if command.MicStatus == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		case "roomID":
			if command.RoomID == 0 {
				missFields = append(missFields, field)
				ok = false
			}

		}

	}

	return ok, missFields
}

// 判斷欄位Command(個別指令專屬欄位)是否齊全，若不齊全直接Response給客戶端
/**
 * @param fields []string 檢查的欄位名稱array
 * @param clientPointer *client 連線指標
 * @param command Command 客戶端的指令
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @return bool 回傳是否齊全
 */
func (cTool *CommandTool) checkFieldsCompletedAndResponseIfFail(fields []string, clientPointer *client, command serverResponseStruct.Command, whatKindCommandString string) bool {

	//fields := []string{"roomID"}
	ok, missFields := cTool.checkCommandFields(command, fields)

	if !ok {

		m := strings.Join(missFields, ",")

		details := `-欄位不齊全:` + m

		// Response: 失敗
		jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
		clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

		// 警告logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return false

	} else {

		return true
	}

}

// // 取得連線資訊物件
// // 取得登入基本資訊字串
// func getLoginBasicInfoString(c *client) string {

// 	s := " UserID:" + clientInfoMap[c].AccountPointer.UserID

// 	if clientInfoMap[c].DevicePointer != nil {
// 		s += ",DeviceID:" + clientInfoMap[c].DevicePointer.DeviceID
// 		s += ",DeviceBrand:" + clientInfoMap[c].DevicePointer.DeviceBrand
// 		s += ",DeviceName:" + clientInfoMap[c].DevicePointer.DeviceName
// 		//s += ",DeviceType:" + (string)(clientInfoMap[c].DevicePointer.DeviceType)
// 		s += ",DeviceType:" + strconv.Itoa(clientInfoMap[c].DevicePointer.DeviceType)                                      // int轉string
// 		s += ",Area:" + strings.Replace(strings.Trim(fmt.Sprint(clientInfoMap[c].DevicePointer.Area), "[]"), " ", ",", -1) //將int[]內容轉成string
// 		s += ",RoomID:" + strconv.Itoa(clientInfoMap[c].DevicePointer.RoomID)                                              // int轉string
// 		s += ",OnlineStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.OnlineStatus)                                  // int轉string
// 		s += ",DeviceStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.DeviceStatus)                                  // int轉string
// 		s += ",CameraStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.CameraStatus)                                  // int轉string
// 		s += ",MicStatus:" + strconv.Itoa(clientInfoMap[c].DevicePointer.MicStatus)                                        // int轉string
// 	}

// 	return s
// }

// 確認是否有此帳號
/**
 * @param id string 使用者帳號
 * @return bool 回傳是否存在此帳號
 * @return *serverDataStruct.Account 回傳帳號指標,若不存在回傳nil
 */
func (cTool *CommandTool) checkAccountExistAndCreateAccountPointer(id string) (bool, *serverDataStruct.Account) {

	result := []model.Account{}
	result = mongoDB.FindAccountByUserID(id)
	// 找到
	if len(result) > 0 {

		accountPointer := &serverDataStruct.Account{
			UserID:       result[0].UserID,
			UserPassword: result[0].UserPassword,
			UserName:     result[0].UserName,
			IsExpert:     result[0].IsExpert,
			IsFrontline:  result[0].IsFrontline,
			Area:         result[0].Area,
			AreaName:     cTool.getAreaNameArrayByAreaIDAarray(result[0].Area), //等待設定
			Pic:          result[0].Pic,
			// verificationCodeTime: result[0].VerificationCodeTime,
			VerificationCodeTime: result[0].VerificationCodeTime,
		}

		return true, accountPointer
	}
	// for _, accountPointer := range allAccountPointerList {
	// 	if nil != accountPointer {
	// 		if id == accountPointer.UserID {
	// 			return true, accountPointer
	// 		}
	// 	}
	// }
	//找不到帳號
	return false, nil
}

// Struct結構: 儲存email之夾帶內容
type mailInfo struct {
	VerificationCode string
}

// 寄送郵件功能
/**
 * @receiver myInfo mailInfo 可以使用此函數的主體結構為 mailInfo 的實體變數，可使用實體變數名稱.sendMail() 來直接使用此函數)
 * @param accountPointer *serverDataStruct.Account 帳戶指標
 * @return success bool 回傳寄送成功或失敗
 * @return otherMessage string 回傳處理的細節
 */
func (myInfo mailInfo) sendMail(accountPointer *serverDataStruct.Account) (success bool, otherMessage string) {

	var err error

	// 帳號不為空
	if nil != accountPointer {

		//建立樣板物件
		t := template.New(`templateVerificationCode.html`) //物件名稱
		fmt.Println("測試", t)

		//轉檔
		t, err = t.ParseFiles(`./template/templateVerificationCode.html`)
		if err != nil {
			log.Println(err)
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, myInfo); err != nil {
			log.Println(err)
		}

		//取得樣板文字結果
		result := tpl.String()

		//建立與設定電子郵件
		m := gomail.NewMessage()
		m.SetHeader("From", "sw@leapsyworld.com")
		m.SetHeader("To", accountPointer.UserID) //Email 就是 userID

		//副本
		//m.SetAddressHeader("Cc", "<RECIPIENT CC>", "<RECIPIENT CC NAME>")

		m.SetHeader("Subject", "Leapsy專家系統-驗證通知信")
		m.SetBody("text/html", result)

		//夾帶檔案
		//m.Attach("template.html") // attach whatever you want

		d := gomail.NewDialer("smtp.qiye.aliyun.com", 25, "sw@leapsyworld.com", "Leapsy123!")

		//寄發電子郵件
		if err := d.DialAndSend(m); err != nil {
			// 寄信發生錯誤
			//panic(err)
			otherMessage = "-寄信發生錯誤"
			success = false
		} else {
			otherMessage = "-順利寄出"
			// 紀錄驗證信發送時間
			// accountPointer.verificationCodeTime = time.Now()
			accountPointer.VerificationCodeTime = time.Now()
			success = true
			fmt.Println("順利寄出", success)
		}
	} else {
		// 帳號為空
		otherMessage = "-帳號為空"
		success = false
	}

	return
}

// 產生驗證碼並寄送郵件
/**
 * @param accountPointer *serverDataStruct.Account 帳戶指標
 * @param whatKindCommandString string 是哪個指令呼叫此函數
 * @param details string 之前已經處理過的細節
 * @param command serverResponseStruct.Command 客戶端的指令
 * @param clientPointer *client 連線指標
 * @return success bool 回傳寄送成功或失敗
 * @return returnMessages string 回傳處理的細節
 */
func (cTool *CommandTool) processSendVerificationCodeMail(accountPointer *serverDataStruct.Account, whatKindCommandString string, details string, command serverResponseStruct.Command, clientPointer *client) (success bool, returnMessages string) {

	// 建立隨機密string六碼
	verificationCode := strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10))

	//給logger用的
	details += `-建立隨機六碼,verificationCode=` + verificationCode

	//可能會回給前端用的
	returnMessages += "-建立驗證碼六碼"

	// 正常logger
	fmt.Println(details)
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	//密碼記在account中:(未來驗證碼不用紀錄在資料庫，因為設計中，密碼是一次性密碼。即使主機當機，所有USER也需要重新登入，自然要重新取一次驗證信)
	if nil != accountPointer {

		details += "-找到帳號,userID=" + accountPointer.UserID
		returnMessages += "-找到帳號,userID=" + accountPointer.UserID

		// 儲存隨機六碼到帳戶
		accountPointer.UserPassword = verificationCode

		details += "-帳戶已更新驗證碼"
		returnMessages += "-帳戶已更新驗證碼"

		// 準備寄送包含密碼的EMAIL
		d := mailInfo{verificationCode}
		// emailString := accountPointer.UserID

		// 寄送郵件
		ok, errMsg := d.sendMail(accountPointer)

		if ok {
			// 已寄出
			success = true
			details += "-驗證碼已成功寄出"
			returnMessages += "-驗證碼已成功寄出"

			// 正常logger
			fmt.Println(details)
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return
		} else {
			// 寄出失敗:請確認您的電子信箱是否正確
			success = false
			details += "-驗證碼寄出失敗:請確認您的電子信箱是否正確-錯誤訊息:" + errMsg
			returnMessages += "-驗證碼寄出失敗:請確認您的電子信箱是否正確-錯誤訊息:" + errMsg

			// 警告logger
			fmt.Println(details)
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return
		}
	} else {
		// accountPointer 為nil
		success = false
		details += "-帳號不存在"
		returnMessages += "-帳號不存在"

		// 警告logger
		fmt.Println(details)
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return
	}

	// return
	//額外:進行登出時要去把對應的password移除
}

// 處理<我的裝置區域>的廣播，廣播內容為某些裝置的狀態變更
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command serverResponseStruct.Command 客戶端的指令
* @param clientPointer *client 連線指標(我的裝置區域從此變數來)
* @param devicePointerArray []*serverDataStruct.Device 要廣播出去的所有Device內容
* @param details string 之前已經處理過的細節
* @return string 回傳處理的細節
**/
func (cTool *CommandTool) processBroadcastingDeviceChangeStatusInMyArea(whatKindCommandString string, command serverResponseStruct.Command, clientPointer *client, devicePointerArray []*serverDataStruct.Device, details string) string {

	// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 Device[] 陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(serverResponseStruct.DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, DevicePointer: devicePointerArray}); err == nil {
		//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

		// 準備廣播

		// 取出自己的場域
		area := cTool.getMyAreaByClientPointer(whatKindCommandString, command, clientPointer, details)

		if len(area) > 0 {
			// 若有找到場域
			strArea := fmt.Sprintln(area)
			details += `-執行（場域）廣播成功,場域代碼=` + strArea

			// 廣播(場域、排除個人)
			cTool.broadcastByArea(area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, whatKindCommandString, command, clientPointer, details) // 排除個人進行Area廣播

			// 一般logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return details

		} else {
			// 若沒找到場域
			details += `-執行（場域）廣播失敗,沒找到自己的場域`

			// 警告logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
			return details
		}

	} else {

		// 錯誤logger
		details += `-執行（場域）廣播失敗,後端json轉換出錯`
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return details
	}
}

// 處理<指定區域>的廣播，廣播內容為某些裝置狀態的變更
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command serverResponseStruct.Command 客戶端的指令
* @param clientPointer *client 連線指標(我的區域從此變數來)
* @param device []*serverDataStruct.Device 要廣播出去的所有Device內容
* @param area []int 想廣播的區域
* @param details string 之前已經處理過的細節
**/
func (cTool *CommandTool) processBroadcastingDeviceChangeStatusInSomeArea(whatKindCommandString string, command serverResponseStruct.Command, clientPointer *client, device []*serverDataStruct.Device, area []int, details string) {

	// 進行廣播:(此處仍使用Marshal工具轉型，因考量有 serverDataStruct.Device[] 陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(serverResponseStruct.DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInArea, CommandType: CommandTypeNumberOfBroadcast, DevicePointer: device}); err == nil {
		//jsonBytes = []byte(fmt.Sprintf(baseBroadCastingJsonString1, CommandNumberOfBroadcastingInArea, CommandTypeNumberOfBroadcast, device))

		// 廣播(場域、排除個人)
		cTool.broadcastByArea(area, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, whatKindCommandString, command, clientPointer, details) // 排除個人進行Area廣播

		// logger
		details := `執行（指定場域）廣播成功-場域代碼=` + strconv.Itoa(area[0])
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	} else {

		// logger
		details := `執行（指定場域）廣播失敗：後端json轉換出錯`
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

	}
}

// 處理<我的裝置房間>的廣播，廣播內容為我的裝置狀態的變更
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command serverResponseStruct.Command 客戶端的指令
* @param clientPointer *client 連線指標(我的房號從此變數來)
* @param devicePointerArray []*serverDataStruct.Device 要廣播出去的所有Device內容
* @param details string 之前已經處理過的細節
* @return string 回傳處理的細節
**/
func (cTool *CommandTool) processBroadcastingDeviceChangeStatusInRoom(whatKindCommandString string, command serverResponseStruct.Command, clientPointer *client, devicePointerArray []*serverDataStruct.Device, details string) string {

	// (此處仍使用Marshal工具轉型，因考量Device[]的陣列形態，轉成string較為複雜。)
	if jsonBytes, err := json.Marshal(serverResponseStruct.DeviceStatusChangeByPointer{Command: CommandNumberOfBroadcastingInRoom, CommandType: CommandTypeNumberOfBroadcast, DevicePointer: devicePointerArray}); err == nil {

		var roomID = 0

		if infoPointer, ok := clientInfoMap[clientPointer]; ok {
			devicePointer := infoPointer.DevicePointer
			if nil != devicePointer {
				// 找到房號
				roomID = devicePointer.RoomID
				details += `-找到裝置ID=` + devicePointer.DeviceID + `,裝置品牌=` + devicePointer.DeviceBrand + `,與裝置房號=` + strconv.Itoa(roomID)

				// 房間廣播:改變麥克風/攝影機狀態
				cTool.broadcastByRoomID(roomID, websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}, clientPointer) // 排除個人進行Area廣播

				// 一般logger
				details += `-執行（房間）廣播成功,於房間號碼=` + strconv.Itoa(roomID)
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
				cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return details
			} else {
				//找不到裝置
				details += `-找不到裝置,執行（房間）廣播失敗`

				// 警告logger
				myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
				cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

				return details
			}
		} else {
			//找不到連線
			details += `-找不到連線,執行（房間）廣播失敗`

			// 警告logger
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

			return details
		}

	} else {
		// 後端json轉換出錯

		details += `-後端json轉換出錯,執行（房間）廣播失敗`

		// 錯誤logger
		myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
		cTool.processLoggerErrorf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		return details
	}
}

// 取得某場域的線上閒置專家數
/**
* @param area []int 想要計算的場域代碼array
* @param whatKindCommandString string 是哪個指令呼叫此函數 (for log)
* @param command serverResponseStruct.Command 客戶端的指令 (for log)
* @param clientPointer *client 連線指標 (for log)
* @return int 回傳結果
**/
func (cTool *CommandTool) getOnlineIdleExpertsCountInArea(area []int, whatKindCommandString string, command serverResponseStruct.Command, clientPointer *client) int {

	counter := 0
	for _, e := range clientInfoMap {

		//找出同場域的專家帳號＋裝置閒置
		accountPointer := e.AccountPointer
		if accountPointer != nil {

			if 1 == accountPointer.IsExpert {

				intersection := intersect.Hash(accountPointer.Area, area) //取交集array

				// 若場域有交集
				if len(intersection) > 0 {

					//是否閒置
					if 1 == e.DevicePointer.DeviceStatus {
						counter++
					}
				}
			}
		} else {
			//帳號為空 不Response特別處理 僅記錄
			//logger
			details := `-發現accountPointer為空`
			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
			cTool.processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

		}
	}

	return counter
}

//取得同房間其他人連線
// func getOtherClientsInTheSameRoom(clientPoint *client, roomID int) []*client {

// 	results := []*client{}

// 	for i, e := range clientInfoMap {
// 		if clientPoint != i {
// 			if roomID == e.DevicePointer.RoomID {
// 				results = append(results, i)
// 			}
// 		}
// 	}

// 	return results
// }

// 取得同房號的其他所有人裝置指標
/**
* @param roomID int (房號)
* @param clientPoint *client 排除的連線指標(通常為自己)
* @return []*serverDataStruct.Device 回傳結果
**/
func (cTool *CommandTool) getOtherDevicesInTheSameRoom(roomID int, clientPoint *client) []*serverDataStruct.Device {

	results := []*serverDataStruct.Device{}

	for cPointer, infoPointer := range clientInfoMap {

		// 排除自己
		if clientPoint != cPointer {

			//取連線
			if nil != infoPointer {
				//取裝置
				if nil != infoPointer.DevicePointer {
					//找同房間
					if roomID == infoPointer.DevicePointer.RoomID {
						//加入結果清單
						results = append(results, infoPointer.DevicePointer)
					}
				} else {
					//找不到裝置
				}
			} else {
				// 找不到連線
			}
		}
	}

	return results
}

// 取得所有Log需要的參數(都取COPY，為了記錄當時狀況，避免平行處理值改變了)
/**
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param details string 之前已經處理過的細節
* @param command serverResponseStruct.Command 客戶端的指令
* @param clientPointer *client 連線指標

* @return myAccount serverDataStruct.Account 帳戶實體COPY
* @return myDevice serverDataStruct.Device 裝置實體COPY
* @return myClient client 連線實體COPY
* @return myClientInfoMap map[*client]*Info 連線與Info對應Map實體COPY
* @return myAllDevices []serverDataStruct.Device 所有裝置清單實體COPY (為了印出log，先而取出所有實體，若使用pointer無法直接透過%+v印出)
* @return nowRoomId int 最後取到的房號COPY
**/
func (cTool *CommandTool) getLoggerParrameters(whatKindCommandString string, details string, command serverResponseStruct.Command, clientPointer *client) (myAccount serverDataStruct.Account, myDevice serverDataStruct.Device, myClient client, myClientInfoMap map[*client]*serverDataStruct.Info, myAllDevices []serverDataStruct.Device, nowRoomId int) {

	if clientInfoMap != nil {
		myClientInfoMap = clientInfoMap

		if clientPointer != nil {
			myClient = *clientPointer

			if infoPointer, ok := clientInfoMap[clientPointer]; ok {

				myAccount = *infoPointer.AccountPointer
				myDevice = *infoPointer.DevicePointer

				// if e.AccountPointer != nil {
				// 	myAccount = *e.AccountPointer
				// }

				// if e.DevicePointer != nil {
				// 	myDevice = *e.DevicePointer
				// }
			}
		}
	}

	myAllDevices = cTool.getAllDeviceByList() // 取得裝置清單-實體(為了印出log，先而取出所有實體，若使用pointer無法直接透過%+v印出)
	nowRoomId = roomID                        // 目前房號

	return
}

// 處理發現nil的logger
// func processNilLoggerInfof(whatKindCommandString string, details string, otherMessages string, command serverResponseStruct.Command) {

// 	go fmt.Printf(baseLoggerInfoNilMessage+"\n", whatKindCommandString, details, otherMessages, command)
// 	go logger.Infof(baseLoggerInfoNilMessage, whatKindCommandString, details, otherMessages, command)

// }

// 處理<一般logger>
/**
* @param whatKindCommandString string 哪個指令發出的
* @param details string 詳細訊息
* @param command serverResponseStruct.Command 客戶端的指令
* @param myAccount serverDataStruct.Account 帳戶(連線本身)
* @param myDevice serverDataStruct.Device 裝置(連線本身)
* @param myClientPointer client 連線
* @param myClientInfoMap map[*client]*Info 所有在線連線的info(裝置+帳戶之配對)
* @param myAllDevices []serverDataStruct.Device 所有已匯入裝置的狀態訊息
* @param nowRoomID int 線在房號
**/
func (cTool *CommandTool) processLoggerInfof(whatKindCommandString string, details string, command serverResponseStruct.Command, myAccount serverDataStruct.Account, myDevice serverDataStruct.Device, myClientPointer client, myClientInfoMap map[*client]*serverDataStruct.Info, myAllDevices []serverDataStruct.Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏

	strClientInfoMap := cTool.getStringOfClientInfoMap() //所有連線、裝置、帳號資料
	go fmt.Printf(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)
	go logger.Infof(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)

}

// 處理<警告logger>
/**
* @param 參數同處理<一般logger>
**/
func (cTool *CommandTool) processLoggerWarnf(whatKindCommandString string, details string, command serverResponseStruct.Command, myAccount serverDataStruct.Account, myDevice serverDataStruct.Device, myClientPointer client, myClientInfoMap map[*client]*serverDataStruct.Info, myAllDevices []serverDataStruct.Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏

	strClientInfoMap := cTool.getStringOfClientInfoMap() //所有連線、裝置、帳號資料
	go fmt.Printf(baseLoggerCommonMessage+"\n\n", whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)
	go logger.Warnf(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)

}

// 處理<錯誤logger>
/**
* @param 參數同處理<一般logger>
**/
func (cTool *CommandTool) processLoggerErrorf(whatKindCommandString string, details string, command serverResponseStruct.Command, myAccount serverDataStruct.Account, myDevice serverDataStruct.Device, myClientPointer client, myClientInfoMap map[*client]*serverDataStruct.Info, myAllDevices []serverDataStruct.Device, nowRoomID int) {

	myAccount.UserPassword = "" //密碼隱藏

	strClientInfoMap := cTool.getStringOfClientInfoMap() //所有連線、裝置、帳號資料
	go fmt.Printf(baseLoggerCommonMessage+"\n\n", whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)
	go logger.Errorf(baseLoggerCommonMessage, whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, strClientInfoMap, myAllDevices, nowRoomID)

}

// 取得帳號圖片
/**
* @param fileName string 檔名
* @return string 回傳檔案內容
**/
func (cTool *CommandTool) getAccountPicString(fileName string) string {

	content, err := ioutil.ReadFile(fileName)

	//錯誤
	if err != nil {
		log.Fatal(err)
		return ""
	}

	// Convert []byte to string and print to screen
	text := string(content)
	fmt.Println(text)

	return text
}

// // 檢查clientInfoMap 是否有nil pointer狀況
// func checkAndGetClientInfoMapNilPoter(whatKindCommandString string, details string, command serverResponseStruct.Command, clientPointer *client) (myInfoPointer *Info, myDevicePointer *serverDataStruct.Device, myAccountPointer *serverDataStruct.Account) {

// 	myInfoPointer = &Info{}
// 	myDevicePointer = &serverDataStruct.Device{}
// 	myAccountPointer = &serverDataStruct.Account{}

// 	if e, ok := clientInfoMap[clientPointer]; ok {

// 		if e.DevicePointer != nil {
// 			myDevicePointer = e.DevicePointer
// 		} else {

// 			// logger
// 			details := `發現DevicePointer為nil`
// 			myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
// 			processLoggerInfof(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)

// 		}

// 		if e.AccountPointer != nil {
// 			myAccountPointer = e.AccountPointer
// 		} else {

// 		}

// 	} else {
// 		//info is nil
// 	}

// 	return
// }

// 處理連線Info為空Response給客戶端
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command serverResponseStruct.Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func (cTool *CommandTool) processResponseInfoNil(clientPointer *client, whatKindCommandString string, command serverResponseStruct.Command, details string) {
	// Response:失敗
	details += `-執行失敗-找不到連線`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現infoPointer為空`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 處理帳號為空Response給客戶端
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command serverResponseStruct.Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func (cTool *CommandTool) processResponseAccountNil(clientPointer *client, whatKindCommandString string, command serverResponseStruct.Command, details string) {
	// Response:失敗
	details += `-執行失敗-找不到帳號`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現accountPointer為空`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 處理裝置為空Response給客戶端
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command serverResponseStruct.Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func (cTool *CommandTool) processResponseDeviceNil(clientPointer *client, whatKindCommandString string, command serverResponseStruct.Command, details string) {
	// Response:失敗
	details += `-執行失敗-找不到裝置`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現devicePointer為空`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 處理某指標為空Response
/**
* @param clientPointer *client 連線指標
* @param whatKindCommandString string 是哪個指令呼叫此函數
* @param command serverResponseStruct.Command 客戶端的指令
* @param details string 之前已經處理的細節
**/
func (cTool *CommandTool) processResponseNil(clientPointer *client, whatKindCommandString string, command serverResponseStruct.Command, details string) {
	// Response:失敗
	details += `-執行失敗`

	jsonBytes := []byte(fmt.Sprintf(baseResponseJsonString, command.Command, CommandTypeNumberOfAPIResponse, ResultCodeFail, details, command.TransactionID))
	clientPointer.outputChannel <- websocketData{wsOpCode: ws.OpText, dataBytes: jsonBytes}

	// logger:發現Device指標為空
	details += `-發現空指標`
	myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom := cTool.getLoggerParrameters(whatKindCommandString, details, command, clientPointer) //所有值複製一份做logger
	cTool.processLoggerWarnf(whatKindCommandString, details, command, myAccount, myDevice, myClientPointer, myClientInfoMap, myAllDevices, nowRoom)
}

// 取得log字串:針對ClientInfoMap(即所有在線的連線、裝置、帳號配對)
func (cTool *CommandTool) getStringOfClientInfoMap() (results string) {

	for myClient, myInfo := range clientInfoMap {
		results += fmt.Sprintf(`【連線%v,裝置%v,帳號%v】
		
		`, myClient, myInfo.DevicePointer, myInfo.AccountPointer)
	}

	return
}

// 取得log字串:針對指定的 info array(可用來取得紀錄平查詢到的所有連線、裝置、帳號配對)
func (cTool *CommandTool) getStringByInfoPointerArray(infoPointerArray []*serverDataStruct.Info) (results string) {

	for i, infoPointer := range infoPointerArray {
		if nil != infoPointer {

			results += `第` + strconv.Itoa(i+1) + `組裝置與帳號:[`

			devicePointer := infoPointer.DevicePointer
			if nil != devicePointer {
				//str := fmt.Sprint(devicePointer)
				stringArea := fmt.Sprint(devicePointer.Area)
				stringAreaName := fmt.Sprint(devicePointer.AreaName)

				results += `裝置{` +
					`裝置ID=` + devicePointer.DeviceID +
					`,裝置Brand=` + devicePointer.DeviceBrand +
					`,裝置OnlineStatus=` + strconv.Itoa(devicePointer.OnlineStatus) +
					`,裝置DeviceStatus=` + strconv.Itoa(devicePointer.DeviceStatus) +
					`,裝置場域代號=` + stringArea +
					`,裝置場域名稱=` + stringAreaName +
					`}`
			} else {
				results += `裝置{找不到此裝置}`
			}
			accountPointer := infoPointer.AccountPointer
			if nil != accountPointer {
				//str := fmt.Sprint(accountPointer)
				results += `,帳號{` +
					`,帳號userID=` + accountPointer.UserID +
					`,帳號是否為專家=` + strconv.Itoa(accountPointer.IsExpert) +
					`,帳號是否為一線人員=` + strconv.Itoa(accountPointer.IsFrontline) +
					`}`
			} else {
				results += `,帳號{尚未登入}`
			}

			results += `]`

		}
	}
	results += `以上為所有結果。`

	return
}

// 取得場域
func (cTool *CommandTool) getAreaById(areaID int) []model.Area {

	result := []model.Area{}
	result = mongoDB.FindAreaById(areaID)
	return result
}

// 更新為新的areaID
func (cTool *CommandTool) updateDeviceAreaIDInMongoDBAndGetOneDevicePointer(newAreaID int, devicePointer *serverDataStruct.Device) (newDevicePointer *serverDataStruct.Device) {

	result := []model.Device{}
	result = mongoDB.UpdateOneDeviceArea(newAreaID, devicePointer.DeviceID, devicePointer.DeviceBrand)

	if len(result) > 0 {

		newDevicePointer = &serverDataStruct.Device{
			DeviceID:     result[0].DeviceID,
			DeviceBrand:  result[0].DeviceBrand,
			DeviceType:   result[0].DeviceType,
			Area:         result[0].Area,
			AreaName:     cTool.getAreaNameArrayByAreaIDAarray(result[0].Area),
			DeviceName:   result[0].DeviceName,
			Pic:          "",                         // 客戶端求助時才設定
			OnlineStatus: devicePointer.OnlineStatus, // 沿用裝置之前設定
			DeviceStatus: devicePointer.DeviceStatus, // 沿用裝置之前設定
			CameraStatus: devicePointer.CameraStatus, // 沿用裝置之前設定
			MicStatus:    devicePointer.MicStatus,    // 沿用裝置之前設定
			RoomID:       devicePointer.RoomID,       // 沿用裝置之前設定
		}

	} else {

		newDevicePointer = nil
	}

	return
}
