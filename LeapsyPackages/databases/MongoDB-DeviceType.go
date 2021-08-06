package databases

import (
	"context"
	"fmt"

	"leapsy.com/packages/model"
	"leapsy.com/packages/network"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// findAlertRecords - 查找裝置類型
/**
 * @param bson.M filter 過濾器
 * @param ...*options.FindOptions opts 選項
 * @return results []model.DeviceType 取得結果
 */
func (mongoDB *MongoDB) findDeviceType(filter primitive.M, opts ...*options.FindOptions) (results []model.DeviceType) {

	mongoClientPointer := mongoDB.Connect() // 資料庫指標

	if nil != mongoClientPointer { // 若資料庫指標不為空

		defer mongoDB.Disconnect(mongoClientPointer) // 記得關閉資料庫指標

		// 預設主機
		address := fmt.Sprintf(
			`%s:%d`,
			mongoDB.GetConfigValueOrPanic(`mongoDB`, `server`),
			mongoDB.GetConfigPositiveIntValueOrPanic(`mongoDB`, `port`),
		)

		defaultArgs := network.GetAliasAddressPair(address) // 預設參數

		// .RLock() // 讀鎖

		// 查找紀錄
		cursor, findError := mongoClientPointer.
			Database(mongoDB.GetConfigValueOrPanic(`mongoDB`, `database`)).
			Collection(mongoDB.GetConfigValueOrPanic(`mongoDB`, `deviceType-table`)).
			Find(
				context.TODO(),
				filter,
				opts...,
			)

		// alertRWMutex.RUnlock() // 讀解鎖

		go logger.Errorf(`%+v %s 查找裝置類型 `, append(defaultArgs, filter), findError)

		if nil != findError { // 若查找警報紀錄錯誤
			return // 回傳
		}

		defer cursor.Close(context.TODO()) // 記得關閉

		for cursor.Next(context.TODO()) { // 針對每一紀錄

			var account model.DeviceType

			cursorDecodeError := cursor.Decode(&account) // 解析紀錄

			go logger.Errorf(`%+v 取得裝置類型 %+s`, append(defaultArgs, account), cursorDecodeError)

			if nil != cursorDecodeError { // 若解析記錄錯誤
				return // 回傳
			}

			// device.AlertEventTime = device.AlertEventTime.Local() // 儲存為本地時間格式

			results = append(results, account) // 儲存紀錄
		}

		cursorErrError := cursor.Err() // 游標錯誤

		go logger.Errorf(`%+v %s 查找裝置類型遊標運作`, defaultArgs, cursorErrError)

		if nil != cursorErrError { // 若遊標錯誤
			return // 回傳
		}

		go logger.Infof(`%+v 取得裝置類型`, append(defaultArgs, results))

	}

	return // 回傳
}

// FindAllAlertRecords - 取得所有裝置類型
/**
 * @return results []model.DeviceType 取得結果
 */
func (mongoDB *MongoDB) FindAllDeviceTypes() (results []model.DeviceType) {

	// 取得警報紀錄
	// results = mongoDB.findAlertRecords(bson.M{}, options.Find().SetSort(bson.M{`alerteventtime`: -1}).SetBatchSize(int32(batchSize)))

	results = mongoDB.findDeviceType(bson.M{}, nil)

	return // 回傳
}

func (mongoDB *MongoDB) FindDeviceTypesById(id int) (results []model.DeviceType) {

	// 取得警報紀錄
	// results = mongoDB.findAlertRecords(bson.M{}, options.Find().SetSort(bson.M{`alerteventtime`: -1}).SetBatchSize(int32(batchSize)))

	results = mongoDB.findDeviceType(bson.M{`id`: id}, nil)

	return // 回傳
}
