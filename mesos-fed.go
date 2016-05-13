package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os/exec"
	//	    "io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
	//	"net/http"
	"strconv"
	//"path/filepath"
)

type GossiperConfig struct {
	Name           string //Name of this gossiper
	City           string
	Country        string
	MasterEndPoint string //MesosMaster's IP address
	ConfigType     string //What type of gossiper is this ? Join a federation or start a federation? Values = JOIN or NEW
	JoinEndPoint   string //If we are joining an already runnig federatoin the what is the EndPoint?
	LogFile        string //Name of the logfile
	HTTPPort       string //Defaults to 8080 if otherwise specify explicitly
	TCPPort        string //TCP port at which gossiper will bind and listen for anon module to connect to
	GPort          int    //Port at which gossper should start
	AdvertiseAddr  string //The advertised address of the gossiper so that other gossipers coudl connect
	ConsConfig     CoConfig
}

//global consul config
type CoConfig struct {
	IsLeader    bool
	DCEndpoint  string
	StorePreFix string
	DCName      string
}

type ConsulConfig struct {
	Name               string
	Data_dir           string
	Log_level          string
	Server             bool
	Bootstrap          bool
	Start_join         string
	Bind_addr          string
	Ad                 Addresses
	Retry_join_wan     []string
	Retry_interval_wan string
	AdvAdd             Advertise_addrs
}

type Addresses struct {
	Http string
}

type Advertise_addrs struct {
	Serf_lan string
	Serf_wan string
	Rpc      string
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
	var conf *ConsulConfig
	if name == "1" {
		conf = &ConsulConfig{Name: name, Data_dir: "/home/ubuntu/fedCloud/consul", Log_level: "INFO", Server: true, Bootstrap: true, Start_join: privateIp, Bind_addr: privateIp, Ad: Addresses{Http: privateIp}, Retry_join_wan:[]string{""}, Retry_interval_wan: "", AdvAdd: Advertise_addrs{Serf_lan: publicIp + ":8301", Serf_wan: publicIp + ":8302", Rpc: publicIp + ":8303"}}
	} else {
		conf = &ConsulConfig{Name: name, Data_dir: "/home/ubuntu/fedCloud/consul", Log_level: "INFO", Server: true, Bootstrap: true, Start_join: privateIp, Bind_addr: privateIp, Ad: Addresses{Http: privateIp}, Retry_join_wan:[]string{fConsulIp}, Retry_interval_wan: "5s", AdvAdd: Advertise_addrs{Serf_lan: publicIp + ":8301", Serf_wan: publicIp + ":8302", Rpc: publicIp + ":8303"}}
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
		conf = &GossiperConfig{Name: name, ConfigType: "New", GPort: 4400, MasterEndPoint: mPublicIp, AdvertiseAddr: publicIp, Country: country, City: city, ConsConfig: CoConfig{IsLeader: true, DCEndpoint: consulIp + ":8500", StorePreFix: "Federa"}}
	} else {

		conf = &GossiperConfig{Name: name, ConfigType: "Old", JoinEndPoint: fGossiperIp, GPort: 4400, MasterEndPoint: mPublicIp, AdvertiseAddr: publicIp, Country: country, City: city, ConsConfig: CoConfig{IsLeader: false, DCEndpoint: consulIp + ":8500", StorePreFix: "Federa"}}

	}
	list, _ := json.MarshalIndent(conf, " ", "  ")

	err := ioutil.WriteFile(path, list, 0644)
	if err != nil {
		fmt.Printf("WriteFile json Error: %tv", err)
	}
}

