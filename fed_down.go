package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"sync"
)

type Config struct {
	List []DC `json:List`
}

type DC struct {
	Master   []System
	Slave    []System
	Gossiper []System
	Key_pem  string `json:Key_pem`
	Username string `json:Username`
}

type System struct {
	IsPublic bool
	Ip       string
}

func (conf *Config) Json_unmarshal_conf(file []byte) {
	err := json.Unmarshal(file, &conf)
	if err != nil {
		log.Fatal(err)
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < len(conf.List); i++ {
	wg.Add(1)
		go ssh_con(conf.List[i].Master[0].Ip,wg, "0", conf.List[i].Key_pem)
		go  ssh_con(conf.List[i].Slave[0].Ip, wg,"1", conf.List[i].Key_pem)
		go  ssh_con(conf.List[i].Gossiper[0].Ip,wg, "2", conf.List[i].Key_pem)
	}
	wg.Wait()
}
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
func ssh_con(publicIp string, wg *sync.WaitGroup, host string, path_pem string) {
	defer wg.Done()
	var b bytes.Buffer
	var inbyte bytes.Buffer
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			PublicKeyFile(path_pem),
		},
	}
	client, err := ssh.Dial("tcp", publicIp+":22", config)
	if err != nil {
		fmt.Println("Failed to dial: " + err.Error())
	}
	defer client.Close()

	session, err := client.NewSession()
    if err!=nil{
        log.Println("Creat session err",err)
    }

	session.Stdout = &b
	session.Stdin = &inbyte
	if host == "0" {
		fmt.Println("publicIp is\n", publicIp)
		session.Run("ps -ef | grep master | grep -v grep | awk '{print $2}' | xargs sudo kill -9\n")
	} else if host == "1" {
		fmt.Println("publicIp is\n", publicIp)
		session.Run("ps -ef | grep slave | grep -v grep | awk '{print $2}' | xargs sudo kill -9\n")
	} else if host == "2" {
		fmt.Println("publicIp is\n", publicIp)
		session.Run("ps -ef | grep gossiper | grep -v grep | awk '{print $2}' | xargs sudo kill -9 \n")
		//fmt.Printf("success: %v \n",b.String())
	}
    logfile,er := ioutil.
   	session.Wait()
	//session.Close()
}
func main() {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	config := Config{}
	config.Json_unmarshal_conf(file)
}
