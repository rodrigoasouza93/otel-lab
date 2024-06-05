package configs

import (
	"os"

	"github.com/spf13/viper"
)

type Cfg struct {
	WeatherAPIKey string `mapstructure:"WEATHER_API_KEY"`
}

func LoadConfig(path string) *Cfg {
	var cfg Cfg
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if err != err.(*os.PathError) {
			return nil
		}
		cfg = Cfg{
			WeatherAPIKey: os.Getenv("WEATHER_API_KEY"),
		}
	} else {
		err = viper.Unmarshal(&cfg)
		if err != nil {
			panic(err)
		}
	}
	return &cfg
}
