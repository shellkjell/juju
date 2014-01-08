// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package manual

import (
	"fmt"
	"os"

	gc "launchpad.net/gocheck"

	envtesting "launchpad.net/juju-core/environs/testing"
	"launchpad.net/juju-core/instance"
	"launchpad.net/juju-core/juju/testing"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/state/api/params"
	jc "launchpad.net/juju-core/testing/checkers"
	"launchpad.net/juju-core/version"
)

type provisionerSuite struct {
	testing.JujuConnSuite
}

var _ = gc.Suite(&provisionerSuite{})

func (s *provisionerSuite) getArgs(c *gc.C) ProvisionMachineArgs {
	hostname, err := os.Hostname()
	c.Assert(err, gc.IsNil)
	return ProvisionMachineArgs{
		Host:    hostname,
		EnvName: "dummyenv",
	}
}

func (s *provisionerSuite) TestProvisionMachine(c *gc.C) {
	const series = "precise"
	const arch = "amd64"

	args := s.getArgs(c)
	hostname := args.Host
	args.Host = "ubuntu@" + args.Host

	envtesting.RemoveTools(c, s.Conn.Environ.Storage())
	defer fakeSSH{
		Series:             series,
		Arch:               arch,
		InitUbuntuUser:     true,
		SkipProvisionAgent: true,
	}.install(c).Restore()
	// Attempt to provision a machine with no tools available, expect it to fail.
	machineId, err := ProvisionMachine(args)
	c.Assert(err, jc.Satisfies, params.IsCodeNotFound)
	c.Assert(machineId, gc.Equals, "")

	cfg := s.Conn.Environ.Config()
	number, ok := cfg.AgentVersion()
	c.Assert(ok, jc.IsTrue)
	binVersion := version.Binary{number, series, arch}
	envtesting.AssertUploadFakeToolsVersions(c, s.Conn.Environ.Storage(), binVersion)

	for i, errorCode := range []int{255, 0} {
		c.Logf("test %d: code %d", i, errorCode)
		defer fakeSSH{
			Series:                 series,
			Arch:                   arch,
			InitUbuntuUser:         true,
			ProvisionAgentExitCode: errorCode,
		}.install(c).Restore()
		machineId, err = ProvisionMachine(args)
		if errorCode != 0 {
			c.Assert(err, gc.ErrorMatches, fmt.Sprintf("exit status %d", errorCode))
			c.Assert(machineId, gc.Equals, "")
		} else {
			c.Assert(err, gc.IsNil)
			c.Assert(machineId, gc.Not(gc.Equals), "")
			// machine ID will be incremented. Even though we failed and the
			// machine is removed, the ID is not reused.
			c.Assert(machineId, gc.Equals, fmt.Sprint(i+1))
			m, err := s.State.Machine(machineId)
			c.Assert(err, gc.IsNil)
			instanceId, err := m.InstanceId()
			c.Assert(err, gc.IsNil)
			c.Assert(instanceId, gc.Equals, instance.Id("manual:"+hostname))
		}
	}

	// Attempting to provision a machine twice should fail. We effect
	// this by checking for existing juju upstart configurations.
	defer fakeSSH{
		Provisioned:        true,
		InitUbuntuUser:     true,
		SkipDetection:      true,
		SkipProvisionAgent: true,
	}.install(c).Restore()
	_, err = ProvisionMachine(args)
	c.Assert(err, gc.Equals, ErrProvisioned)
	defer fakeSSH{
		Provisioned:              true,
		CheckProvisionedExitCode: 255,
		InitUbuntuUser:           true,
		SkipDetection:            true,
		SkipProvisionAgent:       true,
	}.install(c).Restore()
	_, err = ProvisionMachine(args)
	c.Assert(err, gc.ErrorMatches, "error checking if provisioned: exit status 255")
}

func (s *provisionerSuite) TestCreateMachineConfig(c *gc.C) {
	const series = "precise"
	const arch = "amd64"
	defer fakeSSH{
		Series:         series,
		Arch:           arch,
		InitUbuntuUser: true,
	}.install(c).Restore()
	machineId, err := ProvisionMachine(s.getArgs(c))
	c.Assert(err, gc.IsNil)

	// Now check what we would've configured it with.
	client := s.APIConn.State.Client()
	mcfg, err := createMachineConfig(client, machineId, series, arch, state.BootstrapNonce, "/var/lib/juju")
	c.Assert(err, gc.IsNil)
	c.Assert(mcfg, gc.NotNil)
	c.Assert(mcfg.APIInfo, gc.NotNil)
	c.Assert(mcfg.StateInfo, gc.NotNil)

	stateInfo, apiInfo, err := s.APIConn.Environ.StateInfo()
	c.Assert(err, gc.IsNil)
	c.Assert(mcfg.APIInfo.Addrs, gc.DeepEquals, apiInfo.Addrs)
	c.Assert(mcfg.StateInfo.Addrs, gc.DeepEquals, stateInfo.Addrs)
}
