package auto_update

import (
	"../common"
	"log"
	"os/exec"
	"time"
)

func Fedmesos() {
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
	common.ParseCmd(build_p)
	cmd_end = time.Now()
	common.GetTime(cmd_start, cmd_end)

	cmd_start = time.Now()
	create_cdir := exec.Command("sudo", "mkdir", "consul")
	create_cdir.Dir = "/home/ubuntu/fedCloud/"
	create_cerr := create_cdir.Start()
	if create_cerr != nil {
		log.Fatal(create_cerr)
	}
	log.Printf("Waiting for creating dir cmd to finish...")
	create_cerr = create_cdir.Wait()
	cmd_end = time.Now()
	common.GetTime(cmd_start, cmd_end)

	cmd_start = time.Now()
	copy_consul := exec.Command("cp","/usr/local/bin/consul", "consul/.")
	copy_consul.Dir = "/home/ubuntu/fedCloud/"
	copy_cerr := copy_consul.Start()
	if copy_cerr != nil {
		log.Fatal(copy_cerr)
	}
	log.Printf("Waiting for copying consul cmd to finish...")
	copy_cerr = copy_consul.Wait()
	cmd_end = time.Now()
    common.GetTime(cmd_start, cmd_end)

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
	common.GetTime(cmd_start, cmd_end)

	cmd_start = time.Now()
	create_dir := exec.Command("sudo", "mkdir", "build")
	create_dir.Dir = "/home/ubuntu/fedCloud/mesos"
	create_err := create_dir.Start()
	if create_err != nil {
		log.Fatal(create_err)
	}
	log.Printf("Waiting for creating dir cmd to finish...")
	create_err = create_dir.Wait()
	log.Printf("Command finished with error: %v", create_err)
	cmd_end = time.Now()
	common.GetTime(cmd_start, cmd_end)

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
	common.GetTime(cmd_start, cmd_end)

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
	common.GetTime(cmd_start, cmd_end)

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
	common.GetTime(cmd_start, cmd_end)

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
	common.GetTime(cmd_start, cmd_end)

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
	common.GetTime(cmd_start, cmd_end)

	cmd_start = time.Now()
	tarMesos := exec.Command("sudo", "tar", "-zcf", "fedCloud.tar.gz", "fedCloud")
	tarMesos.Dir = "/home/ubuntu/"
	tarMesos_err := tarMesos.Start()
	if tarMesos_err != nil {
		log.Fatal(tarMesos_err)
	}
	log.Printf("Waiting for tar file creation to finish...")
	tarMesos_err = tarMesos.Wait()
	log.Printf("Command finished with error: %v", tarMesos_err)
	cmd_end = time.Now()
	common.GetTime(cmd_start, cmd_end)

}
