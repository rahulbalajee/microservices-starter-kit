/*
Package env provides a simple way to get environment variables.
*/
package env

import (
	"log"
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("val %v cannot be converted to int %v\n", val, err)
		return fallback
	}

	return valAsInt
}

func GetFloat(key string, fallback float64) float64 {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsFloat, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("val %v cannot be converted to float %v\n", val, err)
		return fallback
	}

	return valAsFloat
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("val %v cannot be converted to bool %v\n", val, err)
		return fallback
	}

	return boolVal
}
