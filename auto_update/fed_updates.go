package auto_update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"io/ioutil"
	"os"
	"sync"
	"time"
	"strconv"
    	"../common"
)

type GossiperConfig struct {
	Name           string //Name of this gossiper
	City           string
	Country        string
	MasterEndPoint string //MesosMaster's IP address
	ConfigType     string //What type of gossiper is this ? Join a federation or start a federation? Values = JOIN or NEW
	JoinEndPoint   string //If we are joining an already runnig federatoin the what is the EndPoint?
	TCPPort        string //TCP port at which gossiper will bind and listen for anon module to connect to
	GPort          int    //Port at which gossper should start
	AdvertiseAddr  string //The advertised address of the gossiper so that other gossipers coudl connect
	ConsulConfig     ConsulConfig  `json:ConsulConfig`
}

//global consul config
type ConsulConfig struct {
	IsLeader    bool
	DCEndpoint  string
	StorePreFix string
	DCName      string
}

type ConsulConfigfirst struct {
	Datacenter  *string  `json:"datacenter"`
	Data_dir          *string `json:"data_dir"`
	Log_level          *string `json:"log_level"`
	Server            *bool   `json:"server"`
	Bootstrap          *bool   `json:"bootstrap"`
	Start_join         *[]string    `json:"start_join"`
	Bind_addr          *string       `json:"bind_addr"`
	Ad                 *Addresses     `json:"addresses"`
        Retry_join_wan *[]string `json:"retry_join_wan,omitempty"`
        Retry_interval_wan *string `json:"retry_interval_wan,omitempty"`
	AdvAdd             *Advertise_addrs `json:"advertise_addrs"`
}

type Addresses struct {
	Http string    `json:"http"`
}

type Advertise_addrs struct {
	Serf_lan string   `json:"serf_lan"`
	Serf_wan string   `json:"serf_wan"`
	Rpc      string   `json:"rpc"`
}

type Config struct {
	List []DC `json:List`
}

type DC struct {
	DC_id    int
	Master   []System
	Slave    []System
	Gossiper []System
	Consul   []System
	Country  string
	City     string
	Key_pem  string `json:Key_pem`
	Username string `json:Username`
}

type System struct {
	IsPublic bool
	Ip       string
}

func ConsulMarshal(id int, fConsulIp string, publicIp string, privateIp string) {
	name := strconv.Itoa(id)
	path := "/home/ubuntu/DC" + name + "/consul.json"
	var conf ConsulConfigfirst
	datacenter := name
	data_dir :=  "/home/ubuntu/fedCloud/consul"
	log_level := "INFO"
	server := true
	bootstrap := true
	start_join := []string{privateIp}
	bind_addr := privateIp
	addresses := Addresses{Http: privateIp}
	advertise_addrs := Advertise_addrs{Serf_lan: publicIp + ":8301", Serf_wan: publicIp + ":8302", Rpc: publicIp + ":8303"}	
	
	conf.Datacenter = &datacenter
        conf.Data_dir = &data_dir
        conf.Log_level = &log_level
        conf.Server = &server
        conf.Bootstrap = &bootstrap
        conf.Start_join = &start_join
        conf.Bind_addr = &bind_addr
        conf.Ad = &addresses
        conf.AdvAdd =&advertise_addrs

	
      if name != "1" {
		retry_interval_wan := "5s"
		retry_join_wan := []string{fConsulIp}
		conf.Retry_join_wan =&retry_join_wan
		conf.Retry_interval_wan =&retry_interval_wan
	}
		
		list, _ := json.MarshalIndent(conf, " ", "  ")
		err := ioutil.WriteFile(path, list, 0644)
		if err != nil {
			fmt.Printf("WriteFile json Error: %tv", err)
		}
}

func GossiperMarshal(id int, mPublicIp string, publicIp string, fGossiperIp string, consulIp string, city string, country string) {
	name := strconv.Itoa(id)
	path := "/home/ubuntu/DC" + name + "/gossiper.json"
	var conf *GossiperConfig
	if name == "1" {
		conf = &GossiperConfig{Name: name, ConfigType: "New", GPort: 4400, MasterEndPoint: mPublicIp + ":5050", AdvertiseAddr: publicIp, Country: country, City: city, ConsulConfig: ConsulConfig{IsLeader: true, DCEndpoint: consulIp + ":8500", StorePreFix: "Federa",DCName: name}}
	} else {

		conf = &GossiperConfig{Name: name, ConfigType: "Old", JoinEndPoint: fGossiperIp+":4400", GPort: 4400, MasterEndPoint: mPublicIp +":5050", AdvertiseAddr: publicIp, Country: country, City: city, ConsulConfig: ConsulConfig{IsLeader: false, DCEndpoint: consulIp + ":8500", StorePreFix: "Federa",DCName: name}}

	}
	list, _ := json.MarshalIndent(conf, " ", "  ")

	err := ioutil.WriteFile(path, list, 0644)
	if err != nil {
		fmt.Printf("WriteFile json Error: %tv", err)
	}
}

