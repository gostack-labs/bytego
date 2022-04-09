package bytego

import "path"

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
