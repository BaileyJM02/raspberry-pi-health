package main

import (
	"log"
	"os"
	"time"

	jexia "github.com/baileyjm02/jexia-sdk-go"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/shirou/gopsutil/host"
)

type system struct {
	Hostname  string            `json:"hostname"`
	Platform  string            `json:"platform"`
	CPU       linuxproc.CPUStat `json:"cpu"`
	Processes uint64            `json:"processes"`
	Memory    linuxproc.MemInfo `json:"memory"`
	Uptime    linuxproc.Uptime  `json:"uptime"`
	Load      linuxproc.LoadAvg `json:"load"`
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func collectInfo() *system {
	hostStat, _ := host.Info()
	memory, _ := linuxproc.ReadMemInfo("/proc/meminfo")
	stat, _ := linuxproc.ReadStat("/proc/stat")
	uptime, _ := linuxproc.ReadUptime("/proc/uptime")
	load, _ := linuxproc.ReadLoadAvg("/proc/loadavg")

	info := new(system)

	info.Hostname = hostStat.Hostname
	info.Platform = hostStat.Platform
	info.CPU = stat.CPUStatAll
	info.Processes = stat.Processes
	info.Memory = *memory
	info.Uptime = *uptime
	info.Load = *load

	return info
}

func setupJexia() *jexia.Client {
	client := jexia.NewClient(
		os.Getenv("PROJECT_ID"),
		os.Getenv("PROJECT_ZONE"),
	)
	client.UseAPKToken(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	client.AutoRefreshToken()
	return client
}

func getTimer() *time.Timer {
	return time.NewTimer(30 * time.Second)
}

func main() {
	client := setupJexia()
	healthDataset := client.GetDataset("health")
	done := make(chan bool)
	go func() {
		// start a timer counting down from the token lifetime
		timer := getTimer()
		for {
			select {
			// triggered when the abortRefresh channel is closed
			case <-timer.C:
				info := collectInfo()
				_, err := healthDataset.Insert([]interface{}{info})
				if err != nil {
					log.Println(err)
				}
				timer = getTimer()
			}
		}
	}()
	<-done // Block forever
}
