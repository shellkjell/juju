// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package worker_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/worker"
)

type finishedSuite struct {}

var _ = gc.Suite(&finishedSuite{})

func (s *finishedSuite) TestFinishedWorker(c *gc.C) {
	// Pretty dumb test if interface is implemented
	// and Wait() returns nil.
	var fw worker.Worker = worker.FinishedWorker{}
	c.Assert(fw.Wait(), gc.IsNil)
}
