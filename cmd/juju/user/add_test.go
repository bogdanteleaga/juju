// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package user_test

import (
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/names"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/api/base"
	"github.com/juju/juju/apiserver/common"
	"github.com/juju/juju/cmd/juju/user"
	"github.com/juju/juju/testing"
)

// All of the functionality of the AddUser api call is contained elsewhere.
// This suite provides basic tests for the "add-user" command
type UserAddCommandSuite struct {
	BaseSuite
	mockAPI *mockAddUserAPI
}

var _ = gc.Suite(&UserAddCommandSuite{})

func (s *UserAddCommandSuite) SetUpTest(c *gc.C) {
	s.BaseSuite.SetUpTest(c)
	s.mockAPI = &mockAddUserAPI{}
	s.mockAPI.secretKey = []byte(strings.Repeat("X", 32))
}

func (s *UserAddCommandSuite) run(c *gc.C, args ...string) (*cmd.Context, error) {
	addCommand, _ := user.NewAddCommandForTest(s.mockAPI, s.store, &mockModelApi{})
	return testing.RunCommand(c, addCommand, args...)
}

func (s *UserAddCommandSuite) TestInit(c *gc.C) {
	for i, test := range []struct {
		args        []string
		user        string
		displayname string
		models      string
		acl         string
		outPath     string
		errorString string
	}{{
		errorString: "no username supplied",
	}, {
		args: []string{"foobar"},
		user: "foobar",
	}, {
		args:        []string{"foobar", "Foo Bar"},
		user:        "foobar",
		displayname: "Foo Bar",
	}, {
		args:        []string{"foobar", "Foo Bar", "extra"},
		errorString: `unrecognized args: \["extra"\]`,
	}, {
		args:   []string{"foobar", "--models", "foo,bar", "--acl=read"},
		user:   "foobar",
		models: "foo,bar",
		acl:    "read",
	}, {
		args:   []string{"foobar", "--models", "baz", "--acl=write"},
		user:   "foobar",
		models: "baz",
		acl:    "write",
	}} {
		c.Logf("test %d (%q)", i, test.args)
		wrappedCommand, command := user.NewAddCommandForTest(s.mockAPI, s.store, &mockModelApi{})
		err := testing.InitCommand(wrappedCommand, test.args)
		if test.errorString == "" {
			c.Check(err, jc.ErrorIsNil)
			c.Check(command.User, gc.Equals, test.user)
			c.Check(command.DisplayName, gc.Equals, test.displayname)
			if len(test.models) > 0 {
				c.Check(command.ModelNames, gc.Equals, test.models)
			}
			if test.acl != "" {
				c.Check(command.ModelAccess, gc.Equals, test.acl)
			}
		} else {
			c.Check(err, gc.ErrorMatches, test.errorString)
		}
	}
}

func (s *UserAddCommandSuite) TestAddUserWithUsername(c *gc.C) {
	context, err := s.run(c, "foobar")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.mockAPI.username, gc.Equals, "foobar")
	c.Assert(s.mockAPI.displayname, gc.Equals, "")
	c.Assert(s.mockAPI.access, gc.Equals, "read")
	c.Assert(s.mockAPI.models, gc.HasLen, 0)
	expected := `
User "foobar" added
Please send this command to foobar:
    juju register MD0TBmZvb2JhcjAREw8xMjcuMC4wLjE6MTIzNDUEIFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhY

"foobar" has not been granted access to any models. You can use "juju grant" to grant access.
`[1:]
	c.Assert(testing.Stdout(context), gc.Equals, expected)
	c.Assert(testing.Stderr(context), gc.Equals, "")
}

