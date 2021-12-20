package fileCache

import (
	"github.com/PeterYangs/tools/secret"
	"io/ioutil"
	"os"
	"time"
)

type fileCache struct {
}

func NewFileCache() *fileCache {

	return &fileCache{}
}

func (f fileCache) Put(key string, value string, ttl time.Duration) error {
	//TODO implement me
	//panic("implement me")

	//secret.AesEncryptCFB()

	fileName := secret.AesEncryptCFB([]byte(key), []byte(os.Getenv("KEY")))

	panic(fileName)

	file, err := os.OpenFile("storage/"+string(fileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)

	if err != nil {

		return err
	}

	defer file.Close()

	_, err = file.Write([]byte(value))

	return err

}

func (f fileCache) Get(key string) (string, error) {
	//TODO implement me
	//panic("implement me")

	fileName := secret.AesEncryptCFB([]byte(key), []byte(os.Getenv("KEY")))

	file, err := os.Open(string(fileName))

	if err != nil {

		return "", err
	}

	data, err := ioutil.ReadAll(file)

	return string(data), err
}
