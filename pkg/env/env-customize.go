package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type TypeEnv int

const (
	INT = iota
	STRING
	FLOAT
	BOOL
)

type ConfigEnv struct {
	IsRequire bool
	Type      TypeEnv
}

func validateTypeEnv(value string, config ConfigEnv) bool {
	switch config.Type {
	case INT:
		_, err := strconv.Atoi(value)
		return err == nil
	case STRING:
		return true
	case FLOAT:
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case BOOL:
		return value == "true" || value == "false"
	}
	return false
}

func ValidateEnv(config map[string]ConfigEnv) {
	for k, v := range config {
		value, ok := os.LookupEnv(k)
		if !ok && v.IsRequire {
			log.Fatalf("Env required : %v", k)
		}

		if !validateTypeEnv(value, v) {
			log.Fatalf("Invalid type env: %v", k)
		}
	}
}

func mergeMaps[K comparable, V any](m1 map[K]V, m2 map[K]V) map[K]V {
	merged := make(map[K]V)
	for key, value := range m1 {
		merged[key] = value
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}

var managerEnv map[string]ConfigEnv = make(map[string]ConfigEnv)

func Load(path string, config ...map[string]ConfigEnv) {
	godotenv.Load(path)
	for _, v := range config {
		ValidateEnv(v)
		managerEnv = mergeMaps(managerEnv, v)
	}
}

func GetEnv(key string) interface{} {
	v, ok := managerEnv[key]
	if !ok {
		return nil
	}
	switch v.Type {
	case INT:
		value, _ := strconv.Atoi(os.Getenv(key))
		return value
	case STRING:
		return os.Getenv(key)
	case FLOAT:
		value, _ := strconv.ParseFloat(os.Getenv(key), 64)
		return value
	case BOOL:
		if os.Getenv(key) == "true" {
			return true
		} else {
			return false
		}
	}
	return nil
}
