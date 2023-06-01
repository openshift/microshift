package system

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type boot struct {
	Index      int    `json:"index"`
	BootID     BootID `json:"boot_id"`
	FirstEntry int64  `json:"first_entry"`
	LastEntry  int64  `json:"last_entry"`
}

type boots []boot

func (bs boots) getBootByIndex(index int) boot {
	for _, b := range bs {
		if b.Index == index {
			return b
		}
	}
	return boot{}
}

func getBoots() (boots, error) {
	cmd := exec.Command("journalctl", "--list-boots", "--output=json", "--reverse")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("getting list of boots failed: %w", err)
	}

	boots := boots{}
	err = json.Unmarshal(output, &boots)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling list of boots failed: %w", err)
	}

	return boots, nil
}
