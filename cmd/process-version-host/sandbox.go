package main

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

func setupSandbox() (string, string, error) {
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

func runSandbox(in, out string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", errors.Wrap(err, "could not create client")
	}

	reader, err := cli.ImagePull(ctx, DOCKER_IMAGE, types.ImagePullOptions{})
	if err != nil {
		return "", errors.Wrap(err, "could not pull image")
	}
	io.Copy(os.Stdout, reader)

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
		}, nil, nil, "")
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

	return buff.String(), nil
}
