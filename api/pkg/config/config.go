// Package config will create the configuration object
package config

import (
	"github.com/spf13/viper"
)

// NewConfig Create a new configuration
func NewConfig(filename string, defaults map[string]interface{}) (*viper.Viper, error) {
	viper := viper.New()
	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
	viper.SetConfigName(filename)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// viper.WatchConfig()
	// viper.OnConfigChange(func(e fsnotify.Event) {
	// 	fmt.Println("Config file changed:", e.Name)
	// })
	return viper, nil
}
