package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang/mock/gomock"
	"github.com/jakepatzer/banking-service/lambda/internal"
	"github.com/jakepatzer/banking-service/lambda/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	testAccountID = "123456789"
)

type deleteAccountTestSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	mockAccountManager *mocks.MockAccountManager
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(deleteAccountTestSuite))
}

func (suite *deleteAccountTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockAccountManager = mocks.NewMockAccountManager(suite.ctrl)
}

func (suite *deleteAccountTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *deleteAccountTestSuite) TestHandler_Success() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.DeleteAccountInput{
		AccountType: "savings",
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().DeleteAccount(ctx, testAccountID, expectedInput).Return(nil)
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, response.StatusCode)
}

func (suite *deleteAccountTestSuite) TestHandler_UnmarshalRequestError() {
	// === Given ===
	ctx := context.Background()
	request := getRequest(testAccountID, "}invalidJSON{")

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *deleteAccountTestSuite) TestHandler_ErrorWhenAccountTypeIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.CreateAccountInput{}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *deleteAccountTestSuite) TestHandler_ErrorWhenAccountHasNonZeroBalance() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.DeleteAccountInput{
		AccountType: "savings",
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().DeleteAccount(ctx, testAccountID, expectedInput).Return(internal.NonZeroBalanceError{})
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *deleteAccountTestSuite) TestHandler_InternalError() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.DeleteAccountInput{
		AccountType: "savings",
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().DeleteAccount(ctx, testAccountID, expectedInput).Return(errors.New("ERROR"))
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 500, response.StatusCode)
}

func getRequest(accountID, requestBody string) events.LambdaFunctionURLRequest {
	return events.LambdaFunctionURLRequest{
		RequestContext: events.LambdaFunctionURLRequestContext{
			Authorizer: &events.LambdaFunctionURLRequestContextAuthorizerDescription{
				IAM: &events.LambdaFunctionURLRequestContextAuthorizerIAMDescription{
					AccountID: accountID,
				},
			},
		},
		Body: requestBody,
	}
}
