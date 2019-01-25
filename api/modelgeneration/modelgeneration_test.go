// Copyright 2019 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package modelgeneration_test

import (
	"github.com/golang/mock/gomock"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v2"

	"github.com/juju/juju/api/base/mocks"
	"github.com/juju/juju/api/modelgeneration"
	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/core/model"
)

type modelGenerationSuite struct {
	tag     names.ModelTag
	fCaller *mocks.MockFacadeCaller
}

var _ = gc.Suite(&modelGenerationSuite{})

func (s *modelGenerationSuite) SetUpTest(c *gc.C) {
	s.tag = names.NewModelTag("deadbeef-abcd-4fd2-967d-db9663db7bea")
}

func (s *modelGenerationSuite) TeadDownTest(c *gc.C) {
	s.fCaller = nil
}

func (s *modelGenerationSuite) setUpMocks(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)

	caller := mocks.NewMockAPICallCloser(ctrl)
	caller.EXPECT().BestFacadeVersion(gomock.Any()).Return(0).AnyTimes()

	s.fCaller = mocks.NewMockFacadeCaller(ctrl)
	s.fCaller.EXPECT().RawAPICaller().Return(caller).AnyTimes()

	return ctrl
}

func (s *modelGenerationSuite) TestAddGeneration(c *gc.C) {
	defer s.setUpMocks(c).Finish()

	resultSource := params.ErrorResult{}
	arg := params.Entity{Tag: s.tag.String()}

	s.fCaller.EXPECT().FacadeCall("AddGeneration", arg, gomock.Any()).SetArg(2, resultSource).Return(nil)

	api := modelgeneration.NewStateFromCaller(s.fCaller)
	err := api.AddGeneration(s.tag)
	c.Assert(err, gc.IsNil)
}

func (s *modelGenerationSuite) TestAdvanceGeneration(c *gc.C) {
	defer s.setUpMocks(c).Finish()

	resultsSource := params.ErrorResults{
		Results: []params.ErrorResult{
			{Error: nil},
			{Error: nil},
		},
	}
	arg := params.AdvanceGenerationArg{
		Model: params.Entity{Tag: s.tag.String()},
		Entities: []params.Entity{
			{Tag: "unit-mysql-0"},
			{Tag: "application-mysql"},
		},
	}

	s.fCaller.EXPECT().FacadeCall("AdvanceGeneration", arg, gomock.Any()).SetArg(2, resultsSource).Return(nil)

	api := modelgeneration.NewStateFromCaller(s.fCaller)
	err := api.AdvanceGeneration(s.tag, []string{"mysql/0", "mysql"})
	c.Assert(err, gc.IsNil)
}

func (s *modelGenerationSuite) TestAdvanceGenerationError(c *gc.C) {
	defer s.setUpMocks(c).Finish()

	api := modelgeneration.NewStateFromCaller(s.fCaller)
	err := api.AdvanceGeneration(s.tag, []string{"mysql/0", "mysql", "machine-3"})
	c.Assert(err, gc.ErrorMatches, "Must be application or unit")
}

func (s *modelGenerationSuite) TestCancelGeneration(c *gc.C) {
	defer s.setUpMocks(c).Finish()

	resultSource := params.ErrorResult{}
	arg := params.Entity{Tag: s.tag.String()}

	s.fCaller.EXPECT().FacadeCall("CancelGeneration", arg, gomock.Any()).SetArg(2, resultSource).Return(nil)

	api := modelgeneration.NewStateFromCaller(s.fCaller)
	err := api.CancelGeneration(s.tag)
	c.Assert(err, gc.IsNil)
}

func (s *modelGenerationSuite) TestSwitchGeneration(c *gc.C) {
	defer s.setUpMocks(c).Finish()

	resultSource := params.ErrorResult{}
	arg := params.GenerationVersionArg{
		Model:   params.Entity{Tag: s.tag.String()},
		Version: model.GenerationNext,
	}

	s.fCaller.EXPECT().FacadeCall("SwitchGeneration", arg, gomock.Any()).SetArg(2, resultSource).Return(nil)

	api := modelgeneration.NewStateFromCaller(s.fCaller)
	err := api.SwitchGeneration(s.tag, "next")
	c.Assert(err, gc.IsNil)
}

func (s *modelGenerationSuite) TestSwitchGenerationError(c *gc.C) {
	defer s.setUpMocks(c).Finish()

	api := modelgeneration.NewStateFromCaller(s.fCaller)
	err := api.SwitchGeneration(s.tag, "summer")
	c.Assert(err, gc.ErrorMatches, "version must be 'next' or 'current'")
}
