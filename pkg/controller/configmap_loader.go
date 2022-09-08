package controller

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func UnmarshalConf(configmapPath string) map[string][]string {
	yfile, err := ioutil.ReadFile(configmapPath)
	if err != nil {
		log.Println(err)
	}

	conf := make(map[string][]string)

	err2 := yaml.Unmarshal(yfile, &conf)
	if err2 != nil {
		log.Println(err2)
	}

	return conf
}
