package main

import (
	"crypto/rand"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/servicelayernetworking/topogen/pkg"
)


func WriteToFile(filename string, fileSize int) {
	file, err := os.Create(filename)
	if err != nil {
		file, err = os.Open(filename)
		if err != nil {
			fmt.Printf("failed to open file: %v\n", err)
			return
		}
	}
	data := make([]byte, fileSize)
	rand.Read(data)
	if _, err := file.Write(data); err != nil {
		fmt.Printf("failed to write to file: %v\n", err)
	}
}

func RunCPULoad(millicoreCount int, timeMillis int) {
	if timeMillis == 0 || millicoreCount == 0 {
		return
	}

	// 500 millicore -> 500 microseconds
	runFor := time.Duration(millicoreCount) * time.Microsecond
	sleepFor := time.Duration(1000-millicoreCount) * time.Microsecond

	// runtime.LockOSThread()
	// make timer
	timer := time.NewTimer(time.Duration(timeMillis) * time.Millisecond)
	d := time.Duration(timeMillis) * time.Millisecond

	fmt.Printf("Timer duration %s, sleepFor %s, runFor %s\n", d.String(), sleepFor.String(), runFor.String())
	fmt.Printf("starting load for %dms, current time %d\n", timeMillis, time.Now().UnixMilli())
	for {
		select {
		case <-timer.C:
			fmt.Printf("finished load at for %dms, current time %d\n", timeMillis, time.Now().UnixMilli())
			runtime.UnlockOSThread()
			return
		default:
			begin := time.Now()
			for {
				if time.Since(begin) > runFor {
					break
				}
			}
			time.Sleep(sleepFor)
		}
	}
}

func main() {
	// parse command-line arguments for topology file
	filename := flag.String("topology", "topology.yaml", "path to topology file")
	codeOutputDir := flag.String("codeout", "./generated-topology", "path to output generated code")
	experimentName := flag.String("experiment", "experiment", "name of the experiment")
	containerRegistryPrefix := flag.String("registry", "ghcr.io/adiprerepa", "container registry prefix (for example ghcr.io/adiprerepa or gangmuk)")
	kubernetesOutput := flag.String("out", "./generated-topology/kubernetes.yaml", "path to kubernetes output file")
	buildAndPush := flag.Bool("build", true, "build and push the docker images")
	flag.Parse()
	// if the code output directory does not exist, create it
	if _, err := os.Stat(*codeOutputDir); os.IsNotExist(err) {
		os.MkdirAll(*codeOutputDir, 0755)
	}
	topology, err := pkg.ParseTopology(*filename)
	if err != nil {
		log.Fatalf("failed to parse topology file: %v", err)
	}
	generator := &pkg.TopoCodeGenerator{
		CodeOutputDir:           *codeOutputDir,
		Topo:                    topology,
		K8sOutfile:              *kubernetesOutput,
		ExperimentName:          *experimentName,
		ContainerRegistryPrefix: *containerRegistryPrefix,
		BuildAndPush:            *buildAndPush,
	}
	generator.Generate()
	fmt.Printf("Generated code in %s\n", *codeOutputDir)
}
