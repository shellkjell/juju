// Copyright 2019 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package exec_test

import (
	"bytes"
	"net/url"
	"time"

	"github.com/golang/mock/gomock"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/juju/juju/caas/kubernetes/provider/exec"
	coretesting "github.com/juju/juju/testing"
)

type execSuite struct {
	BaseSuite
}

var _ = gc.Suite(&execSuite{})

func (s *execSuite) TestExecParamsValidateComandsAndPodName(c *gc.C) {
	ctrl := s.setupExecClient(c)
	defer ctrl.Finish()

	type testcase struct {
		Params  exec.ExecParams
		Err     string
		PodName string
	}

	for _, tc := range []testcase{
		{
			Params: exec.ExecParams{},
			Err:    "empty commands not valid",
		},
		{
			Params: exec.ExecParams{
				Commands: []string{"echo", "'hello world'"},
				PodName:  "",
			},
			Err: `podName "" not valid`,
		},
		{
			Params: exec.ExecParams{
				Commands: []string{"echo", "'hello world'"},
				PodName:  "cm/gitlab-k8s-0",
			},
			Err: `podName "cm/gitlab-k8s-0" not valid`,
		},
		{
			Params: exec.ExecParams{
				Commands: []string{"echo", "'hello world'"},
				PodName:  "cm/",
			},
			Err: `podName "cm/" not valid`,
		},
		{
			Params: exec.ExecParams{
				Commands: []string{"echo", "'hello world'"},
				PodName:  "pod/",
			},
			Err: `podName "pod/" not valid`,
		},
	} {
		c.Check(tc.Params.Validate(s.mockPodGetter), gc.ErrorMatches, tc.Err)
	}

}

func (s *execSuite) TestProcessEnv(c *gc.C) {
	ctrl := s.setupExecClient(c)
	defer ctrl.Finish()

	c.Assert(exec.ProcessEnv(
		[]string{
			"AAA=1", "BBB=1", "CCC=1", "DDD=1", "EEE=1",
		},
	), gc.Equals, "export AAA=1; export BBB=1; export CCC=1; export DDD=1; export EEE=1; ")
}

func (s *execSuite) TestExecParamsValidatePodContainerExistence(c *gc.C) {
	ctrl := s.setupExecClient(c)
	defer ctrl.Finish()

	// failed - completed pod.
	params := exec.ExecParams{
		Commands: []string{"echo", "'hello world'"},
		PodName:  "gitlab-k8s-uid",
	}
	pod := core.Pod{}
	pod.SetUID("gitlab-k8s-uid")
	pod.SetName("gitlab-k8s-0")
	pod.Status = core.PodStatus{
		Phase: core.PodSucceeded,
	}
	gomock.InOrder(
		s.mockPodGetter.EXPECT().Get("gitlab-k8s-uid", metav1.GetOptions{}).Times(1).
			Return(nil, s.k8sNotFoundError()),
		s.mockPodGetter.EXPECT().List(metav1.ListOptions{}).Times(1).
			Return(&core.PodList{Items: []core.Pod{pod}}, nil),
	)
	c.Assert(params.Validate(s.mockPodGetter), gc.ErrorMatches, "cannot exec into a container in a completed pod; current phase is Succeeded")

	// failed - failed pod
	params = exec.ExecParams{
		Commands: []string{"echo", "'hello world'"},
		PodName:  "gitlab-k8s-uid",
	}
	pod = core.Pod{}
	pod.SetUID("gitlab-k8s-uid")
	pod.SetName("gitlab-k8s-0")
	pod.Status = core.PodStatus{
		Phase: core.PodFailed,
	}
	gomock.InOrder(
		s.mockPodGetter.EXPECT().Get("gitlab-k8s-uid", metav1.GetOptions{}).Times(1).
			Return(nil, s.k8sNotFoundError()),
		s.mockPodGetter.EXPECT().List(metav1.ListOptions{}).Times(1).
			Return(&core.PodList{Items: []core.Pod{pod}}, nil),
	)
	c.Assert(params.Validate(s.mockPodGetter), gc.ErrorMatches, "cannot exec into a container in a completed pod; current phase is Failed")

	// failed - containerName not found
	params = exec.ExecParams{
		Commands:      []string{"echo", "'hello world'"},
		PodName:       "gitlab-k8s-uid",
		ContainerName: "non-existing-container-name",
	}
	pod = core.Pod{}
	pod.SetUID("gitlab-k8s-uid")
	pod.SetName("gitlab-k8s-0")
	gomock.InOrder(
		s.mockPodGetter.EXPECT().Get("gitlab-k8s-uid", metav1.GetOptions{}).Times(1).
			Return(nil, s.k8sNotFoundError()),
		s.mockPodGetter.EXPECT().List(metav1.ListOptions{}).Times(1).
			Return(&core.PodList{Items: []core.Pod{pod}}, nil),
	)
	c.Assert(params.Validate(s.mockPodGetter), gc.ErrorMatches, `container "non-existing-container-name" not found`)

	// all good - container name specified.
	params = exec.ExecParams{
		Commands:      []string{"echo", "'hello world'"},
		PodName:       "gitlab-k8s-uid",
		ContainerName: "gitlab-container",
	}
	pod = core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{Name: "gitlab-container"},
			},
		},
	}
	pod.SetUID("gitlab-k8s-uid")
	pod.SetName("gitlab-k8s-0")
	gomock.InOrder(
		s.mockPodGetter.EXPECT().Get("gitlab-k8s-uid", metav1.GetOptions{}).Times(1).
			Return(nil, s.k8sNotFoundError()),
		s.mockPodGetter.EXPECT().List(metav1.ListOptions{}).Times(1).
			Return(&core.PodList{Items: []core.Pod{pod}}, nil),
	)
	c.Assert(params.Validate(s.mockPodGetter), jc.ErrorIsNil)

	// all good - no container name specified, pick the 1st container.
	params = exec.ExecParams{
		Commands: []string{"echo", "'hello world'"},
		PodName:  "gitlab-k8s-uid",
	}
	c.Assert(params.ContainerName, gc.Equals, "")
	pod = core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{Name: "gitlab-container"},
			},
		},
	}
	pod.SetUID("gitlab-k8s-uid")
	pod.SetName("gitlab-k8s-0")
	gomock.InOrder(
		s.mockPodGetter.EXPECT().Get("gitlab-k8s-uid", metav1.GetOptions{}).Times(1).
			Return(nil, s.k8sNotFoundError()),
		s.mockPodGetter.EXPECT().List(metav1.ListOptions{}).Times(1).
			Return(&core.PodList{Items: []core.Pod{pod}}, nil),
	)
	c.Assert(params.Validate(s.mockPodGetter), jc.ErrorIsNil)
	c.Assert(params.ContainerName, gc.Equals, "gitlab-container")
}

