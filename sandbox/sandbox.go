package sandbox

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
)

var (
	DOCKER_IMAGE = os.Getenv("DOCKER_IMAGE")
)

func Setup() (string, string, error) {
	tmpDir := os.TempDir()
	inDir, err := ioutil.TempDir(tmpDir, "in")
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create in directory")
	}
	outDir, err := ioutil.TempDir(tmpDir, "out")
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create out directory")
	}

	return inDir, outDir, nil
}

func Init(ctx context.Context) error {
	if DOCKER_IMAGE == "" {
		return errors.New("DOCKER_IMAGE needs to be present")
	}

	if false {
		cli, err := getCli()
		if err != nil {
			return errors.Wrap(err, "could not create client")
		}

		reader, err := cli.ImagePull(ctx, DOCKER_IMAGE, types.ImagePullOptions{})
		if err != nil {
			return errors.Wrap(err, "could not pull image")
		}
		if _, err := io.Copy(os.Stdout, reader); err != nil {
			return errors.Wrap(err, "failed to display pull logs")
		}
	}
	return nil
}

func getCli() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func Run(ctx context.Context, containerName, in, out string) (string, error) {
	cli, err := getCli()
	if err != nil {
		return "", errors.Wrap(err, "could not create client")
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: DOCKER_IMAGE,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   in,
					Target:   "/input",
					ReadOnly: true,
				},
				{
					Type:   mount.TypeBind,
					Source: out,
					Target: "/output",
				},
			},
		}, nil, nil, containerName)
	if err != nil {
		return "", errors.Wrap(err, "could not create container")
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", errors.Wrap(err, "could not start container")
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", errors.Wrap(err, "failed to wait for container")
		}
	case <-statusCh:
	}

	opts := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}
	logsReader, err := cli.ContainerLogs(ctx, resp.ID, opts)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve logs")
	}

	buff := new(bytes.Buffer)

	_, err = stdcopy.StdCopy(buff, buff, logsReader)
	if err != nil {
		return "", errors.Wrap(err, "could not display logs")
	}

	// once we are done remove the container to free the name in case we rerun it
	removeopts := types.ContainerRemoveOptions{}
	if err := cli.ContainerRemove(ctx, resp.ID, removeopts); err != nil {
		return "", errors.Wrapf(err, "could not remove container %s / %s", resp.ID, containerName)
	}

	return buff.String(), nil
}
