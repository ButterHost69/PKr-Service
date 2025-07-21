package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/joho/godotenv"
)

var PKrPath = os.Getenv("LOCALAPPDATA") + "\\PKr\\"

const ServiceLogger = "Logs\\PKR-Service.log"
const EnvFilePath = "Config\\.env"

const RepoOwner = "ButterHost69"
const BaseRepoName = "PKr-Base"
const CliRepoName = "PKr-Cli"

// Update .env File
func setEnvValue(key, value, filepath string, service_logger *log.Logger) error {
	data_bytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	consts := strings.Split(string(data_bytes), "\n")
	if len(consts) == 0 {
		service_logger.Println(".env is Empty Inserting Key:", key)
		consts[0] = key + "=" + value
	}

	updated := false
	for i, c := range consts {
		if strings.HasPrefix(c, key+"=") {
			consts[i] = key + "=" + value
			updated = true
		}
	}

	if !updated {
		service_logger.Println("When Updating .env for key:", key, " was not found")
		service_logger.Println("Inserting Key and Value for the First Time: ", key)
		consts = append(consts, key+"="+value)
	}

	return os.WriteFile(filepath, []byte(strings.Join(consts, "\n")), 0644)
}

func main() {
	// Make sure log directory exists
	_ = os.MkdirAll(PKrPath, 0755)

	// Open log file
	f, err := os.OpenFile(PKrPath+ServiceLogger, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		_ = beeep.Notify("PKr-Service", "Failed to Start PKr-Service:\n"+err.Error(), "")
		return
	}
	defer f.Close()

	service_logger := log.New(f, "", log.Ldate|log.Ltime|log.Lshortfile)
	service_logger.Println("Logger Started ...")
	service_logger.Println("PKR Path:", PKrPath)

	// Checking is .env Present
	dotenv_f, err := os.OpenFile(PKrPath+EnvFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		service_logger.Println("Error: Could not Open from the .env file")
		service_logger.Println("Error:", err.Error())
		service_logger.Println("Path: ", PKrPath)
		err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
		if err != nil {
			service_logger.Println("Error while displaying Push Notification:", err)
		}
		return
	}
	defer dotenv_f.Close()

	// Open Latest File
	err = godotenv.Load(PKrPath + EnvFilePath)
	if err != nil {
		service_logger.Println("Error: Could not Load .env file using godotenv")
		service_logger.Println("Error:", err.Error())
		service_logger.Println("Path: ", PKrPath)
		err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
		if err != nil {
			service_logger.Println("Error while displaying Push Notification:", err)
		}
		return
	}

	service_logger.Println("Loaded Constants From .env File")

	// Check and Fetch Base
	// '-' doesnt work in godotenv.load
	curr_base_ver := os.Getenv("PKr_Base_Version")
	base_latest_tag, err := getLatestTag(RepoOwner, BaseRepoName)
	if err != nil {
		service_logger.Println("Error: Could Not Fetch Latest Tag for PKr-Base")
		service_logger.Println("Error: [Base]", err)
		err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
		if err != nil {
			service_logger.Println("Error while displaying Push Notification:", err)
		}
		return
	}
	if curr_base_ver != base_latest_tag {
		service_logger.Printf("PKr_Base_Version is different from Latest [Curr: %v - Latest: %v]\n", curr_base_ver, base_latest_tag)
		service_logger.Println("Fetching Latest Base Version")

		service_logger.Println("Latest Base Version - ", base_latest_tag)
		service_logger.Println("Downloading Latest Base Vesion", base_latest_tag)

		err = downloadExeFromTag(RepoOwner, BaseRepoName, base_latest_tag, PKrPath+BaseRepoName+".exe")
		if err != nil {
			service_logger.Println("Error: Downloading Latest Version of Base: ", base_latest_tag)
			service_logger.Println("Error: ", err)

			err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
			if err != nil {
				service_logger.Println("Error while displaying Push Notification:", err)
			}
			return
		}

		// Update .env
		// '-' doesnt work in godotenv.load
		err = setEnvValue("PKr_Base_Version", base_latest_tag, PKrPath+EnvFilePath, service_logger)
		if err != nil {
			service_logger.Println("Error: Updating .env file to reflect latest base tag: ", base_latest_tag)
			service_logger.Println("Error: ", err)

			err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
			if err != nil {
				service_logger.Println("Error while displaying Push Notification:", err)
			}
			return
		}

		service_logger.Println("PKr-Base Updated to Version: " + base_latest_tag)
		err = beeep.Notify("PKr-Service", "PKr-Base Updated to: "+base_latest_tag, "")
		if err != nil {
			service_logger.Println("Error while displaying Push Notification:", err)
		}
	}

	// Check and Fetch Cli
	// '-' doesnt work in godotenv.load
	curr_cli_ver := os.Getenv("PKr_Cli_Version")
	cli_latest_tag, err := getLatestTag(RepoOwner, CliRepoName)
	if err != nil {
		service_logger.Println("Error: Could Not Fetch Latest Tag for PKr-Cli")
		service_logger.Println("Error: [Cli]", err)
		err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
		if err != nil {
			service_logger.Println("Error while displaying Push Notification:", err)
		}
		return
	}
	if curr_cli_ver != cli_latest_tag {
		service_logger.Printf("PKr_Cli_Version is different from Latest [Curr: %v - Latest: %v]\n", curr_cli_ver, cli_latest_tag)
		service_logger.Println("Fetching Latest Cli Version")

		service_logger.Println("Latest Cli Version - ", cli_latest_tag)
		service_logger.Println("Downloading Latest Cli Vesion", cli_latest_tag)

		err = downloadExeFromTag(RepoOwner, CliRepoName, cli_latest_tag, PKrPath+CliRepoName+".exe")
		if err != nil {
			service_logger.Println("Error: Downloading Latest Version of Cli: ", cli_latest_tag)
			service_logger.Println("Error: ", err)

			err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
			if err != nil {
				service_logger.Println("Error while displaying Push Notification:", err)
			}
			return
		}

		// Update .env
		// '-' doesnt work in godotenv.load
		err = setEnvValue("PKr_Cli_Version", cli_latest_tag, PKrPath+EnvFilePath, service_logger)
		if err != nil {
			service_logger.Println("Error: Updating .env file to reflect latest cli	 tag: ", cli_latest_tag)
			service_logger.Println("Error: ", err)

			err = beeep.Notify("PKr-Service", "Failed to Start PKr-Service - Check Logs", "")
			if err != nil {
				service_logger.Println("Error while displaying Push Notification:", err)
			}
			return
		}

		service_logger.Println("PKr-Cli Updated to Version: " + cli_latest_tag)
		err = beeep.Notify("PKr-Service", "PKr-Cli Updated to: "+cli_latest_tag, "")
		if err != nil {
			service_logger.Println("Error while displaying Push Notification:", err)
		}

	}

	// Configure and Start Base
	cmd := exec.Command(PKrPath + "PKr-Base.exe")

	// Optional: Set output to the current terminal
	cmd.Stdout = service_logger.Writer()
	cmd.Stderr = service_logger.Writer()

	err = cmd.Run()
	if err != nil {
		service_logger.Println("Error: In Starting PKr-Base")
		service_logger.Println("Error: ", err)

		err = beeep.Notify("PKr-Service", "Failed to Start PKr-Base - Check Logs", "")
		if err != nil {
			service_logger.Println("Error while displaying Push Notification:", err)
		}
		return
	}

	service_logger.Println("Latest PKr-Base Running...")
}
