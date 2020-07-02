package master

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiPort               int      `json:"apiPort"`
	ApiReadTimeout        int      `json:"apiReadTimeout"`
	ApiWriteTimeout       int      `json:"apiWriteTimeout"`
	EtcdEndPoints         []string `json:"etcdEndPoints"`
	EtcdDialTimeout       int      `json:"etcdDialTimeout"`
	WebRoot               string   `json:"webroot"`
	MongoDbURI            string   `json:"mongodbURI"`
	MongodbConnectTimeout int64    `json:"mongodbTimeout"`
	MongoDatabase         string   `json:"mongo_database"`
	MongoCollection       string   `json:"mongo_collection"`
}

var (
	G_config *Config
)

func InitConfig(filename string) (err error) {
	var (
		context []byte
		conf    Config
	)

	if context, err = ioutil.ReadFile(filename); err != nil {
		return err
	}
	if err = json.Unmarshal(context, &conf); err != nil {
		return err
	}
	G_config = &conf
	return

}
