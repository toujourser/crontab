package master

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeout  int      `json:"apiReadTimeout"`
	ApiWriteTimeout int      `json:"apiWriteTimeout"`
	EtcdEndPoints   []string `json:"etcdEndPoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
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
