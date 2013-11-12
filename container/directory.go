// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package container

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	ContainerDir        = "/var/lib/juju/containers"
	RemovedContainerDir = "/var/lib/juju/removed-containers"
)

// NewContainerDirectory creates a new directory for the container name in the
// directory identified by `ContainerDir`.
func NewContainerDirectory(containerName string) (directory string, err error) {
	directory = jujuContainerDirectory(containerName)
	logger.Tracef("create directory: %s", directory)
	if err = os.MkdirAll(directory, 0755); err != nil {
		logger.Errorf("failed to create container directory: %v", err)
		return "", err
	}
	return directory, nil
}

// RemoveContainerDirectory moves the container directory from `ContainerDir`
// to `RemovedContainerDir` and makes sure that the names don't clash.
func RemoveContainerDirectory(containerName string) error {
	// Move the directory.
	logger.Tracef("create old container dir: %s", RemovedContainerDir)
	if err := os.MkdirAll(RemovedContainerDir, 0755); err != nil {
		logger.Errorf("failed to create removed container directory: %v", err)
		return err
	}
	removedDir, err := uniqueDirectory(RemovedContainerDir, containerName)
	if err != nil {
		logger.Errorf("was not able to generate a unique directory: %v", err)
		return err
	}
	if err := os.Rename(jujuContainerDirectory(containerName), removedDir); err != nil {
		logger.Errorf("failed to rename container directory: %v", err)
		return err
	}
	return nil

}

func jujuContainerDirectory(containerName string) string {
	return filepath.Join(ContainerDir, containerName)
}

// uniqueDirectory returns "path/name" if that directory doesn't exist.  If it
// does, the method starts appending .1, .2, etc until a unique name is found.
func uniqueDirectory(path, name string) (string, error) {
	dir := filepath.Join(path, name)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return dir, nil
	}
	for i := 1; ; i++ {
		dir := filepath.Join(path, fmt.Sprintf("%s.%d", name, i))
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return dir, nil
		} else if err != nil {
			return "", err
		}
	}
}
