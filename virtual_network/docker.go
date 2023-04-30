package virtual_network

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
)

type ContainerConfig struct {
	Name       string
	Image      string
	Env        []string
	Binds      []string
	Cmd        []string
	MappedPort string
	Id         string
}

type DockerDevEnvironment struct {
	Client            *client.Client
	NetworkName       string
	NetworkingConfig  network.NetworkingConfig
	Containers        map[string]*ContainerConfig
	SharedVolumeName  string
	DockerHubUsername string
	DockerHubPassword string
}

func (de *DockerDevEnvironment) CleanupContainers(ctx context.Context) error {
	for _, container := range de.Containers {
		err := de.FindAndRemoveContainer(ctx, container.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (de *DockerDevEnvironment) CleanupDefaultVolumes(ctx context.Context) error {
	for _, container := range de.Containers {
		err := de.FindAndRemoveVolume(ctx, container.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (de *DockerDevEnvironment) AddContainer(name string, image string, binds []string, env []string, cmd []string,
	mappedPort string, id string) *ContainerConfig {
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
		Id:         id,
	}
	return de.Containers[name]
}

func (de *DockerDevEnvironment) pullImage(ctx context.Context, imageName string) (err error) {
	// pull image if we don't have it already
	// locally built images should be built before calling this function
	if de.DockerHubUsername != "" && de.DockerHubPassword != "" {
		images, err := de.Client.ImageList(ctx, types.ImageListOptions{})
		if err != nil {
			return errors.Wrap(err, "Listing available docker images")
		}

		imageAlreadyExists := false
		for _, image := range images {
			for _, tag := range image.RepoTags {
				if strings.Contains(tag, imageName) {
					imageAlreadyExists = true
				}
			}
		}

		// pull image using docker hub authentication to prevent us hitting download limits
		if !imageAlreadyExists {
			authConfig := types.AuthConfig{
				Username: de.DockerHubUsername,
				Password: de.DockerHubPassword,
			}
			encodedJSON, err := json.Marshal(authConfig)
			if err != nil {
				return errors.Wrap(err, "JSON marshall of auth config")
			}
			authStr := base64.URLEncoding.EncodeToString(encodedJSON)

			_, err = de.Client.ImagePull(ctx, imageName, types.ImagePullOptions{RegistryAuth: authStr})
			if err != nil {
				return errors.Wrap(err, "Docker pulling image")
			}
		}
	}
	return nil
}

func (de *DockerDevEnvironment) CreateContainer(ctx context.Context, container *ContainerConfig) (err error) {

	if err = de.pullImage(ctx, container.Image); err != nil {
		return errors.Wrap(err, "Pulling image")
	}

	hostConfig := &dockercontainer.HostConfig{
		Binds: container.Binds,
	}
	// Map ports
	openPorts := nat.PortSet{}
	if container.MappedPort != "" {
		// Split MappedPort port into host and container port
		hostPort := strings.Split(container.MappedPort, ":")[0]
		containerPort := strings.Split(container.MappedPort, ":")[1]

		hostConfig.PortBindings = nat.PortMap{
			nat.Port(containerPort) + "/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			},
		}
		openPorts = nat.PortSet{
			nat.Port(containerPort) + "/tcp": struct{}{},
		}
	}

	r, err := de.Client.ContainerCreate(ctx, &dockercontainer.Config{
		Hostname:     container.Name,
		Image:        container.Image,
		Env:          container.Env,
		Cmd:          container.Cmd,
		ExposedPorts: openPorts,
	}, hostConfig, &de.NetworkingConfig, nil, container.Name)
	if err != nil {
		return errors.Wrapf(err, "Creating %s container", container.Name)
	}
	container.Id = r.ID
	return nil
}

func (de *DockerDevEnvironment) InitContainer(ctx context.Context, container *ContainerConfig) error {
	err := de.CreateContainer(ctx, container)
	if err != nil {
		return err
	}
	if err := de.Client.ContainerStart(ctx, container.Id, types.ContainerStartOptions{}); err != nil {
		return errors.Wrapf(err, "can't start %s container", container.Name)
	}
	return nil
}

func (de *DockerDevEnvironment) StartContainer(ctx context.Context, cc *ContainerConfig) (err error) {
	container, err := de.FindContainerByName(ctx, cc.Name)
	if err != nil {
		return errors.Wrapf(err, "Can't find container with the name %s", cc.Name)
	}
	if err := de.Client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrapf(err, "can't start %s container", cc.Name)
	}
	cc.Id = container.ID
	return nil
}

func (de *DockerDevEnvironment) StopContainer(ctx context.Context, name string) (err error) {
	container, err := de.FindContainerByName(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "can't find container with the name %s", name)
	}
	timeout := 30
	if err := de.Client.ContainerStop(ctx, container.ID, dockercontainer.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		return errors.Wrapf(err, "can't stop %s container", name)
	}
	return nil
}

func (de *DockerDevEnvironment) CreateNetwork(ctx context.Context) (nc network.NetworkingConfig, err error) {
	nw, err := de.Client.NetworkCreate(ctx, de.NetworkName, types.NetworkCreate{})
	if err != nil {
		return network.NetworkingConfig{}, errors.Wrapf(err, "Creating %s network", de.NetworkName)
	}
	nc = network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	nc.EndpointsConfig[nw.ID] = &network.EndpointSettings{Links: []string{de.NetworkName + "-btcd:blockchain"}}
	return nc, nil
}

func (de *DockerDevEnvironment) FindAndRemoveContainer(ctx context.Context, name string) error {
	container, err := de.FindContainerByName(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "Removing %s container", name)
	}
	if container != nil {
		log.Printf("%s container found; removing\n", name)

		if container.State == "running" {
			if err := de.Client.ContainerStop(ctx, container.ID, dockercontainer.StopOptions{}); err != nil {
				return errors.Wrapf(err, "Stopping %s container", name)
			}
		}
		if err := de.Client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
			return errors.Wrapf(err, "Removing %s container", name)
		}
	}
	return nil
}

func (de *DockerDevEnvironment) FindAndRemoveVolume(ctx context.Context, name string) error {
	volume, err := de.FindVolumeByName(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "Removing old %s volume", name)
	}
	if volume != nil {
		log.Printf("Old %s volume found; removing\n", name)
		if err := de.Client.VolumeRemove(ctx, volume.Name, false); err != nil {
			return errors.Wrapf(err, "Removing old %s volume", name)
		}
	}
	return nil
}

