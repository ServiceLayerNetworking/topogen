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
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Size   int    `yaml:"size"`
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
