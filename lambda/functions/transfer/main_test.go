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

type transferTestSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	mockAccountManager *mocks.MockAccountManager
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(transferTestSuite))
}

func (suite *transferTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockAccountManager = mocks.NewMockAccountManager(suite.ctrl)
}

func (suite *transferTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *transferTestSuite) TestHandler_Success() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType:  "savings",
		DestAccountID:   testAccountID,
		DestAccountType: "checking",
		Amount:          aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().Transfer(ctx, testAccountID, expectedInput).Return(nil)
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, response.StatusCode)
}

func (suite *transferTestSuite) TestHandler_UnmarshalRequestError() {
	// === Given ===
	ctx := context.Background()
	request := getRequest(testAccountID, "}invalidJSON{")

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *transferTestSuite) TestHandler_ErrorWhenSrcAccountTypeIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		DestAccountID:   testAccountID,
		DestAccountType: "checking",
		Amount:          aws.Int(5),
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

func (suite *transferTestSuite) TestHandler_ErrorWhenDestAccountIDIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType:  "savings",
		DestAccountType: "checking",
		Amount:          aws.Int(5),
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

func (suite *transferTestSuite) TestHandler_ErrorWhenDestAccountTypeIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType: "savings",
		DestAccountID:  testAccountID,
		Amount:         aws.Int(5),
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

func (suite *transferTestSuite) TestHandler_ErrorWhenAmountIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType:  "savings",
		DestAccountID:   testAccountID,
		DestAccountType: "checking",
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

func (suite *transferTestSuite) TestHandler_ErrorWhenAmountIsInvalid() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType:  "savings",
		DestAccountID:   testAccountID,
		DestAccountType: "checking",
		Amount:          aws.Int(0),
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

func (suite *transferTestSuite) TestHandler_ErrorWhenAccountHasInsufficientBalance() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType:  "savings",
		DestAccountID:   testAccountID,
		DestAccountType: "checking",
		Amount:          aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().Transfer(ctx, testAccountID, expectedInput).Return(internal.InsufficientFundsError{})
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *transferTestSuite) TestHandler_ErrorWhenDestAccountDoesNotExist() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType:  "savings",
		DestAccountID:   testAccountID,
		DestAccountType: "checking",
		Amount:          aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().Transfer(ctx, testAccountID, expectedInput).Return(internal.AccountDoesNotExistError{})
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *transferTestSuite) TestHandler_InternalError() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.TransferInput{
		SrcAccountType:  "savings",
		DestAccountID:   testAccountID,
		DestAccountType: "checking",
		Amount:          aws.Int(5),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().Transfer(ctx, testAccountID, expectedInput).Return(errors.New("ERROR"))
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
