package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	EtcdEndPoints         []string `json:"etcdEndPoints"`
	EtcdDialTimeout       int      `json:"etcdDialTimeout"`
	MongoDbURI            string   `json:"mongodbURI"`
	MongodbConnectTimeout int64    `json:"mongodbTimeout"`
	JogLogBatchSize       int      `json:"jogLogBatchSize"`
	JobLogCommitTimeout   int      `json:"jobLogCommitTimeout"`
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
