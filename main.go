package main

import (
	_ "embed"
	"flag"
	"github.com/servicelayernetworking/topogen/pkg"
	"log"
	"os"
)

func main() {
	// parse command-line arguments for topology file
	filename := flag.String("filename", "topology.yaml", "path to topology file")
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
		CodeOutputDir: *codeOutputDir,
		Topo: topology,
		K8sOutfile: *kubernetesOutput,
		ExperimentName: *experimentName,
		ContainerRegistryPrefix: *containerRegistryPrefix,
		BuildAndPush: *buildAndPush,
	}
	generator.Generate()
}
