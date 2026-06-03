package gstool

import "github.com/google/uuid"

func Uuid() string {
	return uuid.New().String()
}
func UuidPrefix(prefix string) string {
	return uuid.NewMD5(uuid.NameSpaceDNS, []byte(prefix)).String()
}
