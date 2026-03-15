package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "uso: compose-generator <archivo-salida> <cantidad-clientes>")
		os.Exit(1)
	}

	outputFile := os.Args[1]

	numClients, err := strconv.Atoi(os.Args[2])
	if err != nil || numClients < 1 {
		fmt.Fprintln(os.Stderr, "error: la cantidad de clientes debe ser un entero positivo")
		os.Exit(1)
	}

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	yamlPath := filepath.Join(dir, "docker-compose-base.yaml")

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error leyendo yaml: %v\n", err)
		os.Exit(1)
	}

	var compose Compose
	if err := yaml.Unmarshal(data, &compose); err != nil {
		fmt.Fprintf(os.Stderr, "error parseando yaml: %v\n", err)
		os.Exit(1)
	}

	for i := range numClients {
		client := Client{ID: i + 1}
		service := client.toService()
		compose.Services[service.ContainerName] = service
	}

	out, err := yaml.Marshal(&compose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serializando yaml: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error escribiendo archivo: %v\n", err)
		os.Exit(1)
	}
}
