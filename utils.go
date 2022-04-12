package bytego

import (
	"path"
	"unsafe"
)

func joinPath(basePath, relativePath string) string {
	if relativePath == "" {
		return basePath
	}
	finalPath := path.Join(basePath, relativePath)
	if endWithChar(relativePath, '/') && !endWithChar(finalPath, '/') {
		return finalPath + "/"
	}
	return finalPath
}

func endWithChar(str string, c byte) bool {
	return str[len(str)-1] == c
}

// StringToBytes converts string to byte slice without a memory allocation.
func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
