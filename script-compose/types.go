package main

import "fmt"

type Service struct {
	ContainerName string   `yaml:"container_name"`
	Image         string   `yaml:"image"`
	Entrypoint    string   `yaml:"entrypoint"`
	Environment   []string `yaml:"environment"`
	Networks      []string `yaml:"networks"`
	DependsOn     []string `yaml:"depends_on,omitempty"`
	Volumes       []string `yaml:"volumes,omitempty"`
}

type ClientBet struct {
	Nombre     string `yaml:"nombre"`
	Apellido   string `yaml:"apellido"`
	Documento  string `yaml:"documento"`
	Nacimiento string `yaml:"nacimiento"`
	Numero     string `yaml:"numero"`
}

type Client struct {
	ID  int
	Bet ClientBet
}

func (c Client) toService() Service {
	name := fmt.Sprintf("client%d", c.ID)
	return Service{
		ContainerName: name,
		Image:         "client:latest",
		Entrypoint:    "/client",
		Environment: []string{
			fmt.Sprintf("CLI_ID=%d", c.ID),
			fmt.Sprintf("NOMBRE=%s", c.Bet.Nombre),
			fmt.Sprintf("APELLIDO=%s", c.Bet.Apellido),
			fmt.Sprintf("DOCUMENTO=%s", c.Bet.Documento),
			fmt.Sprintf("NACIMIENTO=%s", c.Bet.Nacimiento),
			fmt.Sprintf("NUMERO=%s", c.Bet.Numero),
		},
		Networks:  []string{"testing_net"},
		DependsOn: []string{"server"},
		Volumes:   []string{"./client/config.yaml:/config.yaml"},
	}
}

type Compose struct {
	Name     string             `yaml:"name"`
	Services map[string]Service `yaml:"services"`
	Networks any                `yaml:"networks"`
}
