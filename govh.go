package main

/*
author: ronron@architux.com

Please get credential for access
to the ovh api :
https://eu.api.ovh.com/createToken/
https://api.ovh.com/createToken/index.cgi?GET=/sms&GET=/sms/%2a&PUT=/sms/%2a&DELETE=/sms/%2a&POST=/sms/%2a
https://docs.ovh.com/gb/en/customer/first-steps-with-ovh-api/

The ovh golang wrapper
https://github.com/ovh/go-ovh
*/

import (
	"flag"
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const (
	MISImage   string = "debian9-64bits-singlepart-RAID1"
	MDSImage   string = "debian9-plesk17_64"
	SshKeyName string = "mykey"
)

var (
	configPtr  = flag.String("c", "", "yaml configuration file")
	installPtr = flag.Bool("i", false, "Install server")
	statusPtr  = flag.Bool("s", false, "Install server status")
	verifyPtr  = flag.Bool("v", false, "Check server state")
)

type ConfigFile struct {
	Servers []struct {
		Name    string `yaml:"name"`
		Image   string `yaml:"image"`
		Options string `yaml:"options"`
	} `yaml:"servers"`
}

type DSInstallStatus struct {
	ElapsedTime int `json:"elapsedTime"`
	Progress    []ProgressStruct
}

type ProgressStruct struct {
	Error   string `yaml:"error"`
	Status  string `yaml:"status"`
	Comment string `yaml:"comment"`
}

type DS struct {
	Name       string `json:"name"`
	State      string `json:"state"`
	Datacenter string `json:"datacenter"`
}

type Install struct {
	Details      InstallDetails `json:"details"`
	TemplateName string         `json:"templateName"`
}
type InstallDetails struct {
	CustomHostname             string `json: self.customHostname`
	Language                   string `json:"language"`
	SshKeyName                 string `json:"sshKeyName"`
	UseDistribKernel           string `json:"useDistribKernel"`
	PostInstallationScriptLink string `json:"postInstallationScriptLink"`
}

func init() {
	flag.Parse()
	switch {
	case len(*configPtr) == 0:
		os.Exit(0)
	default:
	}
}

var config ConfigFile

func ParseConfigFile(inputflag string) {
	filename := inputflag
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
}

func (c ConfigFile) DSGetInstallStatus() {
	var dsis DSInstallStatus
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}
	for _, s := range c.Servers {
		fmt.Printf("-- %s --\n", s.Name)
		if err := client.Get("/dedicated/server/"+s.Name+"/install/status", &dsis); err != nil {
			fmt.Printf("Error: %q\n", err)
		}
		fmt.Printf("Temps passé: %d secondes\n", dsis.ElapsedTime)
		for _, progress := range dsis.Progress {
			fmt.Printf("--Comment: %s\nStatus :%s\nError: %s\n", progress.Comment, progress.Status, progress.Error)
		}
	}
}

func (c ConfigFile) DSGetStatus() {
	var ds DS
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}
	for _, s := range c.Servers {
		if err := client.Get("/dedicated/server/"+s.Name, &ds); err != nil {
			fmt.Printf("Error: %q\n", err)
			return
		}
		fmt.Printf("server: %s %s %s\n", ds.Name, ds.State, ds.Datacenter)
	}
}

func (c ConfigFile) DSPostInstall() {
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}
	for _, s := range c.Servers {
		switch {
		case s.Image == "mds":
			s.Image = MDSImage
		case s.Image == "mis":
			s.Image = MISImage
		default:
			s.Image = MISImage
		}
		params := &Install{
			TemplateName: s.Image,
			Details: InstallDetails{
				CustomHostname:   s.Name,
				Language:         "fr",
				SshKeyName:       SshKeyName,
				UseDistribKernel: "true",
			},
		}
		if err := client.Post("/dedicated/server/"+s.Name+"/install/start", params, nil); err != nil {
			fmt.Printf("Error: %q\n", err)
			return
		}
		fmt.Printf("Starting installation of: %s\n", s.Name)
	}
}

func main() {
	ParseConfigFile(*configPtr)
	if *verifyPtr {
		config.DSGetStatus()
	} else if *installPtr {
		config.DSPostInstall()
	} else if *statusPtr {
		config.DSGetInstallStatus()
	}
}
