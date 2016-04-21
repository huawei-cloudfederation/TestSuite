package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os/exec"
	//    "io"
	"io/ioutil"
	//  "os"
	//"time"
	"strings"
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
        var privateIp string
        err := json.Unmarshal(file, &conf)
        if err != nil {
                log.Fatal(err)
        }
        wg := new(sync.WaitGroup)
        for i := 0; i < len(conf.List); i++ {
                wg.Add(1)
                fmt.Println(conf.List[i].Master[0].Ip)
                str1 := [][]string{
                        []string{"ssh", "-i", conf.List[i].Key_pem, conf.List[i].Master[0].Ip, "hostname", "-i"},
                }
                privateIp = get_ip(str1)

                go create_hosts(conf.List[i].Master[0].Ip, wg, "0", privateIp, conf.List[i].Key_pem)
                fmt.Println(conf.List[i].Slave[0].Ip)
                go create_hosts(conf.List[i].Slave[0].Ip, wg, "1", privateIp, conf.List[i].Key_pem)
                fmt.Println(conf.List[i].Gossiper[0].Ip)
                go create_hosts(conf.List[i].Gossiper[0].Ip, wg, "2", privateIp, conf.List[i].Key_pem)
        }
        wg.Wait()
}

func get_ip(str [][]string) string {
        for _, val := range str {
                cmd := exec.Command(val[0], val[1:]...)
                fmt.Println("I am inside goroutine", val[0], val[1:])
                OP, err := cmd.Output()
                if err != nil {
                        fmt.Println(err.Error())
                } else {
                        log.Println("success", string(OP))
                }
                s := string(OP[:])
                s1 := strings.TrimSpace(s)
                fmt.Println(s1)
                return s1
        }
        return "nil"
}

func create_hosts(publicIp string, wg *sync.WaitGroup, host string,privateIp string, path_pem string) {
      Ip := "ubuntu@" + publicIp + ":/home/ubuntu"
      fmt.Println(Ip)
      Ip1 := "ubuntu@" + publicIp
      con_arr := [][]string{
              []string{"ssh", "-i", path_pem, "-yes", Ip1, "sudo rm -rf /home/ubuntu/src", ";", "rm /home/ubuntu/src.tar.gz"},
              []string{"scp", "-i", path_pem, "/home/ubuntu/src.tar.gz", Ip},
      }
      log.Println("create_hosts called")
      process_arr(con_arr)
      ssh_con(publicIp, host, privateIp, path_pem)

      wg.Done()
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
func ssh_con(publicIp string, host string, privateIp string, path_pem string) {

    var b bytes.Buffer
    var inbyte bytes.Buffer
    master := "sudo ./bin/mesos-master.sh --ip="+privateIp+" --work_dir=$HOME --modules='file://../../FedModules/FedModules.json' --allocator='mesos_fed_allocator_module' \n"
    slave := "sudo ./bin/mesos-slave.sh --master=" + privateIp +":5050 \n"
    gossiper := "./gossiper -config=first.json \n"
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			PublicKeyFile(path_pem),
		},
	}

      apt_arr := []string{
                "sudo apt-get install -y git",
                "sudo apt-get update",
                "sudo apt-get install -y tar",
                "sudo apt-get install -y wget",
                "sudo apt-get install -y openjdk-7-jdk",
                "sudo apt-get install -y g++",
                "sudo apt-get install -y zlib1g-dev",
                "sudo apt-get install -y autoconf",
                "sudo apt-get install -y libtool",
                "sudo apt-get install -y build-essential",
                "sudo apt-get install -y python-dev",
                "sudo apt-get install -y python-boto",
                "sudo apt-get install -y libcurl4-nss-dev",
                "sudo apt-get install -y libsasl2-dev",
                "sudo apt-get install -y libsasl2-modules",
                "sudo apt-get install -y maven",
                "sudo apt-get install -y libapr1-dev",
                "sudo apt-get install -y libsvn-dev",
                "sudo rm -rf go1.6.1.linux-amd64.tar.gz src",
                "sudo wget https://storage.googleapis.com/golang/go1.6.1.linux-amd64.tar.gz",
                "sudo tar -xvzf go1.6.1.linux-amd64.tar.gz",
                "echo export GOROOT=/home/ubuntu/go > .bash_profile",
                "echo export PATH='${GOROOT}/bin:${PATH}' >> .bash_profile",
                "echo export GOPATH=/home/ubuntu/gopkg >> .bash_profile",
                "chmod 777 .bash_profile",
                "./.bash_profile",
                "sudo tar -xzf /home/ubuntu/src.tar.gz -C /home/ubuntu",
        }

	client, err := ssh.Dial("tcp", publicIp+":22", config)
	if err != nil {
		fmt.Println("Failed to dial: " + err.Error())
	}
	defer client.Close()

      for _, value := range apt_arr {
                session, err := client.NewSession()
                if err != nil {
                        fmt.Println("Failed to create session: " + err.Error())
                }
                session.Stdout = &b
        if err := session.Run(value); err != nil {
             fmt.Println("Failed to run: " + err.Error())
        }
                fmt.Println(b.String())
                session.Close()
      }

		session, err := client.NewSession()
		session.Stdout = &b
        session.Stdin = &inbyte
		err = session.Shell()
        if host == "0" {
          _,err = inbyte.WriteString("cd /home/ubuntu/src/mesos/build\n")
            _, err = inbyte.WriteString(master)
    	} else if host == "1" {
          _,err = inbyte.WriteString("cd /home/ubuntu/src/mesos/build\n")
            _, err = inbyte.WriteString(slave)
    	} else if host == "2" {
          _,err = inbyte.WriteString("cd /home/ubuntu/src/gossiper\n")
//            _, err = inbyte.WriteString("./gossiper -config=first.json \n")
            _, err = inbyte.WriteString(gossiper)
          
	    }
		//fmt.Println("Failed to run: " + err.Error())
        session.Wait()
		fmt.Println(b.String())
		//session.Close()
		
}

func process_arr(str [][]string) {
        for _, val := range str {
                cmd := exec.Command(val[0], val[1:]...)
                fmt.Println("I am inside goroutine process_arr", val[0], val[1:])
                OP, err := cmd.Output()
                log.Println("process_arr OP frm the cmd is ", string(OP))
                if err != nil {
                        fmt.Println(err.Error())
                } else {
                        log.Println("success", string(OP))
                }
        }

}


func main() {
        var ver int
        file, err := ioutil.ReadFile("./config.json")
        if err != nil {
                log.Fatal(err)
        } else {
                fmt.Println("currently installed mesos git dir : https://github.com/apache/mesos.git")
                fmt.Println("press any key, if you want to use current mesos version,\n press 1, if you want to update the mesos with latest version")
                fmt.Scanf("%d", &ver)
                if ver == 1 {
                        //   fed_mesos()
                }
                config := Config{}
                config.Json_unmarshal_conf(file)
        }
}
