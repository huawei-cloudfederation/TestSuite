package main

import (
	"./auto_update"
	"fmt"
	"log"
	"io/ioutil"
	"os"
)

func main() {
	var ver int
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("currently installed mesos git dir : https://github.com/apache/mesos.git")
		fmt.Println("press any key, if you want to use current mesos version,\n press 1, if you want to update the mesos with latest version")
		fmt.Scanf("%d", &ver)
		if ver == 1 {
			auto_update.Fedmesos()
		}
		config := auto_update.Config{}
		config.Json_unmarshal_conf(file)
	}
}
