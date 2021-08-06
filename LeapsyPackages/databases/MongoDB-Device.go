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

// findAlertRecords - 查找所有裝置
/**
 * @param bson.M filter 過濾器
 * @param ...*options.FindOptions opts 選項
 * @return results []model.Device 取得結果
 */
func (mongoDB *MongoDB) findDevices(filter primitive.M, opts ...*options.FindOptions) (results []model.Device) {

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
			Collection(mongoDB.GetConfigValueOrPanic(`mongoDB`, `device-table`)).
			Find(
				context.TODO(),
				filter,
				opts...,
			)

		// alertRWMutex.RUnlock() // 讀解鎖

		go logger.Errorf(`%+v %s 查找裝置 `, append(defaultArgs, filter), findError)

		if nil != findError { // 若查找警報紀錄錯誤
			return // 回傳
		}

		defer cursor.Close(context.TODO()) // 記得關閉

		for cursor.Next(context.TODO()) { // 針對每一紀錄

			var device model.Device

			cursorDecodeError := cursor.Decode(&device) // 解析紀錄

			go logger.Errorf(`%+v 取得裝置 %+s`, append(defaultArgs, device), cursorDecodeError)

			if nil != cursorDecodeError { // 若解析記錄錯誤
				return // 回傳
			}

			// device.AlertEventTime = device.AlertEventTime.Local() // 儲存為本地時間格式

			results = append(results, device) // 儲存紀錄
		}

		cursorErrError := cursor.Err() // 游標錯誤

		go logger.Errorf(`%+v %s 查找裝置遊標運作`, defaultArgs, cursorErrError)

		if nil != cursorErrError { // 若遊標錯誤
			return // 回傳
		}

		go logger.Infof(`%+v 取得裝置`, append(defaultArgs, results))

	}

	return // 回傳
}

// FindAllAlertRecords - 取得所有裝置
/**
 * @return results []model.Device 取得結果
 */
func (mongoDB *MongoDB) FindAllDevices() (results []model.Device) {

	// 取得警報紀錄
	// results = mongoDB.findAlertRecords(bson.M{}, options.Find().SetSort(bson.M{`alerteventtime`: -1}).SetBatchSize(int32(batchSize)))

	results = mongoDB.findDevices(bson.M{}, nil)

	return // 回傳
}

func (mongoDB *MongoDB) FindDevicesByDeviceIDAndDeviceBrand(deviceID string, deviceBrand string) (results []model.Device) {

	// 取得警報紀錄
	// results = mongoDB.findAlertRecords(bson.M{}, options.Find().SetSort(bson.M{`alerteventtime`: -1}).SetBatchSize(int32(batchSize)))

	results = mongoDB.findDevices(bson.M{`deviceID`: deviceID, `deviceBrand`: deviceBrand}, nil)

	return // 回傳
}

// findOneAndUpdateAreaSET - 提供更新部分欄位
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @param ...*options.FindOneAndUpdateOptions 選項
 * @return result []records.AlertRecord 更添結果
 */
func (mongoDB *MongoDB) findOneAndUpdateDeviceSET(
	filter, update primitive.M,
	opts ...*options.FindOneAndUpdateOptions) (results []model.Device) {

	return mongoDB.findOneAndUpdateDevice(
		filter,
		bson.M{
			`$set`: update,
		},
	)

}

// findOneAndUpdateArea - 提供可以丟整個物件的更新(primitive.M)
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @param ...*options.FindOneAndUpdateOptions 選項
 * @return result []records.AlertRecord 更添結果
 */
func (mongoDB *MongoDB) findOneAndUpdateDevice(
	filter, update primitive.M,
	opts ...*options.FindOneAndUpdateOptions) (results []model.Device) {

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

		// alertRWMutex.Lock() // 寫鎖

		// 更新警報記錄
		singleResultPointer := mongoClientPointer.
			Database(mongoDB.GetConfigValueOrPanic(`mongoDB`, `database`)).
			Collection(mongoDB.GetConfigValueOrPanic(`mongoDB`, `device-table`)).
			FindOneAndUpdate(
				context.TODO(),
				filter,
				update,
				opts...,
			)

			/*
				FindOneAndUpdate(
					context.TODO(),
					filter,
						bson.M{
							`$set`:update,
						},
						opts...,
					)
			*/

		// alertRWMutex.Unlock() // 寫解鎖

		findOneAndUpdateError := singleResultPointer.Err() // 更添錯誤

		if nil != findOneAndUpdateError { // 若更添警報紀錄錯誤且非檔案不存在錯誤

			go logger.Errorf(`%+v 修改裝置場域，Error= %+v。`, append(defaultArgs, update), findOneAndUpdateError)

			// logings.SendLog(
			// 	[]string{`%s %s 修改裝置場域 %+v `},
			// 	append(defaultArgs, update),
			// 	findOneAndUpdateError,
			// 	logrus.ErrorLevel,
			// )

			return // 回傳
		}

		go logger.Infof(`%+v 修改裝置場域，Error= %+v。`, append(defaultArgs, update), findOneAndUpdateError)

		// logings.SendLog(
		// 	[]string{`%s %s 更添警報記錄 %+v `},
		// 	append(defaultArgs, update),
		// 	findOneAndUpdateError,
		// 	logrus.InfoLevel,
		// )

		results = mongoDB.findDevices(filter)

	}

	return
}

// UpdateOneArea - 更新警報記錄
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @return *mongo.UpdateResult returnUpdateResult 更新結果
 */
func (mongoDB *MongoDB) UpdateOneDeviceArea(newAreaID int, deviceID string, deviceBrand string) (results []model.Device) {

	updatedModelDevice := mongoDB.findOneAndUpdateDeviceSET(
		bson.M{
			`deviceID`:    deviceID,
			`deviceBrand`: deviceBrand,
		},
		bson.M{
			`area`: []int{newAreaID},
		},
	) // 更新的紀錄
	// fmt.Println("標記：", bson.M{`deviceID`: deviceID, `deviceBrand`: deviceBrand}, bson.M{`area.$[]`: newAreaID})

	if nil != updatedModelDevice { // 若更新沒錯誤
		results = append(results, updatedModelDevice...) // 回傳結果
	}

	return
}
