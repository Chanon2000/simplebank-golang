package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	DBDriver		string        `mapstructure:"DB_DRIVER"` // ใช้ mapstructure tag เพื่อกำหนด name ของแต่ละ config field ใน viper
	DBSource		string        `mapstructure:"DB_SOURCE"`
	ServerAddress	string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"` // time.Duration ทำให้สามารถอ่าน value = 15m จาก env ได้
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path) // เพื่อบอก location ของ config file
	viper.SetConfigName("app") // เพื่อบอกชื่อของ file ของ config file
	viper.SetConfigType("env") // เพื่อบอกประเภทของ config file โดยในที่นี้ก็คือ env นั้นเอง // สามารถใช้ได้ทั้ง json, xml และอื่นๆเป็นต้น

	viper.AutomaticEnv()

	err = viper.ReadInConfig() 
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config) // viper ทำการใส่ค่าลง config variable
	return
}
