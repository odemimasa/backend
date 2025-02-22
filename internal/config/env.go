package config

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	envOnce sync.Once
	err     error
	Env     struct {
		DATABASE_URL            string
		TWILIO_ACCOUNT_SID      string
		TWILIO_AUTH_TOKEN       string
		REDIS_URL               string
		TRIPAY_MERCHANT_CODE    string
		TRIPAY_API_KEY          string
		TRIPAY_PRIVATE_KEY      string
		AUTHORIZED_EMAILS       string
		ACCESS_TOKEN_SECRET_KEY string
		ASYNQMON_BASE_URL       string
	}
)

func LoadEnv() error {
	envOnce.Do(func() {
		err = godotenv.Load()
		if err != nil {
			return
		}

		Env.DATABASE_URL = os.Getenv("DATABASE_URL")
		Env.TWILIO_ACCOUNT_SID = os.Getenv("TWILIO_ACCOUNT_SID")
		Env.TWILIO_AUTH_TOKEN = os.Getenv("TWILIO_AUTH_TOKEN")
		Env.REDIS_URL = os.Getenv("REDIS_URL")
		Env.TRIPAY_MERCHANT_CODE = os.Getenv("TRIPAY_MERCHANT_CODE")
		Env.TRIPAY_API_KEY = os.Getenv("TRIPAY_API_KEY")
		Env.TRIPAY_PRIVATE_KEY = os.Getenv("TRIPAY_PRIVATE_KEY")
		Env.AUTHORIZED_EMAILS = os.Getenv("AUTHORIZED_EMAILS")
		Env.ACCESS_TOKEN_SECRET_KEY = os.Getenv("ACCESS_TOKEN_SECRET_KEY")
		Env.ASYNQMON_BASE_URL = os.Getenv("ASYNQMON_BASE_URL")
	})

	return err
}
