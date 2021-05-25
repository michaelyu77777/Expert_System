package jwts

import (

	//"errors"
	//"sync"
	//"time"

	"github.com/dgrijalva/jwt-go"
	//"github.com/google/uuid"
	//"github.com/icrowley/fake"
	//"github.com/robfig/cron"
	//"../configurations"
)

var (
	secretByteArray = getNewSecretByteArray()
)

const key_AES = "RB7Wfa$WHssV4LZce6HCyNYpdPPlYnDn" //固定 KEY 加密解密 32位數

// getSecretByteArray - 取得密鑰
func getSecretByteArray() []byte {
	//secretByteArrayRWMutex.RLock()
	gotSecretByteArray := secretByteArray
	//secretByteArrayRWMutex.RUnlock()
	return gotSecretByteArray
}

// getNewSecretByteArray - 取得新密鑰
/*
 * @return []byte result 結果
 */
func getNewSecretByteArray() (result []byte) {

	//result = []byte(base64.StdEncoding.EncodeToString([]byte(`新的key`)))
	//result = []byte(base64.StdEncoding.EncodeToString([]byte(key_AES)))
	result = []byte(key_AES)
	return
}

// type Token struct {
// 	Data string
// }

type TokenInfo struct {
	Data string
}

// CreateToken - 產生令牌
/*
 * @params TokenInfo tokenInfo 令牌資訊
 * @return *string returnTokenStringPointer 令牌字串指標
 */
func CreateToken(tokenInfoPointer *TokenInfo) (returnTokenStringPointer *string) {

	if nil != tokenInfoPointer {
		claim := jwt.MapClaims{
			`Data`: tokenInfoPointer.Data,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
		tokenString, tokenSignedStringError := token.SignedString(getSecretByteArray())

		// logings.SendLog(
		// 	[]string{`簽署令牌字串 %s `},
		// 	[]interface{}{tokenString},
		// 	tokenSignedStringError,
		// 	logrus.InfoLevel,
		// )

		if nil == tokenSignedStringError {
			returnTokenStringPointer = &tokenString
		}

	}

	return
}

func getSecretFunction() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		return getSecretByteArray(), nil
	}
}

// ParseToken - 解析令牌
/*
 * @params string tokenString 令牌字串
 * @return *TokenInfo tokenInfoPointer 令牌資訊指標
 */
func ParseToken(tokenString string) (tokenInfoPointer *TokenInfo) {

	if tokenString != `` {

		//defaultFormatSlices := []string{`解析令牌 %s `}
		//dafaultArgs := []interface{}{tokenString}

		token, jwtParseError := jwt.Parse(tokenString, getSecretFunction())

		// logings.SendLog(
		// 	[]string{`解析令牌 %s `},
		// 	dafaultArgs,
		// 	jwtParseError,
		// 	logrus.InfoLevel,
		// )

		if jwtParseError != nil {
			return
		}

		mapClaim, ok := token.Claims.(jwt.MapClaims)

		if !ok {

			// logings.SendLog(
			// 	defaultFormatSlices,
			// 	dafaultArgs,
			// 	errors.New(`將Claim轉成MapClaim錯誤`),
			// 	logrus.InfoLevel,
			// )

			return
		}

		// 驗證token，如果token被修改過則為false
		if !token.Valid {

			// logings.SendLog(
			// 	defaultFormatSlices,
			// 	dafaultArgs,
			// 	errors.New(`無效令牌錯誤`),
			// 	logrus.InfoLevel,
			// )

			return
		}

		dataValue, dataOK := mapClaim[`Data`].(string)

		if dataOK {
			tokenInfoPointer = &TokenInfo{
				Data: dataValue,
			}
		}

	}

	return
}