func (conf *Config) Json_unmarshal_conf(file []byte) {
	err := json.Unmarshal(file, &conf)
	if err != nil {
		log.Fatal(err)
	}
	wg := new(sync.WaitGroup)
	f, _ := os.Create("/home/ubuntu/test/log.txt")
	log.SetOutput(f)
	defer f.Close()
	for i := 0; i < len(conf.List); i++ {
		wg.Add(1)
		path_dir := "/home/ubuntu/fed/" + "DC" + strconv.Itoa(conf.List[i].DC_id)
		fmt.Println(path_dir)
		err := os.MkdirAll(path_dir, 0777)
		if err != nil {
			fmt.Println("error in creating directory \n", err)
		}
		fmt.Println("log file is creating\n")
		log.Printf("=============Data Center Id : %d ================\n", conf.List[i].DC_id)
		go create_hosts(conf.List[i].Master[0].Ip,wg,"4",conf.List[i], 0)
		go create_hosts(conf.List[i].Consul[0].Ip, wg, "3", conf.List[i],0)
		go create_hosts(conf.List[i].Master[0].Ip, wg, "0", conf.List[i],0)

		for j := 0; j < len(conf.List[i].Slave); j++ {
		go create_hosts(conf.List[i].Slave[j].Ip, wg, "1", conf.List[i],j)
		}
		go create_hosts(conf.List[i].Gossiper[0].Ip, wg, "2", conf.List[i],0)
	}
	wg.Wait()
}

func create_hosts(publicIp string, wg *sync.WaitGroup, host string, list DC, index int) {
	/*var time_start, time_end time.Time
	Ip := "ubuntu@" + publicIp + ":/home/ubuntu"
	fmt.Println(Ip)
	Ip1 := "ubuntu@" + publicIp
	con_arr := [][]string{
		[]string{"ssh", "-i", list.Key_pem, "-o", "StrictHostKeyChecking=no", Ip1, "sudo rm -rf /home/ubuntu/fedCloud", ";", "rm /home/ubuntu/fedCloud.tar.gz"},
		[]string{"scp", "-i", list.Key_pem, "-o", "StrictHostKeyChecking=no", "/home/ubuntu/fedCloud.tar.gz", Ip},
	}
	log.Println("create_hosts called")
	fmt.Println("Waiting for cleaning the VMs and copying the new Mesos_Federation Version...")
	time_start = time.Now()
	common.ParseCmd(con_arr)
	time_end = time.Now()
	common.GetTime(time_start, time_end)

	ssh_con(publicIp, host, list.Key_pem,list.DC_id)*/
   	up_connection(publicIp,host,list,index)

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
func ssh_con(publicIp string, host string, path_pem string, id int) {
	var time_start, time_end time.Time
	var b bytes.Buffer

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
		"sudo rm -rf go1.6.1.linux-amd64.tar.gz ",
		"sudo wget https://storage.googleapis.com/golang/go1.6.1.linux-amd64.tar.gz",
		"sudo tar -xzf go1.6.1.linux-amd64.tar.gz",
		"echo export GOROOT=/home/ubuntu/go > .bash_profile",
		"echo export PATH='${GOROOT}/bin:${PATH}' >> .bash_profile",
		"echo export GOPATH=/home/ubuntu/gopkg >> .bash_profile",
		"chmod 777 .bash_profile",
		"./.bash_profile",
		"sudo tar -xzf /home/ubuntu/fedCloud.tar.gz -C /home/ubuntu",
	}

	 config := &ssh.ClientConfig{
                User: "ubuntu",
                Auth: []ssh.AuthMethod{
                        PublicKeyFile(path_pem),
                },
    }
    client,err := ssh.Dial("tcp", publicIp+":22", config)
    if err != nil {
        log.Printf("Failed to dial: " + err.Error())
    }
    defer client.Close()

	fmt.Println("Waiting for Installations, updates into the VM...")
	time_start = time.Now()
	for _, value := range apt_arr {
		session, err := client.NewSession()
		if err != nil {
			log.Printf("Failed to create session: " + err.Error())
		}
		session.Stdout = &b
		if err := session.Run(value); err != nil {
			log.Printf("Failed to run: " + err.Error())
		}
		log.Printf(b.String())
		session.Close()
	}
	time_end = time.Now()
	common.GetTime(time_start, time_end)
 }
