package bot

import (
	"encoding/json"
	"log"
	"os/exec"
)

func CheckBattery() {
	cmd := exec.Command("termux-battery-status")

	output, err := cmd.Output()
	if err != nil {
		log.Println("Error checking battery status:", err)
		return
	}

	var battery struct {
		Percentage int `json:"percentage"`
	}

	if err := json.Unmarshal(output, &battery); err != nil {
		log.Println("Error Unmarshalling battery status:", err)
		return
	}
	log.Printf("Battery: %d%%", battery.Percentage)

	if battery.Percentage < 100 {
		DgSession.ChannelMessageSend("1301895231230443530", "Mr.@a0xd I need to charge please help :(")
	}
}
