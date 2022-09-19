package initialize

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// MustInitConfigAndWatch 第一次初始化 config 时必须成功，否则 panic;
func MustInitConfigAndWatch(configFileName string, config interface{}, watch func(config interface{}, err error)) {
	v, err := initConfig(configFileName, config)
	if err != nil {
		panic(err)
	}
	watchConfigFile(v, func(in fsnotify.Event) {
		err = readAndUnmarshalConfig(v, config)
		watch(config, err)
	})
}

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

func watchConfigFile(v *viper.Viper, run func(in fsnotify.Event)) {
	v.WatchConfig()
	v.OnConfigChange(run)
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