func up_connection(publicIp string, host string,list DC,index int) {
	var inbyte bytes.Buffer
	var path, privateIp string

	 config := &ssh.ClientConfig{
                User: "ubuntu",
                Auth: []ssh.AuthMethod{
                        PublicKeyFile(list.Key_pem),
                },
    }
    client,err := ssh.Dial("tcp", publicIp+":22", config)
    if err != nil {
        log.Printf("Failed to dial: " + err.Error())
    }
    defer client.Close()

	session, err := client.NewSession()
//	defer session.Close()

	if host == "0" {
		str1 := [][]string{
			[]string{"ssh", "-i", list.Key_pem, "-o", "StrictHostKeyChecking=no", publicIp, "hostname", "-i"},
		}
		privateIp = common.GetPrivateIp(str1)
		fmt.Println("I am inside", privateIp, "\n")
	    master := "sudo ./bin/mesos-master.sh --ip=" + privateIp + " --work_dir=$HOME --advertise_ip=" + publicIp + " --advertise_port=5050 --modules='file://../../FedModules/FedModules.json' --allocator='mesos_fed_allocator_module' \n"
		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(list.DC_id) + "/master" + strconv.Itoa(index) + ".log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
		session.Stdin = &inbyte
		err = session.Shell()
		_, err = inbyte.WriteString("cd /home/ubuntu/fedCloud/mesos/build\n")
		_, err = inbyte.WriteString(master)
	} else if host == "1" {
	    slave := "sudo ./bin/mesos-slave.sh --master=" + list.Master[0].Ip + ":5050 \n"
		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(list.DC_id) + "/slave" + strconv.Itoa(index) +  ".log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
		session.Stdin = &inbyte
                err = session.Shell()

		_, err = inbyte.WriteString("cd /home/ubuntu/fedCloud/mesos/build\n")
		_, err = inbyte.WriteString(slave)
	} else if host == "2" {
	    gossiper := "./gossiper -config=config.json \n"
		GossiperMarshal(list.DC_id, list.Master[0].Ip, publicIp, list.Gossiper[0].Ip, list.Consul[0].Ip, list.City, list.Country)
		gIp := "ubuntu@" + publicIp + ":/home/ubuntu/fedCloud/gossiper/config.json"
		name := strconv.Itoa(list.DC_id) 
		configPath := "/home/ubuntu/DC" + name + "/gossiper.json"

		con_arr1 := [][]string{
			[]string{"scp", "-i", list.Key_pem, "-o", "StrictHostKeyChecking=no", configPath, gIp},
		}
		common.ParseCmd(con_arr1)

		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(list.DC_id) + "/gossiper" + strconv.Itoa(index) + ".log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
		session.Stdin = &inbyte
                err = session.Shell()

		_, err = inbyte.WriteString("cd /home/ubuntu/fedCloud/gossiper\n")
		_, err = inbyte.WriteString(gossiper)
	} else if host == "3" {
	    consul := "sudo ./consul agent -config-file=/home/ubuntu/fedCloud/consul/config.json \n"
		str1 := [][]string{
			[]string{"ssh", "-i", list.Key_pem, "-o", "StrictHostKeyChecking=no", publicIp, "hostname", "-i"},
		}
		pIp := common.GetPrivateIp(str1)
		ConsulMarshal(list.DC_id, list.Consul[0].Ip, publicIp, pIp)
		cIp := "ubuntu@" + publicIp + ":/home/ubuntu/fedCloud/consul/config.json"
		name :=  strconv.Itoa(list.DC_id) 
		configPath := "/home/ubuntu/DC" + name + "/consul.json"

		fmt.Println(time.Now(), configPath)
		con_arr1 := [][]string{
			[]string{"scp", "-i", list.Key_pem, "-o", "StrictHostKeyChecking=no", configPath, cIp},
		}
		common.ParseCmd(con_arr1)
		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(list.DC_id) + "/consul" + strconv.Itoa(index) + ".log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
		session.Stdin = &inbyte
                err = session.Shell()

		_, err = inbyte.WriteString("sudo rm -rf /home/ubuntu/fedCloud/consul/serf /home/ubuntu/fedCloud/consul/raft /home/ubuntu/fedCloud/cosnul/checkpoint-signature /usr/local/bin/consul\n")
		_, err = inbyte.WriteString("sudo cp /home/ubuntu/fedCloud/consul/consul /usr/local/bin\n")
		_, err = inbyte.WriteString(consul)
	}else if host == "4" {
		healthCheck(list)
    	}

	session.Wait()
	//session.Close()
}
