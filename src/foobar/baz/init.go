package baz

import (
	"log"
	envs "github.com/Gamma169/go-server-helpers/environments"
)


var bazId string

func InitBaz() {
	log.Println("Initializing Baz package")

	bazId = envs.GetRequiredEnv("BAZ_ID")
}

func CheckRequiredEnvs() {
	envs.GetRequiredEnv("BAZ_ID")
}
