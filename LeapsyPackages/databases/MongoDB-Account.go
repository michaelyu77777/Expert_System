package databases

import (
	"context"
	"fmt"
	"time"

	"leapsy.com/packages/model"
	"leapsy.com/packages/network"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// findAlertRecords - 查找帳號
/**
 * @param bson.M filter 過濾器
 * @param ...*options.FindOptions opts 選項
 * @return results []model.Account 取得結果
 */
func (mongoDB *MongoDB) findAccounts(filter primitive.M, opts ...*options.FindOptions) (results []model.Account) {

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
			Collection(mongoDB.GetConfigValueOrPanic(`mongoDB`, `account-table`)).
			Find(
				context.TODO(),
				filter,
				opts...,
			)

		// alertRWMutex.RUnlock() // 讀解鎖

		go logger.Errorf(`%+v %s 查找帳戶 `, append(defaultArgs, filter), findError)

		if nil != findError { // 若查找警報紀錄錯誤
			return // 回傳
		}

		defer cursor.Close(context.TODO()) // 記得關閉

		for cursor.Next(context.TODO()) { // 針對每一紀錄

			var account model.Account

			cursorDecodeError := cursor.Decode(&account) // 解析紀錄

			go logger.Errorf(`%+v 取得帳戶 %+s`, append(defaultArgs, account), cursorDecodeError)

			if nil != cursorDecodeError { // 若解析記錄錯誤
				return // 回傳
			}

			// device.AlertEventTime = device.AlertEventTime.Local() // 儲存為本地時間格式

			results = append(results, account) // 儲存紀錄
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

// FindAllAlertRecords - 取得所有帳號
/**
 * @return results []model.Account 取得結果
 */
func (mongoDB *MongoDB) FindAllAccounts() (results []model.Account) {

	// 取得警報紀錄
	// results = mongoDB.findAlertRecords(bson.M{}, options.Find().SetSort(bson.M{`alerteventtime`: -1}).SetBatchSize(int32(batchSize)))

	results = mongoDB.findAccounts(bson.M{}, nil)

	return // 回傳
}

func (mongoDB *MongoDB) FindAccountByUserID(userID string) (results []model.Account) {

	// 取得警報紀錄
	// results = mongoDB.findAlertRecords(bson.M{}, options.Find().SetSort(bson.M{`alerteventtime`: -1}).SetBatchSize(int32(batchSize)))

	results = mongoDB.findAccounts(bson.M{`userID`: userID}, nil)

	return // 回傳
}

// findOneAndUpdateAccountSET - 提供更新部分欄位
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @param ...*options.FindOneAndUpdateOptions 選項
 * @return result []records.AlertRecord 更添結果
 */
func (mongoDB *MongoDB) findOneAndUpdateAccountSET(
	filter, update primitive.M,
	opts ...*options.FindOneAndUpdateOptions) (results []model.Account) {

	return mongoDB.findOneAndUpdateAccount(
		filter,
		bson.M{
			`$set`: update,
		},
	)

}

// findOneAndUpdateAccount - 提供可以丟整個物件的更新(primitive.M)
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @param ...*options.FindOneAndUpdateOptions 選項
 * @return result []records.AlertRecord 更添結果
 */
func (mongoDB *MongoDB) findOneAndUpdateAccount(
	filter, update primitive.M,
	opts ...*options.FindOneAndUpdateOptions) (results []model.Account) {

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
			Collection(mongoDB.GetConfigValueOrPanic(`mongoDB`, `account-table`)).
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

			go logger.Errorf(`%+v 修改帳號密碼，Error= %+v。`, append(defaultArgs, update), findOneAndUpdateError)

			// logings.SendLog(
			// 	[]string{`%s %s 修改裝置場域 %+v `},
			// 	append(defaultArgs, update),
			// 	findOneAndUpdateError,
			// 	logrus.ErrorLevel,
			// )

			return // 回傳
		}

		go logger.Infof(`%+v 修改帳號密碼，Error= %+v。`, append(defaultArgs, update), findOneAndUpdateError)

		// logings.SendLog(
		// 	[]string{`%s %s 更添警報記錄 %+v `},
		// 	append(defaultArgs, update),
		// 	findOneAndUpdateError,
		// 	logrus.InfoLevel,
		// )

		results = mongoDB.findAccounts(filter)

	}

	return
}

// UpdateOneAccountPassword - 更新帳戶驗證碼
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @return *mongo.UpdateResult returnUpdateResult 更新結果
 */
func (mongoDB *MongoDB) UpdateOneAccountPassword(userPassword string, userID string) (results []model.Account) {

	updatedModelAccount := mongoDB.findOneAndUpdateAccountSET(
		bson.M{
			`userID`: userID,
		},
		bson.M{
			`userPassword`: userPassword,
		},
	) // 更新的紀錄
	// fmt.Println("標記：", bson.M{`deviceID`: deviceID, `deviceBrand`: deviceBrand}, bson.M{`area.$[]`: newAreaID})

	if nil != updatedModelAccount { // 若更新沒錯誤
		results = append(results, updatedModelAccount...) // 回傳結果
	}

	return
}

// UpdateOneArea - 更新帳戶驗證碼
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @return *mongo.UpdateResult returnUpdateResult 更新結果
 */
func (mongoDB *MongoDB) UpdateOneAccountVerificationCode(verificationCode string, userID string) (results []model.Account) {

	updatedModelAccount := mongoDB.findOneAndUpdateAccountSET(
		bson.M{
			`userID`: userID,
		},
		bson.M{
			`verificationCode`: verificationCode,
		},
	) // 更新的紀錄
	// fmt.Println("標記：", bson.M{`deviceID`: deviceID, `deviceBrand`: deviceBrand}, bson.M{`area.$[]`: newAreaID})

	if nil != updatedModelAccount { // 若更新沒錯誤
		results = append(results, updatedModelAccount...) // 回傳結果
	}

	return
}

// UpdateOneArea - 更新帳戶有效時間
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @return *mongo.UpdateResult returnUpdateResult 更新結果
 */
func (mongoDB *MongoDB) UpdateOneAccountPasswordAndVerificationCodeValidPeriod(userPassword string, validPeriod time.Time, userID string) (results []model.Account) {

	updatedModelAccount := mongoDB.findOneAndUpdateAccountSET(
		bson.M{
			`userID`: userID,
		},
		bson.M{
			`userPassword`:                userPassword,
			`verificationCodeValidPeriod`: validPeriod,
			// `verificationCodeTime`: validPeriod,
		},
	) // 更新的紀錄
	// fmt.Println("標記：", bson.M{`deviceID`: deviceID, `deviceBrand`: deviceBrand}, bson.M{`area.$[]`: newAreaID})

	if nil != updatedModelAccount { // 若更新沒錯誤
		results = append(results, updatedModelAccount...) // 回傳結果
	}

	return
}

// UpdateOneArea - 更新帳戶驗證碼與有效時間
/**
 * @param primitive.M filter 過濾器
 * @param primitive.M update 更新
 * @return *mongo.UpdateResult returnUpdateResult 更新結果
 */
func (mongoDB *MongoDB) UpdateOneAccountVerificationCodeAndVerificationCodeValidPeriod(verificationCode string, validPeriod time.Time, userID string) (results []model.Account) {

	updatedModelAccount := mongoDB.findOneAndUpdateAccountSET(
		bson.M{
			`userID`: userID,
		},
		bson.M{
			// `userPassword`:         userPassword,
			// `verificationCodeTime`: validPeriod,
			`verificationCode`:            verificationCode,
			`verificationCodeValidPeriod`: validPeriod,
		},
	) // 更新的紀錄
	// fmt.Println("標記：", bson.M{`deviceID`: deviceID, `deviceBrand`: deviceBrand}, bson.M{`area.$[]`: newAreaID})

	if nil != updatedModelAccount { // 若更新沒錯誤
		results = append(results, updatedModelAccount...) // 回傳結果
	}

	return
}
