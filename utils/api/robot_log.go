package api

import "time"

type RobotLog struct {
	Id              int
	Level           string
	WindowsIdentity string
	ProcessName     string
	TimeStamp       time.Time
	Message         string
	RobotName       string
	HostMachineName string
	MachineId       int
	MachineKey      string
	RuntimeType     string
}

func NewRobotLog(
	id int,
	level string,
	windowsIdentity string,
	processName string,
	timeStamp time.Time,
	message string,
	robotName string,
	hostMachineName string,
	machineId int,
	machineKey string,
	runtimeType string,
) *RobotLog {
	return &RobotLog{
		id,
		level,
		windowsIdentity,
		processName,
		timeStamp,
		message,
		robotName,
		hostMachineName,
		machineId,
		machineKey,
		runtimeType,
	}
}
