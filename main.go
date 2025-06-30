// my_service.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

const serviceName = "PKr-Service"
// const BaseExePath = "C:\\Program Files\\PKr\\"
// const CliExePath = "C:\\Program Files\\PKr\\"
const PKrPath = "C:\\Program Files\\PKr\\"

var cPKrPath = filepath.Join("C:","Program Files", "PKr")
const ServiceLogger = "service.log"


const RepoOwner = "ButterHost69"
const BaseRepoName = "PKr-Base"
const CliRepoName = "PKr-Cli"


type myService struct{}

func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const acceptCmds = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}

	// Make sure log directory exists
	_ = os.MkdirAll(PKrPath, 0755)

	// Open log file
	f, err := os.OpenFile(PKrPath + ServiceLogger, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s <- svc.Status{State: svc.Stopped}
		return false, 1
	}
	defer f.Close()

	service_logger := log.New(f, "", log.Ldate|log.Ltime|log.Lshortfile)
	service_logger.Println("Logger Started ...")
	service_logger.Println("cPKR Path:", cPKrPath)

	// Checking is .env Present
	dotenv_f, err := os.OpenFile(PKrPath+".env", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
	    service_logger.Println("Error: Could not Open from the .env file")
		service_logger.Println("Error:", err.Error())
		service_logger.Println("Path: ", PKrPath)
		s <- svc.Status{State: svc.Stopped}
		return false, 1
	}
	defer dotenv_f.Close()


	// Open Latest File
	err = godotenv.Load(PKrPath + ".env")
    if err != nil {
		service_logger.Println("Error: Could not Load .env file using godotenv")
		service_logger.Println("Error:", err.Error())
		service_logger.Println("Path: ", PKrPath)
		s <- svc.Status{State: svc.Stopped}
		return false, 1
    }

	service_logger.Println("Loaded Constants From .env File")

	// Check and Fetch Base
	curr_base_ver := os.Getenv("PKr-Base-Version")
	base_latest_tag, err := getLatestTag(RepoOwner, BaseRepoName)
	if err != nil {
		service_logger.Println("Error: Could Not Fetch Latest Tag for PKr-Base")
		service_logger.Println("Error: [Base]", err)
		s <- svc.Status{State: svc.Stopped}
		return false, 1
	}
	if curr_base_ver != base_latest_tag {
		service_logger.Printf("PKr-Base-Version is different from Latest [Curr: %v - Latest: %v]\n", curr_base_ver, base_latest_tag)
		service_logger.Println("Fetching Latest Base Version")
		
		service_logger.Println("Latest Base Version - ", base_latest_tag)
		service_logger.Println("Downloading Latest Base Vesion", base_latest_tag)

		err = downloadExeFromTag(RepoOwner, BaseRepoName, base_latest_tag, PKrPath)
		if err != nil {
			service_logger.Println("Error: Downloading Latest Version of Base: ", base_latest_tag)
			service_logger.Println("Error: ", err)

			s <- svc.Status{State: svc.Stopped}
			return false, 1
		}
	}

	// Check and Fetch Cli
	curr_cli_ver := os.Getenv("PKr-Cli-Version")
	cli_latest_tag, err := getLatestTag(RepoOwner, CliRepoName)
	if err != nil {
		service_logger.Println("Error: Could Not Fetch Latest Tag for PKr-Cli")
		service_logger.Println("Error: [Cli]", err)
		s <- svc.Status{State: svc.Stopped}
		return false, 1
	}
	if curr_cli_ver != cli_latest_tag{
		service_logger.Printf("PKr-Cli-Version is different from Latest [Curr: %v - Latest: %v]\n", curr_cli_ver, cli_latest_tag)
		service_logger.Println("Fetching Latest Cli Version")
		
		service_logger.Println("Latest Cli Version - ", cli_latest_tag)
		service_logger.Println("Downloading Latest Cli Vesion", cli_latest_tag)

		err = downloadExeFromTag(RepoOwner, CliRepoName, cli_latest_tag, PKrPath)
		if err != nil {
			service_logger.Println("Error: Downloading Latest Version of Cli: ", cli_latest_tag)
			service_logger.Println("Error: ", err)

			s <- svc.Status{State: svc.Stopped}
			return false, 1
		}
	}
	
	// Start Base
	cmd := exec.Command(PKrPath + "PKr-Cli.exe")

	// Optional: Set output to the current terminal
	cmd.Stdout = service_logger.Writer()
	cmd.Stderr = service_logger.Writer()

	err = cmd.Run()
	if err != nil {
		service_logger.Println("Error: In Starting PKr-Base")
		service_logger.Println("Error: ", err)
		s <- svc.Status{State: svc.Stopped}
		return false, 1
	}

	service_logger.Println("Latest PKr-Base Running...")

	// Start service loop
	s <- svc.Status{State: svc.Running, Accepts: acceptCmds}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			msg := fmt.Sprintf("Hello: %s\n", time.Now().Format(time.RFC1123))
			_, _ = f.WriteString(msg)
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Interrogate:
				s <- c.CurrentStatus
			}
		}
	}

	s <- svc.Status{State: svc.StopPending}
	return false, 0
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		err = debug.Run(name, &myService{})
	} else {
		err = svc.Run(name, &myService{})
	}
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}

func main() {
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	runService(serviceName, !isService)
}
