package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func IsDebug() bool {
	debug := Getenv("DEBUG", "")
	return debug != ""
}

func LoadEnv() error {
	return godotenv.Load(".env")
}

func PrintEnv() {
	envMap, err := godotenv.Read()
	if err != nil {
		panic(err)
	}
	log.Println("Environment:")
	for k, v := range envMap {
		log.Printf("%s: %s\n", k, v)
	}
}

func Getenv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	return v
}
