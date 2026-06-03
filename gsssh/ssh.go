package gsssh

import (
	"fmt"
	"time"

	"github.com/w896736588/go-tool/gstool"
	"golang.org/x/crypto/ssh"
)

type SshConfig struct {
	Name                 string `json:"name"`                   // custom name
	Host                 string `json:"host"`                   // host
	Port                 string `json:"port"`                   // port
	UserName             string `json:"username"`               // username
	Password             string `json:"password"`               // password
	ConnectTimeoutSecond int    `json:"connect_timeout_second"` // connect timeout in seconds, default 5
	ExecAfterConnect     string `json:"exec_after_connect"`     // command executed after SSH connection (e.g. Jumpserver asset selection)
}

type Ssh struct {
	sshConfig *SshConfig
	client    *ssh.Client
}

// NewSsh ssh auth password
func NewSsh(sshConfig *SshConfig) *Ssh {
	return &Ssh{
		sshConfig: sshConfig,
	}
}

// ConnectAuthPassword ssh auth password
func (h *Ssh) ConnectAuthPassword() error {
	if h.client != nil {
		return nil
	}

	connectTimeoutSecond := h.sshConfig.ConnectTimeoutSecond
	if connectTimeoutSecond <= 0 {
		connectTimeoutSecond = 5
	}

	sshConfig := &ssh.ClientConfig{
		User: h.sshConfig.UserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(h.sshConfig.Password),
		},
		Timeout:         time.Duration(connectTimeoutSecond) * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production environments, do not use InsecureIgnoreHostKey
	}

	var sshConnErr error
	h.client, sshConnErr = ssh.Dial("tcp", fmt.Sprintf(`%s:%s`, h.sshConfig.Host, h.sshConfig.Port), sshConfig)
	if sshConnErr != nil {
		gstool.FmtPrintlnLogTime("ssh connect failed: %s", sshConnErr.Error())
		return fmt.Errorf("ssh conn error: %s", sshConnErr.Error())
	}
	if h.sshConfig.ExecAfterConnect != "" {
		if execErr := h.execAfterConnect(); execErr != nil {
			return execErr
		}
	}
	return nil
}

func (h *Ssh) execAfterConnect() error {
	session, sessionErr := h.client.NewSession()
	if sessionErr != nil {
		return fmt.Errorf("ssh create session failed: %s", sessionErr.Error())
	}
	defer session.Close()
	if execErr := session.Run(h.sshConfig.ExecAfterConnect); execErr != nil {
		return fmt.Errorf("ssh exec after connect failed: %s", execErr.Error())
	}
	return nil
}

// Close releases all resources.
func (h *Ssh) Close() {
	if h.client != nil {
		_ = h.client.Close()
		h.client = nil
	}
}
