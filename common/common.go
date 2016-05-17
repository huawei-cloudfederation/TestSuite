package common

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)


func ParseCmd(str [][]string) {
	for _, val := range str {
		cmd := exec.Command(val[0], val[1:]...)
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

func GetTime(cmd_start time.Time, cmd_end time.Time) {
	cmd_start.Format("Mon, Jan 2, 2006 at 3:04pm")
	cmd_end.Format("Mon, Jan 2, 2006 at 3:04pm")
	diff := cmd_end.Sub(cmd_start)
	log.Println("time taken to run the command : ", diff)
}

func GetPrivateIp(str [][]string) string {
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

