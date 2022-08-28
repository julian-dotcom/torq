package dev_setup

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"log"
)

type ContainerConfig struct {
	Name       string
	Image      string
	Env        []string
	Binds      []string
	Cmd        []string
	MappedPort string
	Instance   dockercontainer.ContainerCreateCreatedBody
}

type DockerDevEnvironment struct {
	Client           *client.Client
	NetworkName      string
	NetworkingConfig network.NetworkingConfig
	Containers       map[string]*ContainerConfig
	SharedVolumeName string
}

func (de *DockerDevEnvironment) CleanupContainers(ctx context.Context) {
	for _, container := range de.Containers {
		de.FindAndRemoveContainer(ctx, container.Name)

		// Also remove volumes with the same name as the container
		de.FindAndRemoveVolume(ctx, container.Name)
	}

}

func (de *DockerDevEnvironment) AddContainer(name string, image string, binds []string, env []string, cmd []string,
	mappedPort string) *ContainerConfig {
	if de.Containers == nil {
		de.Containers = make(map[string]*ContainerConfig)
	}

	de.Containers[name] = &ContainerConfig{
		Name:       name,
		Image:      image,
		Binds:      binds,
		Env:        env,
		Cmd:        cmd,
		MappedPort: mappedPort,
		Instance:   dockercontainer.ContainerCreateCreatedBody{},
	}
	return de.Containers[name]
}

func (de *DockerDevEnvironment) InitContainer(ctx context.Context, container *ContainerConfig) (err error) {

	hostConfig := &dockercontainer.HostConfig{
		Binds: container.Binds,
	}
	openPorts := nat.PortSet{}
	if container.MappedPort != "" {
		hostConfig.PortBindings = nat.PortMap{
			nat.Port(container.MappedPort) + "/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: container.MappedPort,
				},
			},
		}
		openPorts = nat.PortSet{
			nat.Port(container.MappedPort) + "/tcp": struct{}{},
		}
	}

	r, err := de.Client.ContainerCreate(ctx, &dockercontainer.Config{
		Image:        container.Image,
		Env:          container.Env,
		Cmd:          container.Cmd,
		ExposedPorts: openPorts,
	}, hostConfig, &de.NetworkingConfig, nil, container.Name)
	if err != nil {
		// log.Fatalf("Creating %s container: %v", name, err)
		return errors.Newf("Creating %s container: %v", container.Name, err)
	}
	if err := de.Client.ContainerStart(ctx, r.ID, types.ContainerStartOptions{}); err != nil {
		//log.Fatalf("Starting %s container: %v", name, err)
		return errors.Newf("Starting %s container: %v", container.Name, err)
	}
	container.Instance = r
	return nil
}

func (de *DockerDevEnvironment) CreateNetwork(ctx context.Context) (nc network.NetworkingConfig, err error) {
	e2eNetwork, err := de.Client.NetworkCreate(ctx, de.NetworkName, types.NetworkCreate{})
	if err != nil {
		return nc, errors.Newf("Creating %s network: %v", de.NetworkName, err)
	}
	nc = network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	nc.EndpointsConfig[e2eNetwork.ID] = &network.EndpointSettings{Links: []string{"e2e-btcd:blockchain"}}
	return nc, err
}

func (de *DockerDevEnvironment) FindAndRemoveContainer(ctx context.Context, name string) {
	container, err := de.FindContainerByName(ctx, name)
	if err != nil {
		log.Fatalf("Removing %s container: %v", name, err)
	}
	if container != nil {
		log.Printf("%s container found; removing\n", name)

		if container.State == "running" {
			if err := de.Client.ContainerStop(ctx, container.ID, nil); err != nil {
				log.Fatalf("Stopping %s container: %v", name, err)
			}
		}
		if err := de.Client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
			log.Fatalf("Removing %s container: %v", name, err)
		}
	}
}

func (de *DockerDevEnvironment) FindAndRemoveVolume(ctx context.Context, name string) {
	volume, err := de.FindVolumeByName(ctx, name)
	if err != nil {
		log.Fatalf("Removing old %s volume: %v", name, err)
	}
	if volume != nil {
		log.Printf("Old %s volume found; removing\n", name)
		if err := de.Client.VolumeRemove(ctx, volume.Name, false); err != nil {
			log.Fatalf("Removing old %s volume: %v", name, err)
		}
	}
}

func (de *DockerDevEnvironment) FindVolumeByName(ctx context.Context, name string) (*types.Volume, error) {
	volumes, err := de.Client.VolumeList(ctx, filters.Args{})
	if err != nil {
		return nil, err
	}
	for _, volume := range volumes.Volumes {
		if volume.Name == name {
			return volume, nil
		}
	}
	return nil, nil
}

func (de *DockerDevEnvironment) FindAndRemoveNetwork(ctx context.Context, name string) {
	network, err := de.FindNetworkByName(ctx, name)
	if err != nil {
		log.Fatalf("Removing old %s network: %v", name, err)
	}
	if network != nil {
		log.Printf("Old %s network found; removing\n", name)
		if err := de.Client.NetworkRemove(ctx, network.ID); err != nil {
			log.Fatalf("Removing old %s network: %v", name, err)
		}
	}
}

func (de *DockerDevEnvironment) FindNetworkByName(ctx context.Context, name string) (*types.NetworkResource, error) {
	networks, err := de.Client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	for _, network := range networks {
		if network.Name == name {
			return &network, nil
		}
	}
	return nil, nil
}

func (de *DockerDevEnvironment) FindContainerByName(ctx context.Context, name string) (*types.Container, error) {
	containers, err := de.Client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		for _, containerName := range container.Names {
			// internal docker names have leading slashes; trim off
			if containerName[1:] == name {
				return &container, nil
			}
		}
	}
	return nil, nil
}
