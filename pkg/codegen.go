package pkg

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"os/exec"
)

//go:embed templates/app.tmpl
var appTmpl string

//go:embed templates/dockerfile.tmpl
var dockerfileTmpl string

//go:embed templates/gomod.tmpl
var gomodTmpl string

//go:embed templates/service.tmpl
var serviceTmpl string

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

type TopoCodeGenerator struct {
	K8sOutfile       string
	CodeOutputDir string
	ExperimentName string
	ContainerRegistryPrefix string
	Topo          Topology
	BuildAndPush bool
}


func (g *TopoCodeGenerator) Generate() error {
	// open kubernetes config file for appending
	g.K8sOutfile = fmt.Sprintf("%s/kubernetes.yaml", g.CodeOutputDir)
	k8sFile, err := os.OpenFile(g.K8sOutfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open kubernetes config file: %v", err)
	}
	// clear file
	k8sFile.Truncate(0)
	defer k8sFile.Close()
	for _, svc := range g.Topo.Services {
		if svc.GatewayNextHop != "" {
			for _, nextHop := range g.Topo.Services {
				if nextHop.Name == svc.GatewayNextHop {
					for _, nxtMethod := range nextHop.Methods {
						m := Method {
							Method: "POST",
							Path: nxtMethod.Path,
						}
						m.Calls = []Call{
							{
								Name:   nextHop.Name,
								Method: "POST",
								Path: nxtMethod.Path,
							},
						}
						svc.Methods = append(svc.Methods, m)
					}
				}
			}
		}
		imageName := g.GenerateService(svc)
		svc.Image = imageName + ":latest"
		tmpl, err := template.New("service").Parse(serviceTmpl)
		if err != nil {
			log.Fatalf("failed to parse service.tmpl: %v", err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, svc); err != nil {
			log.Fatalf("failed to execute template: %v", err)
		}
		k8sFile.Write(buf.Bytes())
		// write newline
		// fmt.Printf("Writing %d bytes to %s\n", len([]byte("\n")), g.K8sOutfile)
		k8sFile.Write([]byte("\n"))
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (g *TopoCodeGenerator) CleanseService(svc *Service) {
	// todo add custom listening port values
	for i, method := range svc.Methods {
		for j, call := range method.Calls {
			if call.Port == 0 {
				svc.Methods[i].Calls[j].Port = 8080
			}
		}
		if len(method.ConcurrentCalls) == 0 {
			if method.CallConcurrency == 0 {
				method.CallConcurrency = 1
			}
			// divide calls into method.CallConcurrency groups of len(method.Calls) / method.CallConcurrency
			concurrentCalls := [][]Call{}
			step := len(method.Calls) / method.CallConcurrency
			for i := 0; i < len(method.Calls); i += step {
				// fmt.Printf("i: %v\n", i)
				concurrentCalls = append(concurrentCalls, method.Calls[i:min(i+method.CallConcurrency, len(method.Calls))])
			}
			svc.Methods[i].ConcurrentCalls = concurrentCalls
		}
		// fmt.Printf("Concurrent calls: %v\n", svc.Methods[i].ConcurrentCalls)
	}

	// print
	

}

func (g *TopoCodeGenerator) GenerateService(svc Service) string {
	// fmt.Printf("Generating service %s\n", svc.Name)
	g.CleanseService(&svc)
	funcMap := template.FuncMap{
		"replace": replace,
	}
	tmpl, err := template.New("app").Funcs(funcMap).Parse(appTmpl)
	if err != nil {
		log.Fatalf("failed to parse app.tmpl: %v", err)
	}
	// execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, svc); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}
	// create directory if it doesn't exist
	os.MkdirAll(fmt.Sprintf("%s/%s", g.CodeOutputDir, svc.Name), 0755)
	// write to file
	fmt.Printf("Writing to %s/%s/main.go\n", g.CodeOutputDir, svc.Name)
	outLoc := fmt.Sprintf("%s/%s/main.go", g.CodeOutputDir, svc.Name)
	f, err := os.Create(outLoc)
	defer f.Close()
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	f.Write(buf.Bytes())

	// write dockerfile - currently no need to template
	fmt.Printf("Writing to %s/%s/Dockerfile\n", g.CodeOutputDir, svc.Name)
	dockerfile := []byte(dockerfileTmpl)
	outLoc = fmt.Sprintf("%s/%s/Dockerfile", g.CodeOutputDir, svc.Name)
	f, err = os.Create(outLoc)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	f.Write(dockerfile)

	// write go.mod
	fmt.Printf("Writing to %s/%s/go.mod\n", g.CodeOutputDir, svc.Name)
	tmpl, err = template.New("gomod").Parse(gomodTmpl)
	if err != nil {
		log.Fatalf("failed to parse gomod.tmpl: %v", err)
	}
	buf = bytes.Buffer{}
	if err := tmpl.Execute(&buf, svc); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}
	outLoc = fmt.Sprintf("%s/%s/go.mod", g.CodeOutputDir, svc.Name)
	f, err = os.Create(outLoc)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	f.Write(buf.Bytes())

	// write build-and-push.sh
	fmt.Printf("Writing to %s/%s/build-and-push.sh\n", g.CodeOutputDir, svc.Name)
	imageName := fmt.Sprintf("%s/%s-%s", g.ContainerRegistryPrefix, g.ExperimentName, svc.Name)
	buildAndPush := []byte(fmt.Sprintf(`
docker build -t %s .
docker push %s
	`, imageName, imageName))
	outLoc = fmt.Sprintf("%s/%s/build-and-push.sh", g.CodeOutputDir, svc.Name)
	f, err = os.Create(outLoc)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	f.Write(buildAndPush)
	// change file permissions to be executable
	os.Chmod(outLoc, 0755)

	// go to service directory
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}
	os.Chdir(fmt.Sprintf("%s/%s", g.CodeOutputDir, svc.Name))
	// run go mod tidy
	fmt.Printf("Running go mod tidy\n")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Run()
	// run build-and-push.sh
	if g.BuildAndPush {
		fmt.Printf("Building and pushing image for %s\n", svc.Name)
		cmd = exec.Command("/bin/bash", "build-and-push.sh")
		err = cmd.Run()
		if err != nil {
			log.Fatalf("failed to run build-and-push.sh: %v", err)
		}
	}
	// go back
	os.Chdir(curDir)
	fmt.Printf("finished generating for service %s\n", svc.Name)
	return imageName
}
