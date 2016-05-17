package auto_update

import (
	"fmt"
	"log"
	"os"
	"net/http"
)

func healthCheck(list DC) {
        var mCount, sCount, gCount int
        var m_cnt, s_cnt, g_cnt bool
        m_cnt = false
        s_cnt = false
        g_cnt = false
        f, _ := os.Create("/home/ubuntu/keep_alive.txt")
        log.SetOutput(f)
        for {
                response, m_err := http.Get("http://" + list.Master[0].Ip + ":5050/health")
                if m_err != nil {
                        if m_cnt == true && mCount == 0 {
                                fmt.Printf("master is down \n")
                                log.Printf("master is down\n", m_err)
                                mCount = 1
                        }
                } else if response != nil {
                        if m_cnt == false {
                                m_cnt = true
                                fmt.Println("response is \n", response)
                                fmt.Println("master is up\n")
                        }
                }
		j := 0
		for  ; j < len(list.Slave) ; j++ {
                sresponse, s_err := http.Get("http://" + list.Slave[j].Ip + ":5051/health")
                if s_err != nil {
                        if s_cnt == true && sCount == 0 {
                                fmt.Printf("slave is down \n")
                                log.Printf("slave is down\n", s_err)
                                sCount = 1
                        }
                } else if sresponse != nil {
                        if s_cnt == false {
                                s_cnt = true
                                fmt.Println("response is \n", sresponse)
                                fmt.Println("slave is up\n")
                        }
                }
		}
                gresponse, g_err := http.Get("http://" + list.Gossiper[0].Ip + ":8080/healthz")
                if g_err != nil {
                        if g_cnt == true && gCount == 0 {
                                fmt.Printf("gossiper is down\n")
                                log.Printf("gossiper is down\n", g_err)

                                gCount = 1
                        }
                } else if gresponse != nil {
                        if g_cnt == false {
                                g_cnt = true
                                fmt.Println("response is \n", gresponse)
                                fmt.Println("gossiper is up\n")
                        }
                }

        }

}
