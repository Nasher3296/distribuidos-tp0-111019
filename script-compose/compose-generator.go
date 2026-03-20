package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

const serverTemplate = `  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL={{LOG_LEVEL}}
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
`

const clientTemplate = `  {{NAME}}:
    container_name: {{NAME}}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={{ID}}
      - CLI_LOG_LEVEL={{LOG_LEVEL}}
    networks:
      - testing_net
    volumes:
      - ./client/config.yaml:/config.yaml
    depends_on:
      - server
`

const networkTemplate = `networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
`

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

	serverCfg, err := ini.Load(filepath.Join(dir, "..", "server", "config.ini"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error leyendo config del server: %v\n", err)
		os.Exit(1)
	}
	serverLogLevel := serverCfg.Section("DEFAULT").Key("LOGGING_LEVEL").String()

	clientCfgData, err := os.ReadFile(filepath.Join(dir, "..", "client", "config.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error leyendo config del client: %v\n", err)
		os.Exit(1)
	}
	var clientCfg struct {
		Log struct {
			Level string `yaml:"level"`
		} `yaml:"log"`
	}
	if err := yaml.Unmarshal(clientCfgData, &clientCfg); err != nil {
		fmt.Fprintf(os.Stderr, "error parseando config del client: %v\n", err)
		os.Exit(1)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "name: tp0\n\nservices:\n")

	sb.WriteString(applyTemplate(serverTemplate, map[string]string{
		"{{LOG_LEVEL}}": serverLogLevel,
	}))

	for i := range numClients {
		name := fmt.Sprintf("client%d", i+1)
		sb.WriteString(applyTemplate(clientTemplate, map[string]string{
			"{{NAME}}":      name,
			"{{ID}}":        strconv.Itoa(i + 1),
			"{{LOG_LEVEL}}": clientCfg.Log.Level,
		}))
	}

	sb.WriteString("\n")
	sb.WriteString(networkTemplate)

	if err := os.WriteFile(outputFile, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error escribiendo archivo: %v\n", err)
		os.Exit(1)
	}
}

func applyTemplate(tmpl string, replacements map[string]string) string {
	for k, v := range replacements {
		tmpl = strings.ReplaceAll(tmpl, k, v)
	}
	return tmpl
}
