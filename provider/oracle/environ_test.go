// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package oracle_test

import (
	"errors"
	"fmt"

	"github.com/juju/juju/environs"
	envtesting "github.com/juju/juju/environs/testing"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/provider/oracle"
	"github.com/juju/juju/testing"
	"github.com/juju/juju/tools"
	"github.com/juju/utils/arch"
	"github.com/juju/version"
	gc "gopkg.in/check.v1"
)

type environSuite struct{}

var _ = gc.Suite(&environSuite{})

func (e *environSuite) TestNewOracleEnviron(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)
}

func (e *environSuite) TestAvailabilityZone(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	zones, err := environ.AvailabilityZones()
	c.Assert(err, gc.IsNil)
	c.Assert(zones, gc.NotNil)
}

func (e *environSuite) TestInstanceAvailabilityZoneNames(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	zones, err := environ.InstanceAvailabilityZoneNames([]instance.Id{
		instance.Id("0"),
	})
	c.Assert(err, gc.IsNil)
	c.Assert(zones, gc.NotNil)
}

func (e *environSuite) TestInstanceAvailabilityZoneNamesWithErrors(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		FakeEnvironAPI{
			FakeInstancer: FakeInstancer{
				InstanceErr: errors.New("FakeInstanceErr"),
			},
		},
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	_, err = environ.InstanceAvailabilityZoneNames([]instance.Id{instance.Id("0")})
	c.Assert(err, gc.NotNil)

	environ, err = oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		FakeEnvironAPI{
			FakeInstance: FakeInstance{
				AllErr: errors.New("FakeInstanceErr"),
			},
		},
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	_, err = environ.InstanceAvailabilityZoneNames([]instance.Id{
		instance.Id("0"),
		instance.Id("1"),
	})
	c.Assert(err, gc.NotNil)
}

func (e *environSuite) TestPrepareForBootstrap(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	ctx := envtesting.BootstrapContext(c)
	err = environ.PrepareForBootstrap(ctx)
	c.Assert(err, gc.IsNil)
}

func (e *environSuite) TestPrepareForBootstrapWithErrors(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		FakeEnvironAPI{
			FakeAuthenticater: FakeAuthenticater{
				AuthenticateErr: errors.New("FakeAuthenticateErr"),
			},
		},
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	ctx := envtesting.BootstrapContext(c)
	err = environ.PrepareForBootstrap(ctx)
	c.Assert(err, gc.NotNil)
}

func makeToolsList(series string) tools.List {
	var toolsVersion version.Binary
	toolsVersion.Number = version.MustParse("1.26.0")
	toolsVersion.Arch = arch.AMD64
	toolsVersion.Series = series
	return tools.List{{
		Version: toolsVersion,
		URL:     fmt.Sprintf("http://example.com/tools/juju-%s.tgz", toolsVersion),
		SHA256:  "1234567890abcdef",
		Size:    1024,
	}}
}

//TODO
//
// func (e *environSuite) TestBootstrap(c *gc.C) {
// 	environ, err := oracle.NewOracleEnviron(
// 		oracle.DefaultProvider,
// 		environs.OpenParams{
// 			Config: testing.ModelConfig(c),
// 		},
// 		DefaultEnvironAPI,
// 	)
// 	c.Assert(err, gc.IsNil)
// 	c.Assert(environ, gc.NotNil)
//
// 	ctx := envtesting.BootstrapContext(c)
// 	_, err = environ.Bootstrap(ctx,
// 		environs.BootstrapParams{
// 			ControllerConfig:     testing.FakeControllerConfig(),
// 			AvailableTools:       makeToolsList("xenial"),
// 			BootstrapSeries:      "xenial",
// 			BootstrapConstraints: constraints.MustParse("mem=3.5G"),
// 		})
// 	c.Assert(err, gc.IsNil)
// }

func (e *environSuite) TestCreate(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	err = environ.Create(environs.CreateParams{
		ControllerUUID: "dsauhdiuashd",
	})
	c.Assert(err, gc.IsNil)
}

func (e *environSuite) TestCreateWithErrors(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		FakeEnvironAPI{
			FakeAuthenticater: FakeAuthenticater{
				AuthenticateErr: errors.New("FakeAuthenticateErr"),
			},
		},
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	err = environ.Create(environs.CreateParams{
		ControllerUUID: "daushdasd",
	})
	c.Assert(err, gc.NotNil)
}

func (e *environSuite) TestAdoptResources(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	err = environ.AdoptResources("", version.Number{})
	c.Assert(err, gc.IsNil)
}

func (e *environSuite) TestStopInstances(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	ids := []instance.Id{instance.Id("0")}
	err = environ.StopInstances(ids...)
	c.Assert(err, gc.IsNil)
}

func (e *environSuite) TestAllInstances(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	_, err = environ.AllInstances()
	c.Assert(err, gc.IsNil)
}

func (e *environSuite) TestMaintainInstance(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	err = environ.MaintainInstance(environs.StartInstanceParams{})
	c.Assert(err, gc.IsNil)
}

func (e *environSuite) TestConfig(c *gc.C) {
	environ, err := oracle.NewOracleEnviron(
		oracle.DefaultProvider,
		environs.OpenParams{
			Config: testing.ModelConfig(c),
		},
		DefaultEnvironAPI,
	)
	c.Assert(err, gc.IsNil)
	c.Assert(environ, gc.NotNil)

	cfg := environ.Config()
	c.Assert(cfg, gc.NotNil)
}
