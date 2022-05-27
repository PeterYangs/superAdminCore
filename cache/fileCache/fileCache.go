package fileCache

import (
	"encoding/json"
	"errors"
	"github.com/PeterYangs/superAdminCore/v2/conf"
	"github.com/PeterYangs/tools"
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

type structure struct {
	Data   string `json:"data"`
	Expire int    `json:"expire"`
}

func (f *fileCache) Put(key string, value string, ttl time.Duration) error {
	//TODO implement me

	d := secret.NewDes()

	fileName, err := d.Encyptog3DES([]byte(key), []byte(os.Getenv("KEY")))

	if err != nil {

		return err
	}

	fileName_ := string(fileName.ToBase64())

	dirName := tools.SubStr(fileName_, 0, 2) + "/" + tools.SubStr(fileName_, 2, 2)

	//生成文件夹
	os.MkdirAll(conf.Get("file_cache_path").(string)+"/"+dirName, 0755)

	file, err := os.OpenFile(conf.Get("file_cache_path").(string)+"/"+dirName+"/"+fileName_, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)

	if err != nil {

		return err
	}

	defer file.Close()

	var v structure

	v.Data = value

	if ttl == 0 {

		v.Expire = 0
	} else {

		v.Expire = int(time.Now().Add(ttl).Unix())
	}

	jsonStr, err := json.Marshal(v)

	if err != nil {

		return err
	}

	_, err = file.Write(jsonStr)

	return err

}

func (f *fileCache) Get(key string) (string, error) {
	//TODO implement me

	d := secret.NewDes()

	fileName, err := d.Encyptog3DES([]byte(key), []byte(os.Getenv("KEY")))

	if err != nil {

		return "", err
	}

	fileName_ := string(fileName.ToBase64())

	dirName := tools.SubStr(fileName_, 0, 2) + "/" + tools.SubStr(fileName_, 2, 2)

	file, err := os.Open(conf.Get("file_cache_path").(string) + "/" + dirName + "/" + fileName_)

	if err != nil {

		return "", err
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)

	if err != nil {

		return "", err
	}

	var v structure

	err = json.Unmarshal(data, &v)

	if err != nil {

		return "", err
	}

	now := time.Now().Unix()

	//判断过期时间
	if v.Expire != 0 && (now > int64(v.Expire)) {

		//删除文件
		defer func(f *os.File, ff string) {

			f.Close()

			os.Remove(ff)

		}(file, conf.Get("file_cache_path").(string)+"/"+dirName+"/"+fileName_)

		return "", errors.New("缓存已超时")
	}

	return v.Data, err
}

func (f *fileCache) Remove(key string) error {

	d := secret.NewDes()

	fileName, err := d.Encyptog3DES([]byte(key), []byte(os.Getenv("KEY")))

	if err != nil {

		return err
	}

	fileName_ := string(fileName.ToBase64())

	dirName := tools.SubStr(fileName_, 0, 2) + "/" + tools.SubStr(fileName_, 2, 2)

	os.Remove(conf.Get("file_cache_path").(string) + "/" + dirName + "/" + fileName_)

	return nil
}

func (f *fileCache) Exists(key string) bool {

	d := secret.NewDes()

	fileName, err := d.Encyptog3DES([]byte(key), []byte(os.Getenv("KEY")))

	if err != nil {

		return false
	}

	fileName_ := string(fileName.ToBase64())

	dirName := tools.SubStr(fileName_, 0, 2) + "/" + tools.SubStr(fileName_, 2, 2)

	file, err := os.Open(conf.Get("file_cache_path").(string) + "/" + dirName + "/" + fileName_)

	if err != nil {

		return false
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)

	if err != nil {

		return false
	}

	var v structure

	err = json.Unmarshal(data, &v)

	if err != nil {

		return false
	}

	now := time.Now().Unix()

	//判断过期时间
	if v.Expire != 0 && (now > int64(v.Expire)) {

		//删除文件
		defer func(f *os.File, ff string) {

			f.Close()

			os.Remove(ff)

		}(file, conf.Get("file_cache_path").(string)+"/"+dirName+"/"+fileName_)

		return false
	}

	return true

}
