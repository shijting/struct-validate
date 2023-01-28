package utils

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

func UnderscoreName(name string) string {
	bs := new(bytes.Buffer)
	for i, v := range name {
		if unicode.IsUpper(v) {
			if i != 0 {
				bs.WriteString("_")
			}
			bs.WriteRune(unicode.ToLower(v))
		} else {
			bs.WriteRune(v)
		}
	}
	return bs.String()
}

func ScanFiles(src string) ([]string, error) {
	srcFiles := make([]string, 0, 10)
	src, _ = filepath.Abs(src)

	files, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") &&
			!strings.HasSuffix(file.Name(), "_validate.go") &&
			!strings.HasSuffix(file.Name(), "_test.go") {
			src, err = filepath.Abs(src)
			if err != nil {
				return nil, err
			}
			srcFiles = append(srcFiles, filepath.Join(src, file.Name()))
		}
	}
	return srcFiles, nil
}

func FileIsExist(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}

func GetWorkDirectory() (string, error) {
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("请确保已经安装git， 错误： %s", err)
	}
	return filepath.Clean(strings.TrimSpace(string(cmdOut))), nil
}

func RandString(size int) string {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func GetModule() (string, error) {
	wd, err := GetWorkDirectory()
	if err != nil {
		return "", err
	}
	// mod
	f, err := os.Open(filepath.Join(wd, "go.mod"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		lineArr := strings.Split(strings.Trim(string(line), " "), " ")

		if len(lineArr) == 2 && lineArr[0] == "module" {
			return lineArr[1], nil
		}

	}
	return "", errors.New("无法解析go.mod")
}
