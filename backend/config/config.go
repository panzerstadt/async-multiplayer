package config

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.

type Config struct {
	GoogleOauthClientID     string `mapstructure:"GOOGLE_OAUTH_CLIENT_ID"`
	GoogleOauthClientSecret string `mapstructure:"GOOGLE_OAUTH_CLIENT_SECRET"`
	GoogleOauthRedirectUrl  string `mapstructure:"GOOGLE_OAUTH_REDIRECT_URL"`
	JwtSecret               string `mapstructure:"JWT_SECRET"`
	FrontendUrl             string `mapstructure:"FRONTEND_URL"`
	MailgunAPIKey           string `mapstructure:"MAILGUN_API_KEY"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	return
}
