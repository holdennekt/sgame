package envvar

import (
	"log"
	"os"
	"strconv"
)

func GetEnvVar(name string) string {
	variable, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("%s env variable is missing", name)
	}
	return variable
}

func GetEnvVarInt(name string) int {
	variable := GetEnvVar(name)
	num, err := strconv.Atoi(variable)
	if err != nil {
		log.Fatalf("%s env variable is not int", name)
	}
	return num
}
