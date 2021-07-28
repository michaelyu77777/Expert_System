package databases

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"leapsy.com/packages/configurations"
	"leapsy.com/packages/network"
)

// MongoDB - 資料庫
type MongoDB struct {
}

// GetConfigValueOrPanic - 取得設定值否則結束程式
/**
 * @param  string key  關鍵字
 * @return string 設定資料區塊下關鍵字對應的值
 */
func (mongoDB MongoDB) GetConfigValueOrPanic(sectionName string, key string) string {
	return configurations.GetConfigValueOrPanic(sectionName, key)
}

// GetConfigPositiveIntValueOrPanic - 取得正整數設定值否則結束程式
/**
 * @param  string key  關鍵字
 * @return string 設定資料區塊下關鍵字對應的正整數值
 */
func (mongoDB MongoDB) GetConfigPositiveIntValueOrPanic(sectionName string, key string) int {
	return configurations.GetConfigPositiveIntValueOrPanic(sectionName, key)
}

// Connect - 連接資料庫
/**
 * @return *mongo.Client mongoClientPointer 資料庫客戶端指標
 */
func (mongoDB *MongoDB) Connect() (returnMongoClientPointer *mongo.Client) {

	// 預設主機
	address := fmt.Sprintf(
		`%s:%d`,
		mongoDB.GetConfigValueOrPanic(`mongoDB`, `server`),
		mongoDB.GetConfigPositiveIntValueOrPanic(`mongoDB`, `port`),
	)

	network.SetAddressAlias(address, `專家系統資料庫`) // 設定預設主機別名

	connectionString := fmt.Sprintf(
		`mongodb://%s`,
		address,
	)

	// 連接預設主機
	mongoClientPointer, mongoConnectError := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))

	go logger.Errorf(`%s %s 連接`, network.GetAliasAddressPair(address), mongoConnectError)

	// logings.SendLog(
	// 	[]string{`%s %s 連接`},
	// 	network.GetAliasAddressPair(address),
	// 	mongoConnectError,
	// 	logrus.ErrorLevel,
	// )

	if nil != mongoConnectError { // 若連接預設主機錯誤
		return // 回傳
	}

	mongoClientPointerPingError := mongoClientPointer.Ping(context.TODO(), nil) // 確認主機可連接

	// 取得記錄器格式和參數

	go logger.Errorf(`%s %s 確認可連接`, network.GetAliasAddressPair(address), mongoClientPointerPingError)

	// logings.SendLog(
	// 	[]string{`%s %s 確認可連接`},
	// 	network.GetAliasAddressPair(address),
	// 	mongoClientPointerPingError,
	// 	logrus.ErrorLevel,
	// )

	if nil != mongoClientPointerPingError { // 若確認主機可連接錯誤
		return // 回傳
	}

	returnMongoClientPointer = mongoClientPointer // 回傳資料庫指標

	return // 回傳
}

// Disconnect - 中斷與資料庫的連線
/*
 * @params *mongo.Client mongoClientPointer 資料庫指標
 */
func (mongoDB *MongoDB) Disconnect(mongoClientPointer *mongo.Client) {

	if nil != mongoClientPointer { // 若客戶端指標不為空

		// 預設主機
		address := fmt.Sprintf(
			`%s:%d`,
			mongoDB.GetConfigValueOrPanic(`mongoDB`, `server`),
			mongoDB.GetConfigPositiveIntValueOrPanic(`mongoDB`, `port`),
		)

		mongoDBClientDisconnectError := mongoClientPointer.Disconnect(context.TODO()) // 斷接主機

		go logger.Errorf(`%s %s 斷接`, network.GetAliasAddressPair(address), mongoDBClientDisconnectError)

		// logings.SendLog(
		// 	[]string{`%s %s 斷接`},
		// 	network.GetAliasAddressPair(address),
		// 	mongoDBClientDisconnectError,
		// 	logrus.ErrorLevel,
		// )

		if nil != mongoDBClientDisconnectError { // 若斷接失敗
			return // 回傳
		}

	}

}
