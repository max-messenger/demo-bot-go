package state_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"demo_bot/internal/app/domain"
	"demo_bot/internal/app/services/state"
	"demo_bot/internal/app/services/state/mocks"
)

type stateServiceSuite struct {
	suite.Suite

	ctrl    *gomock.Controller
	repo    *mocks.MockRepository
	service state.Service
	ctx     context.Context
}

const testUserID int64 = 42

func TestUnitStateService(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(stateServiceSuite))
}

func (s *stateServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.repo = mocks.NewMockRepository(s.ctrl)
	s.service = state.NewService(s.repo)
	s.ctx = context.Background()
}

func (s *stateServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *stateServiceSuite) TestGetState_Success() {
	expected := &domain.ScenarioState{
		Scenario: "greeting",
		Step:     1,
		Data:     map[string]any{"key": "value"},
	}

	s.repo.EXPECT().
		GetState(s.ctx, testUserID).
		Return(expected, nil)

	got, err := s.service.GetState(s.ctx, testUserID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), expected, got)
}

func (s *stateServiceSuite) TestGetState_Error() {
	s.repo.EXPECT().
		GetState(s.ctx, testUserID).
		Return(nil, errTestGetState)

	got, err := s.service.GetState(s.ctx, testUserID)
	require.Error(s.T(), err)
	assert.Nil(s.T(), got)
	assert.ErrorIs(s.T(), err, errTestGetState)
}

func (s *stateServiceSuite) TestSetState_Success() {
	st := &domain.ScenarioState{
		Scenario: "onboarding",
		Step:     2,
	}

	s.repo.EXPECT().
		SetState(s.ctx, testUserID, st).
		Return(nil)

	err := s.service.SetState(s.ctx, testUserID, st)
	require.NoError(s.T(), err)
}

func (s *stateServiceSuite) TestSetState_Error() {
	st := &domain.ScenarioState{
		Scenario: "onboarding",
		Step:     2,
	}

	s.repo.EXPECT().
		SetState(s.ctx, testUserID, st).
		Return(errTestSetState)

	err := s.service.SetState(s.ctx, testUserID, st)
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errTestSetState)
}

func (s *stateServiceSuite) TestClearState_Success() {
	s.repo.EXPECT().
		ClearState(s.ctx, testUserID).
		Return(nil)

	err := s.service.ClearState(s.ctx, testUserID)
	require.NoError(s.T(), err)
}

func (s *stateServiceSuite) TestClearState_Error() {
	s.repo.EXPECT().
		ClearState(s.ctx, testUserID).
		Return(errTestClearState)

	err := s.service.ClearState(s.ctx, testUserID)
	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, errTestClearState)
}

var (
	errTestGetState   = errCustom("get state failed")
	errTestSetState   = errCustom("set state failed")
	errTestClearState = errCustom("clear state failed")
)

type errCustom string

func (e errCustom) Error() string { return string(e) }
