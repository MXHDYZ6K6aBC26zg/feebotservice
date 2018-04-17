package helper

import (
	"github.com/satori/go.uuid"
	"fmt"
	"time"
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