func (de *DockerDevEnvironment) FindVolumeByName(ctx context.Context, name string) (*volume.Volume, error) {
	volumes, err := de.Client.VolumeList(ctx, filters.Args{})
	if err != nil {
		return nil, errors.Wrap(err, "Docker volume list")
	}
	for _, volume := range volumes.Volumes {
		if volume.Name == name {
			return volume, nil
		}
	}
	return nil, nil
}

func (de *DockerDevEnvironment) FindAndRemoveNetwork(ctx context.Context, name string) error {
	network, err := de.FindNetworkByName(ctx, name)
	if err != nil {
		return errors.Wrapf(err, "Removing old %s network", name)
	}
	if network != nil {
		log.Printf("Old %s network found; removing\n", name)
		if err := de.Client.NetworkRemove(ctx, network.ID); err != nil {
			return errors.Wrapf(err, "Removing old %s network", name)
		}
	}
	return nil
}

func (de *DockerDevEnvironment) FindNetworkByName(ctx context.Context, name string) (*types.NetworkResource, error) {
	networks, err := de.Client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Docker network list")
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
		return nil, errors.Wrap(err, "Docker container list")
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

func (de *DockerDevEnvironment) BuildImage(ctx context.Context, path string, name string) error {
	tar, err := archive.TarWithOptions(path, &archive.TarOptions{ExcludePatterns: []string{"web/node_modules", ".git"}})
	if err != nil {
		return errors.Wrapf(err, "Creating %s archive", name)
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{name},
		Remove:     true,
	}

	if de.DockerHubUsername != "" && de.DockerHubPassword != "" {
		log.Debug().Msgf("Using authentication with docker")
		opts.AuthConfigs = map[string]types.AuthConfig{"https://index.docker.io/v1/": {Username: de.DockerHubUsername, Password: de.DockerHubPassword}}
	}

	res, err := de.Client.ImageBuild(ctx, tar, opts)
	if err != nil {
		return errors.Wrapf(err, "Building %s docker image", name)
	}
	defer res.Body.Close()
	err = printBuildOutput(res.Body)
	if err != nil {
		return errors.Wrapf(err, "Printing build output for %s docker image", name)
	}
	return nil
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func printBuildOutput(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		if len(os.Getenv("DEBUG")) > 0 {
			fmt.Println(scanner.Text())
		}
	}

	errLine := &ErrorLine{}
	if err := json.Unmarshal([]byte(lastLine), errLine); err != nil {
		return errors.Wrap(err, "JSON unmarshal of build output")
	}
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "Buffio scanner err")
	}

	return nil
}

