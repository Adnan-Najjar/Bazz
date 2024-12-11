package bot

import (
	"encoding/json"
	"log"
	"os/exec"
	"fmt"
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

	userID := "1267009801423163484"
	if battery.Percentage < 10 {
		if _,err := DgSession.UserChannelCreate(userID); err != nil {
			log.Println("Error creating Direct Message Channel:", err)
			return
		}
		if _,err := DgSession.ChannelMessageSend(userID, fmt.Sprintf("Battery is low %d%%", battery.Percentage)); err != nil {
			log.Println("Error sending Direct Message:", err)
			return
		}
	}
}