func (s *UserAddCommandSuite) TestAddUserWithUsernameAndACL(c *gc.C) {
	context, err := s.run(c, "--acl", "write", "foobar")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.mockAPI.username, gc.Equals, "foobar")
	c.Assert(s.mockAPI.displayname, gc.Equals, "")
	c.Assert(s.mockAPI.access, gc.Equals, "write")
	c.Assert(s.mockAPI.models, gc.HasLen, 0)
	expected := `
User "foobar" added
Please send this command to foobar:
    juju register MD0TBmZvb2JhcjAREw8xMjcuMC4wLjE6MTIzNDUEIFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhY

"foobar" has not been granted access to any models. You can use "juju grant" to grant access.
`[1:]
	c.Assert(testing.Stdout(context), gc.Equals, expected)
	c.Assert(testing.Stderr(context), gc.Equals, "")
}

func (s *UserAddCommandSuite) TestAddUserWithUsernameAndDisplayname(c *gc.C) {
	context, err := s.run(c, "foobar", "Foo Bar")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.mockAPI.username, gc.Equals, "foobar")
	c.Assert(s.mockAPI.displayname, gc.Equals, "Foo Bar")
	c.Assert(s.mockAPI.access, gc.Equals, "read")
	c.Assert(s.mockAPI.models, gc.HasLen, 0)
	expected := `
User "Foo Bar (foobar)" added
Please send this command to foobar:
    juju register MD0TBmZvb2JhcjAREw8xMjcuMC4wLjE6MTIzNDUEIFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhY

"Foo Bar (foobar)" has not been granted access to any models. You can use "juju grant" to grant access.
`[1:]
	c.Assert(testing.Stdout(context), gc.Equals, expected)
	c.Assert(testing.Stderr(context), gc.Equals, "")
}

type mockModelApi struct{}

func (m *mockModelApi) ListModels(user string) ([]base.UserModel, error) {
	return []base.UserModel{{Name: "model", UUID: "modeluuid"}}, nil
}

func (m *mockModelApi) Close() error {
	return nil
}

func (s *UserAddCommandSuite) TestAddUserWithModelAccess(c *gc.C) {
	context, err := s.run(c, "foobar", "--models", "model")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s.mockAPI.username, gc.Equals, "foobar")
	c.Assert(s.mockAPI.displayname, gc.Equals, "")
	c.Assert(s.mockAPI.access, gc.Equals, "read")
	c.Assert(s.mockAPI.models, gc.DeepEquals, []string{"modeluuid"})
	expected := `
User "foobar" added
User "foobar" granted read access to model "model"
Please send this command to foobar:
    juju register MD0TBmZvb2JhcjAREw8xMjcuMC4wLjE6MTIzNDUEIFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhY
`[1:]
	c.Assert(testing.Stdout(context), gc.Equals, expected)
	c.Assert(testing.Stderr(context), gc.Equals, "")
}

func (s *UserAddCommandSuite) TestBlockAddUser(c *gc.C) {
	// Block operation
	s.mockAPI.blocked = true
	_, err := s.run(c, "foobar", "Foo Bar")
	c.Assert(err, gc.ErrorMatches, cmd.ErrSilent.Error())
	// msg is logged
	stripped := strings.Replace(c.GetTestLog(), "\n", "", -1)
	c.Check(stripped, gc.Matches, ".*To unblock changes.*")
}

func (s *UserAddCommandSuite) TestAddUserErrorResponse(c *gc.C) {
	s.mockAPI.failMessage = "failed to create user, chaos ensues"
	_, err := s.run(c, "foobar")
	c.Assert(err, gc.ErrorMatches, s.mockAPI.failMessage)
}

type mockAddUserAPI struct {
	failMessage string
	blocked     bool
	secretKey   []byte

	username    string
	displayname string
	password    string
	access      string
	models      []string
}

func (m *mockAddUserAPI) AddUser(username, displayname, password, access string, models ...string) (names.UserTag, []byte, error) {
	if m.blocked {
		return names.UserTag{}, nil, common.OperationBlockedError("the operation has been blocked")
	}
	m.username = username
	m.displayname = displayname
	m.password = password
	m.access = access
	m.models = models
	if m.failMessage != "" {
		return names.UserTag{}, nil, errors.New(m.failMessage)
	}
	return names.NewLocalUserTag(username), m.secretKey, nil
}

func (*mockAddUserAPI) Close() error {
	return nil
}
