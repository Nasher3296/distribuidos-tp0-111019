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

type Client struct {
	ID int
}

func (c Client) toService() Service {
	name := fmt.Sprintf("client%d", c.ID)
	return Service{
		ContainerName: name,
		Image:         "client:latest",
		Entrypoint:    "/client",
		Environment: []string{
			fmt.Sprintf("CLI_ID=%d", c.ID),
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
