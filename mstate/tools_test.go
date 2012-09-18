package mstate_test

import (
	"fmt"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/mstate"
	"launchpad.net/juju-core/version"
)

type tooler interface {
	AgentTools() (*mstate.Tools, error)
	SetAgentTools(t *mstate.Tools) error
	Kill() error
	Die() error
}

var _ = Suite(&ToolsSuite{})

type ToolsSuite struct {
	ConnSuite
}

func newTools(vers, url string) *mstate.Tools {
	return &mstate.Tools{
		Binary: version.MustParseBinary(vers),
		URL:    url,
	}
}

func testAgentTools(c *C, obj tooler, agent string) {
	// object starts with zero'd tools.
	t, err := obj.AgentTools()
	c.Assert(err, IsNil)
	c.Assert(t, DeepEquals, &mstate.Tools{})

	err = obj.SetAgentTools(&mstate.Tools{})
	c.Assert(err, ErrorMatches, fmt.Sprintf("cannot set agent tools for %s: empty series or arch", agent))
	t2 := newTools("7.8.9-foo-bar", "http://arble.tgz")
	err = obj.SetAgentTools(t2)
	c.Assert(err, IsNil)
	t3, err := obj.AgentTools()
	c.Assert(err, IsNil)
	c.Assert(t3, DeepEquals, t2)

	// Check we can set tools when the object is dying.
	err = obj.Kill()
	c.Assert(err, IsNil)
	err = obj.SetAgentTools(t2)
	c.Assert(err, IsNil)

	// Check we cannot set tools when the object is dead.
	err = obj.Die()
	c.Assert(err, IsNil)
	err = obj.SetAgentTools(t2)
	c.Assert(err, ErrorMatches, ".*: not found or not alive")
}

func (s *ToolsSuite) TestMachineAgentTools(c *C) {
	m, err := s.State.AddMachine()
	c.Assert(err, IsNil)
	testAgentTools(c, m, "machine 0")
}

func (s *ToolsSuite) TestUnitAgentTools(c *C) {
	charm := s.AddTestingCharm(c, "dummy")
	svc, err := s.State.AddService("wordpress", charm)
	c.Assert(err, IsNil)
	unit, err := svc.AddUnit()
	c.Assert(err, IsNil)
	testAgentTools(c, unit, "unit wordpress/0")
}