func (conf *Config) Json_unmarshal_conf(file []byte) {
	var privateIp string
	err := json.Unmarshal(file, &conf)
	if err != nil {
		log.Fatal(err)
	}
	wg := new(sync.WaitGroup)
	f, _ := os.Create("/home/ubuntu/test/log.txt")
	log.SetOutput(f)
	defer f.Close()
	fConsulIp := conf.List[0].Consul[0].Ip
	fGossiperIp := conf.List[0].Gossiper[0].Ip
	for i := 0; i < len(conf.List); i++ {
		wg.Add(1)
		path_dir := "/home/ubuntu/fed/" + "DC" + strconv.Itoa(conf.List[i].DC_id)
		fmt.Println(path_dir)
		err := os.MkdirAll(path_dir, 0777)
		if err != nil {
			fmt.Println("error in creating directory \n", err)
		}
		str1 := [][]string{
			[]string{"ssh", "-i", conf.List[i].Key_pem, "-o", "StrictHostKeyChecking=no", conf.List[i].Master[0].Ip, "hostname", "-i"},
		}
		privateIp = get_ip(str1)
		fmt.Println("I am inside", privateIp, "\n")
		fmt.Println("log file is creating\n")
		log.Printf("=============Data Center Id : %d ================\n", conf.List[i].DC_id)
		go create_hosts(conf.List[i].Master[0].Ip, wg, "0", privateIp, conf.List[i].Key_pem, conf.List[i].DC_id, conf.List[i].Country, conf.List[i].City, fGossiperIp, conf.List[i].Consul[0].Ip)
		go create_hosts(conf.List[i].Slave[0].Ip, wg, "1", conf.List[i].Master[0].Ip, conf.List[i].Key_pem, conf.List[i].DC_id, conf.List[i].Country, conf.List[i].City, fGossiperIp, conf.List[i].Consul[0].Ip)
		go create_hosts(conf.List[i].Gossiper[0].Ip, wg, "2", conf.List[i].Master[0].Ip, conf.List[i].Key_pem, conf.List[i].DC_id, conf.List[i].Country, conf.List[i].City, fGossiperIp, conf.List[i].Consul[0].Ip)
		go create_hosts(conf.List[i].Consul[0].Ip, wg, "3", fConsulIp, conf.List[i].Key_pem, conf.List[i].DC_id, conf.List[i].Country, conf.List[i].City, fGossiperIp, conf.List[i].Consul[0].Ip)
		//		go keep_alive(conf.List[i].Master[0].Ip, conf.List[i].Slave[0].Ip, conf.List[i].Gossiper[0].Ip,wg, conf.List[i].Country, conf.List[i].City, fGossiperIp, conf.List[i].Consul[0].Ip)
	}
	wg.Wait()
}
func get_ip(str [][]string) string {
	for _, val := range str {
		cmd := exec.Command(val[0], val[1:]...)
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

func create_hosts(publicIp string, wg *sync.WaitGroup, host string, privateIp string, path_pem string, id int, country string, city string, fGossiperIp string, consulIp string) {
	var time_start, time_end time.Time
	Ip := "ubuntu@" + publicIp + ":/home/ubuntu"
	fmt.Println(Ip)
	Ip1 := "ubuntu@" + publicIp
	con_arr := [][]string{
		[]string{"ssh", "-i", path_pem, "-o", "StrictHostKeyChecking=no", Ip1, "sudo rm -rf /home/ubuntu/fedCloud", ";", "rm /home/ubuntu/fedCloud.tar.gz"},
		[]string{"scp", "-i", path_pem, "-o", "StrictHostKeyChecking=no", "/home/ubuntu/fedCloud.tar.gz", Ip},
	}
	log.Println("create_hosts called")
	fmt.Println("Waiting for cleaning the VMs and copying the new Mesos_Federation Version...")
	time_start = time.Now()
	process_arr(con_arr)
	time_end = time.Now()
	time_diff(time_start, time_end)

	if host == "2" {
		GossiperMarshal(id, privateIp, publicIp, fGossiperIp, consulIp, city, country)
	} else if host == "3" {
		str1 := [][]string{
			[]string{"ssh", "-i", path_pem, "-o", "StrictHostKeyChecking=no", publicIp, "hostname", "-i"},
		}
		pIp := get_ip(str1)
		ConsulMarshal(id, privateIp, publicIp, pIp)
	}
	ssh_con(publicIp, host, privateIp, path_pem, id)

	wg.Done()
}

/*func keep_alive(masterIp string, slaveIp string, gossIp string,wg *sync.WaitGroup) {
        var mCount, sCount, gCount int
        var m_cnt, s_cnt, g_cnt bool
        m_cnt = false
        s_cnt = false
        g_cnt = false
        f, _ := os.Create("/home/ubuntu/keep_alive.txt")
        log.SetOutput(f)
        for {
                response, m_err := http.Get("http://" + masterIp + ":5050/health")
                if m_err != nil {
                        if m_cnt == true && mCount == 0 {
                                fmt.Printf("master is down \n")
                                log.Printf("master is down\n", m_err)
                                mCount = 1
                        }
                } else if response != nil {
                        if m_cnt == false {
                                m_cnt = true
                                fmt.Println("response is %s\n", response)
                        }
                }
                sresponse, s_err := http.Get("http://" + slaveIp + ":5051/health")
                if s_err != nil {
                        if s_cnt == true && sCount == 0 {
                                fmt.Printf("slave is down \n")
                                log.Printf("slave is down\n", s_err)
                                sCount = 1
                        }
                } else if sresponse != nil {
                        if s_cnt == false {
                                s_cnt = true
                                fmt.Println("response is %s\n", sresponse)
                        }
                }

                gresponse, g_err := http.Get("http://" + gossIp + ":8080/healthz")
                if g_err != nil {
                        if g_cnt == true && gCount == 0 {
                                fmt.Printf("gossiper is down\n")
                                log.Printf("gossiper is down\n", g_err)

                                gCount = 1
                        }
                } else if gresponse != nil {
                        if g_cnt == false {
                                g_cnt = true
                                fmt.Println("response is %s\n", gresponse)
                        }
                }

        }

          wg.Done()
}*/

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
func ssh_con(publicIp string, host string, privateIp string, path_pem string, id int) {
	var time_start, time_end time.Time
	var b bytes.Buffer
	var inbyte bytes.Buffer
	var path string
	master := "sudo ./bin/mesos-master.sh --ip=" + privateIp + " --work_dir=$HOME --advertise_ip=" + publicIp + " --advertise_port=5050 --modules='file://../../FedModules/FedModules.json' --allocator='mesos_fed_allocator_module' \n"
	slave := "sudo ./bin/mesos-slave.sh --master=" + privateIp + ":5050 \n"
	gossiper := "./gossiper -config=config.json \n"
	//	consul := "sudo ./consul agent -config-file=/home/ubuntu/fedCloud/consul/config.json > /tmp/server.log 2>&1 \n"
	consul := "sudo ./consul agent -config-file=/home/ubuntu/fedCloud/consul/config.json \n"
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

	client, err := ssh.Dial("tcp", publicIp+":22", config)
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
	time_diff(time_start, time_end)
	session, err := client.NewSession()
	//defer session.Close()
	if host == "0" {
		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(id) + "/master.log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
	} else if host == "1" {
		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(id) + "/slave.log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
	} else if host == "2" {

		gIp := "ubuntu@" + publicIp + ":/home/ubuntu/fedCloud/gossiper/config.json"
		name := strconv.Itoa(id)
		configPath := "/home/ubuntu/DC" + name + "/gossiper.json"

		con_arr1 := [][]string{
			[]string{"scp", "-i", path_pem, "-o", "StrictHostKeyChecking=no", configPath, gIp},
		}
		process_arr(con_arr1)

		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(id) + "/gossiper.log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
	} else if host == "3" {

		cIp := "ubuntu@" + publicIp + ":/home/ubuntu/fedCloud/consul/config.json"
		name := strconv.Itoa(id)
		configPath := "/home/ubuntu/DC" + name + "/consul.json"

		fmt.Println(time.Now(), configPath)
		con_arr1 := [][]string{
			[]string{"scp", "-i", path_pem, "-o", "StrictHostKeyChecking=no", configPath, cIp},
		}
		process_arr(con_arr1)
		path = "/home/ubuntu/fed/" + "DC" + strconv.Itoa(id) + "/consul.log"
		logfile, _ := os.Create(path)
		session.Stdout = logfile
		session.Stderr = logfile
	}

	session.Stdin = &inbyte
	err = session.Shell()
	if host == "0" {
		_, err = inbyte.WriteString("cd /home/ubuntu/fedCloud/mesos/build\n")
		_, err = inbyte.WriteString(master)
	} else if host == "1" {
		_, err = inbyte.WriteString("cd /home/ubuntu/fedCloud/mesos/build\n")
		_, err = inbyte.WriteString(slave)
	} else if host == "2" {
		_, err = inbyte.WriteString("cd /home/ubuntu/fedCloud/gossiper\n")
		_, err = inbyte.WriteString(gossiper)
	} else if host == "3" {
		_, err = inbyte.WriteString("sudo rm -rf /home/ubuntu/serf /home/ubuntu/raft /home/ubuntu/checkpoint-signature /tmp/server.log /usr/local/bin/consul\n")
		_, err = inbyte.WriteString("sudo cp /home/ubuntu/fedCloud/consul/consul /usr/local/bin\n")
		//		_, err = inbyte.WriteString("sudo cd  /usr/local/bin\n")
		_, err = inbyte.WriteString("cp /home/ubuntu/config_consul.json /home/ubuntu/fedCloud/consul/config.json\n")
		_, err = inbyte.WriteString(consul)
	}
	session.Wait()
	//log.Printf(b.String())
	//session.Close()

}

func process_arr(str [][]string) {
	for _, val := range str {
		cmd := exec.Command(val[0], val[1:]...)
		//                log.Printf("I am inside process_arr" )
		OP, err := cmd.Output()
		log.Println("process_arr OP frm the cmd is ", val[0], val[1:])
		log.Println("process_arr Output frm the cmd is ", string(OP))
		if err != nil {
			log.Printf(err.Error())
		} else {
			log.Printf("success", string(OP))
		}
	}

}
func fed_mesos() {
	var cmd_start, cmd_end time.Time
	build_p := [][]string{
		[]string{"sudo", "rm", "-rf", "/home/ubuntu/fedCloud"},
		[]string{"sudo", "rm", "-rf", "/home/ubuntu/fedCloud.tar.gz"},
		[]string{"mkdir", "/home/ubuntu/fedCloud"},
		[]string{"git", "clone", "https://github.com/huawei-cloudfederation/mesos.git", "/home/ubuntu/fedCloud/mesos"},
		[]string{"git", "clone", "https://github.com/huawei-cloudfederation/FedModules.git", "/home/ubuntu/fedCloud/FedModules"},
		[]string{"git", "clone", "https://github.com/huawei-cloudfederation/gossiper.git", "/home/ubuntu/fedCloud/gossiper"},
	}
	cmd_start = time.Now()
	log.Println("Cleaning previous version and downloading mesos, FedModules, Gossiper and Consul...")
	process_arr(build_p)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	create_cdir := exec.Command("mkdir", "consul")
	create_cdir.Dir = "/home/ubuntu/fedCloud/"
	create_cerr := create_cdir.Start()
	if create_cerr != nil {
		log.Fatal(create_cerr)
	}
	log.Printf("Waiting for creating dir cmd to finish...")
	create_cerr = create_cdir.Wait()
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	copy_consul := exec.Command("sudo", "cp", "/usr/local/bin/consul", "consul/.")
	copy_consul.Dir = "/home/ubuntu/fedCloud/"
	copy_cerr := copy_consul.Start()
	if copy_cerr != nil {
		log.Fatal(copy_cerr)
	}
	log.Printf("Waiting for copying consul cmd to finish...")
	copy_cerr = copy_consul.Wait()
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	boot := exec.Command("sudo", "./bootstrap")
	boot.Dir = "/home/ubuntu/fedCloud/mesos"
	start_err := boot.Start()
	if start_err != nil {
		log.Fatal(start_err)
	}
	log.Printf("Waiting for bootstrap cmd to finish...")
	start_err = boot.Wait()
	log.Printf("Command finished with error: %v", start_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	create_dir := exec.Command("mkdir", "build")
	create_dir.Dir = "/home/ubuntu/fedCloud/mesos"
	create_err := create_dir.Start()
	if create_err != nil {
		log.Fatal(create_err)
	}
	log.Printf("Waiting for creating dir cmd to finish...")
	create_err = create_dir.Wait()
	log.Printf("Command finished with error: %v", create_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	conf := exec.Command("sudo", "../configure")
	conf.Dir = "/home/ubuntu/fedCloud/mesos/build"
	conf_err := conf.Start()
	if conf_err != nil {
		log.Fatal(conf_err)
	}
	log.Printf("Waiting for configure cmd to finish...")
	conf_err = conf.Wait()
	log.Printf("Command finished with error: %v", conf_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	mcmd := exec.Command("sudo", "make", "-j", "2")
	mcmd.Dir = "/home/ubuntu/fedCloud/mesos/build"
	mcmd_err := mcmd.Start()
	if mcmd_err != nil {
		log.Fatal(mcmd_err)
	}
	log.Printf("Waiting for make cmd to finish...")
	mcmd_err = mcmd.Wait()
	log.Printf("Command finished with error: %v", mcmd_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	/* mcheck := exec.Command("sudo","make","check")
	   mcheck.Dir = "/home/ubuntu/fedCloud/mesos/build"
	   mcheck_err := mcheck.Start()
	   if mcheck_err != nil {
	       log.Fatal(mcheck_err)
	   }
	   log.Printf("Waiting for cmd make check to finish...")
	   mcheck_err = mcheck.Wait()
	   log.Printf("Command finished with error: %v",mcheck_err)
	*/

	cmd_start = time.Now()
	bfedmod := exec.Command("sudo", "./build.sh", "/home/ubuntu/fedCloud/mesos")
	bfedmod.Dir = "/home/ubuntu/fedCloud/FedModules"
	bfedmod_err := bfedmod.Start()
	if bfedmod_err != nil {
		log.Fatal(bfedmod_err)
	}
	log.Printf("Waiting for building FedModules to finish...")
	bfedmod_err = bfedmod.Wait()
	log.Printf("Command finished with error: %v", bfedmod_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	goGet := exec.Command("go", "get", "-d", "./...")
	goGet.Dir = "/home/ubuntu/fedCloud/gossiper"
	goGet_err := goGet.Start()
	if goGet_err != nil {
		log.Fatal(goGet_err)
	}
	log.Printf("Waiting for  go get command to finish...")
	goGet_err = goGet.Wait()
	log.Printf("Command finished with error: %v", goGet_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	goBuild := exec.Command("go", "build")
	goBuild.Dir = "/home/ubuntu/fedCloud/gossiper"
	goBuild_err := goBuild.Start()
	if goBuild_err != nil {
		log.Fatal(goBuild_err)
	}
	log.Printf("Waiting for Gossiper build to finish...")
	goBuild_err = goBuild.Wait()
	log.Printf("Command finished with error: %v", goBuild_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

	cmd_start = time.Now()
	tarMesos := exec.Command("sudo", "tar", "-zcvf", "fedCloud.tar.gz", "fedCloud")
	tarMesos.Dir = "/home/ubuntu/"
	tarMesos_err := tarMesos.Start()
	if tarMesos_err != nil {
		log.Fatal(tarMesos_err)
	}
	log.Printf("Waiting for tar file creation to finish...")
	tarMesos_err = tarMesos.Wait()
	log.Printf("Command finished with error: %v", tarMesos_err)
	cmd_end = time.Now()
	time_diff(cmd_start, cmd_end)

}

func time_diff(cmd_start time.Time, cmd_end time.Time) {
	cmd_start.Format("Mon, Jan 2, 2006 at 3:04pm")
	cmd_end.Format("Mon, Jan 2, 2006 at 3:04pm")
	diff := cmd_end.Sub(cmd_start)
	log.Println("time taken to run the command : ", diff)
}

func main() {
	var ver int
	file, err := ioutil.ReadFile("./config2.json")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("currently installed mesos git dir : https://github.com/apache/mesos.git")
		fmt.Println("press any key, if you want to use current mesos version,\n press 1, if you want to update the mesos with latest version")
		fmt.Scanf("%d", &ver)
		if ver == 1 {
			fed_mesos()
		}
		config := Config{}
		config.Json_unmarshal_conf(file)
	}
}
