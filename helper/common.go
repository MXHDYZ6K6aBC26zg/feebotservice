package helper

import (
	"github.com/satori/go.uuid"
	"fmt"
	"time"
	"strconv"
)

func GenerateUuid() string {
	id := uuid.NewV4()
	return id.String()
}

func GetTimeStamp() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%s",now.Format("20060102150405"))
}

func GetAmountByPercentage(payablePercentage, totalAmount float64) float64 {
	exactAmount := ((payablePercentage / 100) * totalAmount)
	return exactAmount
}

//GetType casts to propertype and asserts back to interface
func GetType(anything interface{}) interface{} {
	switch v := anything.(type) {
	case string:
		fmt.Println("value is a string:", v)
		return v
	case int32, int64:
		fmt.Println("value is an int32:", v)
		return v
	case []uint8:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			fmt.Println(err)
		}
		return f
	case float64:
		fmt.Println("value is float64")
	default:
		fmt.Println("unknown")
		return v
	}
	return nil
}
