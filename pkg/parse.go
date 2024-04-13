package pkg

import (
	// "fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type Topology struct {
	Services []Service `yaml:"services"`
}

type Service struct {
	Name    string   `yaml:"name"`
	Methods []Method `yaml:"methods"`
	GatewayNextHop string `yaml:"gatewayNextHop"`
	Image string
}

type Method struct {
	Method     string `yaml:"method"`
	Path       string `yaml:"path"`
	ComputeAmount   int    `yaml:"computeAmount"`
	ComputeDuration int    `yaml:"computeDuration"`
	WriteFileSize  int    `yaml:"writeFileSize"`
	Calls      []Call `yaml:"calls"`
	ReturnSize int    `yaml:"returnSize"`
}

type Call struct {
	Name   string `yaml:"name"`
	Port  int    `yaml:"port"`
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Size   int    `yaml:"size"`
	PropogateQueryParams string `yaml:"propogateQueryParams"`
}


func ParseTopology(filename string) (Topology, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Topology{}, err
	}
	defer file.Close()
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return Topology{}, err
	}

	var topology Topology
	err = yaml.Unmarshal(contents, &topology)
	if err != nil {
		return Topology{}, err
	}
	return topology, nil
}
