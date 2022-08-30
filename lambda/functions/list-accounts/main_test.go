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
	testAccountID      = "123456789"
	testAdminAccountID = "105343117262"
)

type listAccountsTestSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	mockAccountManager *mocks.MockAccountManager
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(listAccountsTestSuite))
}

func (suite *listAccountsTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockAccountManager = mocks.NewMockAccountManager(suite.ctrl)
}

func (suite *listAccountsTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *listAccountsTestSuite) TestHandler_Success() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.ListAccountsInput{
		ExclusiveStartKey: &internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
		Limit: aws.Int32(10),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	expectedOutput := internal.ListAccountsOutput{
		Accounts: []internal.AccountKey{
			{
				AccountID:   testAccountID,
				AccountType: "savings",
			},
		},
		LastEvaluatedKey: internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
	}
	responseBody, err := json.Marshal(expectedOutput)
	assert.NoError(suite.T(), err)

	suite.mockAccountManager.EXPECT().ListAccounts(ctx, testAccountID, expectedInput).Return(expectedOutput, nil)
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, response.StatusCode)
	assert.Equal(suite.T(), string(responseBody), response.Body)
}

func (suite *listAccountsTestSuite) TestHandler_SuccessWhenAdmin() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.ListAccountsInput{
		ExclusiveStartKey: &internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
		Limit: aws.Int32(10),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAdminAccountID, string(requestBody))

	expectedOutput := internal.ListAccountsOutput{
		Accounts: []internal.AccountKey{
			{
				AccountID:   testAccountID,
				AccountType: "savings",
			},
		},
		LastEvaluatedKey: internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
	}
	responseBody, err := json.Marshal(expectedOutput)
	assert.NoError(suite.T(), err)

	suite.mockAccountManager.EXPECT().ListAccountsAdmin(ctx, expectedInput).Return(expectedOutput, nil)
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, response.StatusCode)
	assert.Equal(suite.T(), string(responseBody), response.Body)
}

func (suite *listAccountsTestSuite) TestHandler_SuccessWhenExclusiveStartKeyIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.ListAccountsInput{
		Limit: aws.Int32(10),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	expectedOutput := internal.ListAccountsOutput{
		Accounts: []internal.AccountKey{
			{
				AccountID:   testAccountID,
				AccountType: "savings",
			},
		},
		LastEvaluatedKey: internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
	}
	responseBody, err := json.Marshal(expectedOutput)
	assert.NoError(suite.T(), err)

	suite.mockAccountManager.EXPECT().ListAccounts(ctx, testAccountID, expectedInput).Return(expectedOutput, nil)
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, response.StatusCode)
	assert.Equal(suite.T(), string(responseBody), response.Body)
}

func (suite *listAccountsTestSuite) TestHandler_SuccessWhenLimitIsUndefined() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.ListAccountsInput{
		ExclusiveStartKey: &internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	expectedOutput := internal.ListAccountsOutput{
		Accounts: []internal.AccountKey{
			{
				AccountID:   testAccountID,
				AccountType: "savings",
			},
		},
		LastEvaluatedKey: internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
	}
	responseBody, err := json.Marshal(expectedOutput)
	assert.NoError(suite.T(), err)

	suite.mockAccountManager.EXPECT().ListAccounts(ctx, testAccountID, expectedInput).Return(expectedOutput, nil)
	accountManager = suite.mockAccountManager

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200, response.StatusCode)
	assert.Equal(suite.T(), string(responseBody), response.Body)
}

func (suite *listAccountsTestSuite) TestHandler_UnmarshalRequestError() {
	// === Given ===
	ctx := context.Background()
	request := getRequest(testAccountID, "}invalidJSON{")

	// === When ===
	response, err := handler(ctx, request)

	// === Then ===
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400, response.StatusCode)
}

func (suite *listAccountsTestSuite) TestHandler_InternalError() {
	// === Given ===
	ctx := context.Background()
	expectedInput := internal.ListAccountsInput{
		ExclusiveStartKey: &internal.AccountKey{
			AccountID:   testAccountID,
			AccountType: "savings",
		},
		Limit: aws.Int32(10),
	}
	requestBody, err := json.Marshal(expectedInput)
	assert.NoError(suite.T(), err)
	request := getRequest(testAccountID, string(requestBody))

	suite.mockAccountManager.EXPECT().ListAccounts(ctx, testAccountID, expectedInput).Return(internal.ListAccountsOutput{}, errors.New("ERROR"))
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