func ExecJSONReturningCommand(ctx context.Context, cli *client.Client, containerId string,
	cmd []string, returnObject interface{}) error {

	bufStdout, bufStderr, err := ExecCommand(ctx, cli, containerId, cmd)
	if err != nil {
		if len(os.Getenv("DEBUG")) > 0 {
			log.Printf("%s\n", string(err.Error()))
		}
		return errors.Wrap(err, "Exec command on container")
	}
	if len(bufStderr.Bytes()) > 0 {
		log.Info().Msg("std error not empty")
		return errors.New("Stderr not empty")
	}

	err = json.Unmarshal(bufStdout.Bytes(), returnObject)
	if err != nil {
		return errors.Wrap(err, "json unmarshal")
	}
	return nil
}

func ExecCommand(ctx context.Context, cli *client.Client,
	containerId string,
	cmd []string) (bufStdout bytes.Buffer, bufStderr bytes.Buffer, err error) {

	c := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: cmd}
	execID, err := cli.ContainerExecCreate(ctx, containerId, c)
	if err != nil {
		return bufStdout, bufStderr, errors.Wrap(err, "Container exec create")
	}

	res, err := cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return bufStdout, bufStderr, errors.Wrap(err, "Container exec start")
	}

	err = cli.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return bufStdout, bufStderr, errors.Wrap(err, "Container exec start")
	}

	// stdcopy.StdCopy(os.Stdout, os.Stderr, res.Reader)
	if _, err = stdcopy.StdCopy(&bufStdout, &bufStderr, res.Reader); err != nil {
		return bufStdout, bufStderr, errors.Wrap(err, "Copying data to std out and std error")
	}
	// DEBUG Tip: uncomment below to see raw output of commands
	if len(os.Getenv("DEBUG")) > 0 {
		log.Printf("%s\n", string(bufStdout.String()))
		log.Printf("%s\n", string(bufStderr.String()))
	}
	return bufStdout, bufStderr, nil
}

type noRetryError struct{}

func (nre noRetryError) Error() string {
	return "Skip retries"
}

func Retry(operation func() error, delayMilliseconds int, maxWaitMilliseconds int) error {
	totalWaited := 0
	for {
		if totalWaited > maxWaitMilliseconds {
			return errors.New("Exceeded maximum wait period")
		}
		err := operation()
		var noRetry noRetryError
		if errors.As(err, &noRetry) {
			return err
		}
		if err == nil {
			break
		}
		log.Info().Msg("Checking...")
		time.Sleep(time.Duration(delayMilliseconds) * time.Millisecond)
		totalWaited += delayMilliseconds
	}
	return nil
}

func WriteConnectionDetails(ctx context.Context, cli *client.Client, name string) error {
	// Copy bobs macaroon and tls file to local directory
	tlsFileReader, _, err := cli.CopyFromContainer(ctx, name, "/root/.lnd/tls.cert")
	if err != nil {
		return errors.Wrap(err, "Copying tls file")
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := tlsFileReader.Close(); err != nil {
			panic(err)
		}
	}()

	tlsTar := tar.NewReader(tlsFileReader)
	// hdr gives you the header of the tar file
	_, err = tlsTar.Next()
	if err == io.EOF || err != nil {
		// EOF == end of tar archive
		return errors.Wrap(err, "Reading tls tar header")
	}
	tlsBuf := new(bytes.Buffer)
	_, err = tlsBuf.ReadFrom(tlsTar)
	if err != nil {
		return errors.Wrap(err, "Reading tls tar")
	}
	// write the whole body at once
	err = os.WriteFile(fmt.Sprintf("virtual_network/generated_files/%v-tls.cert", name), tlsBuf.Bytes(), 0600)
	if err != nil {
		panic(err)
	}

	macaroonFileReader, _, err := cli.CopyFromContainer(ctx, name,
		"/root/.lnd/data/chain/bitcoin/simnet/admin.macaroon")
	if err != nil {
		return errors.Wrapf(err, "Copying tls file")
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := macaroonFileReader.Close(); err != nil {
			panic(err)
		}
	}()

	// file comes out as a tar, untar it
	macaroonTar := tar.NewReader(macaroonFileReader)
	// hdr gives you the header of the tar file
	_, err = macaroonTar.Next()
	if err == io.EOF || err != nil {
		// EOF == end of tar archive
		return errors.Wrap(err, "Reading macaroon tar header")
	}
	macaroonBuf := new(bytes.Buffer)
	_, err = macaroonBuf.ReadFrom(macaroonTar)
	if err != nil {
		return errors.Wrap(err, "Reading macaroon tar")
	}

	// write the whole body at once
	err = os.WriteFile(fmt.Sprintf("virtual_network/generated_files/%v-admin.macaroon", name), macaroonBuf.Bytes(), 0600)
	if err != nil {
		panic(err)
	}

	return nil
}