func (s *execSuite) TestExec(c *gc.C) {
	ctrl := s.setupExecClient(c)
	defer ctrl.Finish()

	var stdin, stdout, stderr bytes.Buffer
	params := exec.ExecParams{
		Commands: []string{"echo", "'hello world'"},
		PodName:  "gitlab-k8s-uid",
		Stdout:   &stdout,
		Stderr:   &stderr,
		Stdin:    &stdin,
	}
	c.Assert(params.ContainerName, gc.Equals, "")
	pod := core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{Name: "gitlab-container"},
			},
		},
	}
	pod.SetUID("gitlab-k8s-uid")
	pod.SetName("gitlab-k8s-0")

	request := rest.NewRequest(
		nil,
		"POST",
		&url.URL{Path: "/path/"},
		"",
		rest.ContentConfig{GroupVersion: &core.SchemeGroupVersion},
		rest.Serializers{},
		nil,
		nil,
		0,
	).Resource("pods").Name("gitlab-k8s-0").Namespace("test").
		SubResource("exec").Param("container", "gitlab-container").VersionedParams(
		&core.PodExecOptions{
			Container: "gitlab-container",
			Command:   []string{""},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	gomock.InOrder(
		s.mockPodGetter.EXPECT().Get("gitlab-k8s-uid", metav1.GetOptions{}).Times(1).
			Return(nil, s.k8sNotFoundError()),
		s.mockPodGetter.EXPECT().List(metav1.ListOptions{}).Times(1).
			Return(&core.PodList{Items: []core.Pod{pod}}, nil),

		s.restClient.EXPECT().Post().Return(request),
		s.mockRemoteCmdExecutor.EXPECT().Stream(
			remotecommand.StreamOptions{
				Stdin:  &stdin,
				Stdout: &stdout,
				Stderr: &stderr,
				Tty:    false,
			},
		).Times(1).Return(nil),
	)

	cancel := make(<-chan struct{}, 1)
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.execClient.Exec(params, cancel)
	}()

	select {
	case err := <-errChan:
		c.Assert(err, jc.ErrorIsNil)
	case <-time.After(coretesting.LongWait):
		c.Fatalf("timed out waiting for Exec return")
	}
}