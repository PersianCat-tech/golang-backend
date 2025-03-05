package util

import (
	"math/rand"
	"strings"
	//"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// func init() {
// 	rand.Seed(time.Now().UnixNano())
// }

//RandomInt generate a random integer between min and max
func RandomInt(min, max int64) int64 {
	return  min + rand.Int63n(max - min + 1)
} 

//generate a random string of length n
func RandomString(n int) string {
	var sb strings.Builder	//字符串生成对象
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

//generate a random Owner name
func RandomOwner() string {
	return RandomString(6) 
}

//generate a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

//generate a random currency code
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)

	return currencies[rand.Intn(n)]
}
