package app



import (
	"delay-queue/options"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"reflect"
)

func AddConfigToOptions(options *options.Options) error {
	viper.SetConfigName("config")
	viper.AddConfigPath("config/")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	optDecode := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeDurationHookFunc(), StringToByteSizeHookFunc()))

	err = viper.Unmarshal(options, optDecode)
	fmt.Println(options)
	if err != nil {
		return err
	}
	return nil
}

func StringToByteSizeHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type,
		t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(datasize.ByteSize(5)) {
			return data, nil
		}
		raw := data.(string)
		result := new(datasize.ByteSize)
		result.UnmarshalText([]byte(raw))
		return result.Bytes(), nil
	}
}