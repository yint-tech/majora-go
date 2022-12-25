package initialize

import (
	"github.com/spf13/viper"
)

// MustInitConfig 初始化 config 时必须成功，否则 panic;
func MustInitConfig(configFileName string, config interface{}) {
	_, err := initConfig(configFileName, config)
	if err != nil {
		panic(err)
	}
}

func initConfig(configFileName string, config interface{}) (*viper.Viper, error) {
	if configFileName == "" {
		configFileName = "./conf/majora-dev.yaml"
	}
	v := viper.New()
	v.SetConfigFile(configFileName)
	err := readAndUnmarshalConfig(v, config)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func readAndUnmarshalConfig(v *viper.Viper, config interface{}) error {
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	err = v.Unmarshal(config)
	if err != nil {
		return err
	}
	return nil
}
