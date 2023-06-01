package system

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrBootNotFound = errors.New("boot not found")
)

type BootID string
type DeploymentID string

type Boot struct {
	ID       BootID    `json:"id"`
	BootTime time.Time `json:"boot_time"`
}

type SystemInfo interface {
	IsOSTree() (bool, error)
	GetCurrentDeploymentID() (DeploymentID, error)

	GetCurrentBoot() (*Boot, error)
	GetPreviousBoot() (*Boot, error)
}

var _ SystemInfo = (*systemInfo)(nil)

type systemInfo struct{}

func NewSystemInfo() *systemInfo {
	return &systemInfo{}
}

func (s *systemInfo) IsOSTree() (bool, error) {
	return isOSTree()
}

func (s *systemInfo) GetCurrentDeploymentID() (DeploymentID, error) {
	return getCurrentDeploymentID()
}

func (s *systemInfo) GetCurrentBoot() (*Boot, error) {
	return s.getBootByIndex(0)
}

func (s *systemInfo) GetPreviousBoot() (*Boot, error) {
	return s.getBootByIndex(-1)
}

func (s *systemInfo) getBootByIndex(index int) (*Boot, error) {
	bs, err := getBoots()
	if err != nil {
		return nil, fmt.Errorf("getting boots failed: %w", err)
	}
	boot := bs.getBootByIndex(index)
	if boot.BootID == "" {
		return nil, ErrBootNotFound
	}
	return &Boot{
		ID:       boot.BootID,
		BootTime: time.UnixMicro(boot.FirstEntry),
	}, nil
}
