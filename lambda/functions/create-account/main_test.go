package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
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

type createAccountTestSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	mockAccountManager *mocks.MockAccountManager
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(createAccountTestSuite))
}

func (suite *createAccountTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockAccountManager = mocks.NewMockAccountManager(suite.ctrl)
}

func (suite *createAccountTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *createAccountTestSuite) TestHandler_Success() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.CreateAccountInput{
		AccountType:    "savings",
		InitialBalance: aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().CreateAccount(ctx, testAccountID, expectedInput).Return(nil)
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, response.StatusCode)
}

func (suite *createAccountTestSuite) TestHandler_UnmarshalRequestError() {
	// === Given ===
	ctx := context.Background()
	request := getRequest(testAccountID, "}invalidJSON{")

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *createAccountTestSuite) TestHandler_ErrorWhenInitialBalanceIsInvalid() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.CreateAccountInput{
		AccountType:    "savings",
		InitialBalance: aws.Int(-1),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *createAccountTestSuite) TestHandler_ErrorWhenInitialBalanceIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.CreateAccountInput{
		AccountType: "savings",
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *createAccountTestSuite) TestHandler_ErrorWhenAccountTypeIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.CreateAccountInput{
		InitialBalance: aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *createAccountTestSuite) TestHandler_ErrorWhenAccountAlreadyExists() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.CreateAccountInput{
		AccountType:    "savings",
		InitialBalance: aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().CreateAccount(ctx, testAccountID, expectedInput).Return(internal.AccountAlreadyExistsError{})
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *createAccountTestSuite) TestHandler_InternalError() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.CreateAccountInput{
		AccountType:    "savings",
		InitialBalance: aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().CreateAccount(ctx, testAccountID, expectedInput).Return(errors.New("ERROR"))
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
