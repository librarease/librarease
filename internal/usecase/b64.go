package usecase

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func UUIDToBase64(uuid string) (string, error) {
	uuid = removeDashes(uuid)
	bytes, err := hex.DecodeString(uuid)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func Base64ToUUID(b64 string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(bytes)
	return addDashes(hexStr), nil
}

func removeDashes(uuid string) string {
	return uuid[:8] + uuid[9:13] + uuid[14:18] + uuid[19:23] + uuid[24:]
}

func addDashes(hexStr string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:])
}

// func main() {
// originalUUID := "123e4567-e89b-12d3-a456-426614174000"
// b64, _ := UUIDToBase64(originalUUID)
// fmt.Println("Base64:", b64)
//
// reconstructedUUID, _ := Base64ToUUID(b64)
// fmt.Println("UUID:", reconstructedUUID)
// }
