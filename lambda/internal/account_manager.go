package internal

//go:generate mockgen.exe -source ./account_manager.go -destination ../mocks/account_manager_mock.go -package mocks

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
)

const (
	tableName = "accounts-table"

	accountIDAttr   = "AccountId"
	accountTypeAttr = "AccountType"
	balanceAttr     = "Balance"
)

func NewAccountKeyFromItem(item map[string]types.AttributeValue) (AccountKey, error) {
	accountID, ok := item[accountIDAttr].(*types.AttributeValueMemberS)
	if !ok {
		return AccountKey{}, errors.New("accountID must be a string")
	}

	accountType, ok := item[accountTypeAttr].(*types.AttributeValueMemberS)
	if !ok {
		return AccountKey{}, errors.New("accountType must be a string")
	}

	return AccountKey{
		AccountID:   accountID.Value,
		AccountType: accountType.Value,
	}, nil
}

type AccountKey struct {
	AccountID   string `json:"accountID"`
	AccountType string `json:"accountType"`
}

func (key *AccountKey) toAccountItem() map[string]types.AttributeValue {
	item := make(map[string]types.AttributeValue)
	item[accountIDAttr] = &types.AttributeValueMemberS{Value: key.AccountID}
	item[accountTypeAttr] = &types.AttributeValueMemberS{Value: key.AccountType}
	return item
}

type AccountAlreadyExistsError struct {
	AccountID   string
	AccountType string
}

func (err AccountAlreadyExistsError) Error() string {
	return fmt.Sprintf("The account %s:%s already exists.", err.AccountID, err.AccountType)
}

type NonZeroBalanceError struct {
	AccountID   string
	AccountType string
}

func (err NonZeroBalanceError) Error() string {
	return fmt.Sprintf("The account %s:%s has a non-zero balance.", err.AccountID, err.AccountType)
}

type InsufficientFundsError struct {
	AccountID   string
	AccountType string
}

func (err InsufficientFundsError) Error() string {
	return fmt.Sprintf("The account %s:%s does not have sufficient funds.", err.AccountID, err.AccountType)
}

type AccountDoesNotExistError struct {
	AccountID   string
	AccountType string
}

func (err AccountDoesNotExistError) Error() string {
	return fmt.Sprintf("The account %s:%s does not exist.", err.AccountID, err.AccountType)
}

type AccountManager interface {
	CreateAccount(ctx context.Context, accountID string, createAccountInput CreateAccountInput) error
	DeleteAccount(ctx context.Context, accountID string, deleteAccountInput DeleteAccountInput) error
	Transfer(ctx context.Context, srcAccountID string, transferInput TransferInput) error
	GetBalance(ctx context.Context, accountID string, getBalanceInput GetBalanceInput) (GetBalanceOutput, error)
	ListAccounts(ctx context.Context, accountID string, listAccountsInput ListAccountsInput) (ListAccountsOutput, error)
	ListAccountsAdmin(ctx context.Context, listAccountsInput ListAccountsInput) (ListAccountsOutput, error)
}

func NewAccountManager(ddb *dynamodb.Client) AccountManager {
	return accountManagerImpl{ddb: ddb}
}

type accountManagerImpl struct {
	ddb *dynamodb.Client
}

type CreateAccountInput struct {
	// TODO: Ensure that AccountType is within the maximum length allowed by DynamoDB
	AccountType string `json:"accountType" validate:"required"`
	// Use pointer for InitialBalance to ensure that it's explicitly defined
	InitialBalance *int `json:"initialBalance" validate:"required,gte=0"`
}

func (manager accountManagerImpl) CreateAccount(ctx context.Context, accountID string, createAccountInput CreateAccountInput) error {
	item := make(map[string]types.AttributeValue)
	item[accountIDAttr] = &types.AttributeValueMemberS{Value: accountID}
	item[accountTypeAttr] = &types.AttributeValueMemberS{Value: createAccountInput.AccountType}
	item[balanceAttr] = &types.AttributeValueMemberN{Value: strconv.Itoa(*createAccountInput.InitialBalance)}

	input := &dynamodb.PutItemInput{
		Item:                item,
		TableName:           aws.String(tableName),
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s)", accountIDAttr)),
	}
	_, err := manager.ddb.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return AccountAlreadyExistsError{
				AccountID:   accountID,
				AccountType: createAccountInput.AccountType,
			}
		}
		return err
	}

	return nil
}

type DeleteAccountInput struct {
	AccountType string `json:"accountType" validate:"required"`
}

func (manager accountManagerImpl) DeleteAccount(ctx context.Context, accountID string, deleteAccountInput DeleteAccountInput) error {
	item := make(map[string]types.AttributeValue)
	item[accountIDAttr] = &types.AttributeValueMemberS{Value: accountID}
	item[accountTypeAttr] = &types.AttributeValueMemberS{Value: deleteAccountInput.AccountType}

	expressionAttributeValues := make(map[string]types.AttributeValue)
	expressionAttributeValues[":b"] = &types.AttributeValueMemberN{Value: "0"}

	input := &dynamodb.DeleteItemInput{
		Key:       item,
		TableName: aws.String(tableName),
		// Succeed if the account doesn't exist to simplify error handling and allow for idempotent calls
		ConditionExpression:       aws.String(fmt.Sprintf("%s = :b OR attribute_not_exists(%s)", balanceAttr, accountIDAttr)),
		ExpressionAttributeValues: expressionAttributeValues,
	}
	_, err := manager.ddb.DeleteItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return NonZeroBalanceError{
				AccountID:   accountID,
				AccountType: deleteAccountInput.AccountType,
			}
		}
		return err
	}

	return nil
}

type TransferInput struct {
	SrcAccountType  string `json:"srcAccountType" validate:"required"`
	DestAccountID   string `json:"destAccountID" validate:"required"`
	DestAccountType string `json:"destAccountType" validate:"required"`
	// Use pointer for Amount to ensure that it's explicitly defined
	Amount *int `json:"amount" validate:"gt=0"`
}

func (manager accountManagerImpl) Transfer(ctx context.Context, srcAccountID string, transferInput TransferInput) error {
	exprAttrValues := make(map[string]types.AttributeValue)
	exprAttrValues[":a"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*transferInput.Amount)}

	srcKey := make(map[string]types.AttributeValue)
	srcKey[accountIDAttr] = &types.AttributeValueMemberS{Value: srcAccountID}
	srcKey[accountTypeAttr] = &types.AttributeValueMemberS{Value: transferInput.SrcAccountType}
	srcItemTransaction := types.TransactWriteItem{
		Update: &types.Update{
			Key:                       srcKey,
			TableName:                 aws.String(tableName),
			UpdateExpression:          aws.String(fmt.Sprintf("SET %s = %s - :a", balanceAttr, balanceAttr)),
			ConditionExpression:       aws.String(fmt.Sprintf("attribute_exists(%s) and (%s >= :a)", accountIDAttr, balanceAttr)),
			ExpressionAttributeValues: exprAttrValues,
		},
	}

	destKey := make(map[string]types.AttributeValue)
	destKey[accountIDAttr] = &types.AttributeValueMemberS{Value: transferInput.DestAccountID}
	destKey[accountTypeAttr] = &types.AttributeValueMemberS{Value: transferInput.DestAccountType}
	destItemTransaction := types.TransactWriteItem{
		Update: &types.Update{
			Key:                       destKey,
			TableName:                 aws.String(tableName),
			UpdateExpression:          aws.String(fmt.Sprintf("SET %s = %s + :a", balanceAttr, balanceAttr)),
			ConditionExpression:       aws.String(fmt.Sprintf("attribute_exists(%s)", accountIDAttr)),
			ExpressionAttributeValues: exprAttrValues,
		},
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{srcItemTransaction, destItemTransaction},
	}

	_, err := manager.ddb.TransactWriteItems(ctx, input)
	if err != nil {
		var transactionCanceledException *types.TransactionCanceledException
		if errors.As(err, &transactionCanceledException) {

			// Index of cancellation reasons is dependent on the ordering of TransactWriteItem above
			conditionalCheckFailedException := &types.ConditionalCheckFailedException{}
			if *transactionCanceledException.CancellationReasons[0].Code == conditionalCheckFailedException.ErrorCode() {
				// TODO: Return a separate error if the source account does not exist
				return InsufficientFundsError{
					AccountID:   transferInput.DestAccountID,
					AccountType: transferInput.DestAccountType,
				}
			}

			if *transactionCanceledException.CancellationReasons[1].Code == conditionalCheckFailedException.ErrorCode() {
				return AccountDoesNotExistError{
					AccountID:   transferInput.DestAccountID,
					AccountType: transferInput.DestAccountType,
				}
			}
		}

		return err
	}

	return nil
}

type GetBalanceInput struct {
	AccountType string `json:"accountType" validate:"required"`
}

type GetBalanceOutput struct {
	Balance int `json:"balance"`
}

func (manager accountManagerImpl) GetBalance(ctx context.Context, accountID string, getBalanceInput GetBalanceInput) (GetBalanceOutput, error) {
	item := make(map[string]types.AttributeValue)
	item[accountIDAttr] = &types.AttributeValueMemberS{Value: accountID}
	item[accountTypeAttr] = &types.AttributeValueMemberS{Value: getBalanceInput.AccountType}

	input := &dynamodb.GetItemInput{
		Key:                  item,
		TableName:            aws.String(tableName),
		ConsistentRead:       aws.Bool(true),
		ProjectionExpression: aws.String(balanceAttr),
	}

	output, err := manager.ddb.GetItem(ctx, input)
	if err != nil {
		return GetBalanceOutput{}, err
	}

	if len(output.Item) == 0 {
		return GetBalanceOutput{}, AccountDoesNotExistError{
			AccountID:   accountID,
			AccountType: getBalanceInput.AccountType,
		}
	}

	attrValue, ok := output.Item[balanceAttr].(*types.AttributeValueMemberN)
	if !ok {
		return GetBalanceOutput{}, errors.New("balance must be a number")
	}

	val, err := strconv.Atoi(attrValue.Value)
	if err != nil {
		return GetBalanceOutput{}, err
	}

	return GetBalanceOutput{
		Balance: val,
	}, nil
}

// TODO: Validate fields if they are defined
type ListAccountsInput struct {
	ExclusiveStartKey *AccountKey `json:"exclusiveStartKey"`
	Limit             *int32      `json:"limit"`
}

type ListAccountsOutput struct {
	Accounts         []AccountKey `json:"accounts"`
	LastEvaluatedKey AccountKey   `json:"lastEvaluatedKey"`
}

func (manager accountManagerImpl) ListAccounts(ctx context.Context, accountID string, listAccountsInput ListAccountsInput) (ListAccountsOutput, error) {
	exprAttrValues := make(map[string]types.AttributeValue)
	exprAttrValues[":id"] = &types.AttributeValueMemberS{Value: accountID}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		ExpressionAttributeValues: exprAttrValues,
		KeyConditionExpression:    aws.String(fmt.Sprintf("%s = :id", accountIDAttr)),
		ProjectionExpression:      aws.String(fmt.Sprintf("%s,%s", accountIDAttr, accountTypeAttr)),
	}

	if listAccountsInput.ExclusiveStartKey != nil {
		input.ExclusiveStartKey = listAccountsInput.ExclusiveStartKey.toAccountItem()
	}
	if listAccountsInput.Limit != nil {
		input.Limit = listAccountsInput.Limit
	}

	output, err := manager.ddb.Query(ctx, input)
	if err != nil {
		return ListAccountsOutput{}, err
	}

	var accounts []AccountKey
	for _, item := range output.Items {
		accountKey, err := NewAccountKeyFromItem(item)
		if err != nil {
			return ListAccountsOutput{}, err
		}
		accounts = append(accounts, accountKey)
	}

	var lastEvaluatedKey AccountKey
	if len(output.LastEvaluatedKey) != 0 {
		lastEvaluatedKey, err = NewAccountKeyFromItem(output.LastEvaluatedKey)
		if err != nil {
			return ListAccountsOutput{}, err
		}
	}

	return ListAccountsOutput{
		Accounts:         accounts,
		LastEvaluatedKey: lastEvaluatedKey,
	}, nil
}

func (manager accountManagerImpl) ListAccountsAdmin(ctx context.Context, listAccountsInput ListAccountsInput) (ListAccountsOutput, error) {
	input := &dynamodb.ScanInput{
		TableName:            aws.String(tableName),
		ProjectionExpression: aws.String(fmt.Sprintf("%s,%s", accountIDAttr, accountTypeAttr)),
	}

	if listAccountsInput.ExclusiveStartKey != nil {
		input.ExclusiveStartKey = listAccountsInput.ExclusiveStartKey.toAccountItem()
	}
	if listAccountsInput.Limit != nil {
		input.Limit = listAccountsInput.Limit
	}

	output, err := manager.ddb.Scan(ctx, input)
	if err != nil {
		return ListAccountsOutput{}, err
	}

	var accounts []AccountKey
	for _, item := range output.Items {
		accountKey, err := NewAccountKeyFromItem(item)
		if err != nil {
			return ListAccountsOutput{}, err
		}
		accounts = append(accounts, accountKey)
	}

	var lastEvaluatedKey AccountKey
	if len(output.LastEvaluatedKey) != 0 {
		lastEvaluatedKey, err = NewAccountKeyFromItem(output.LastEvaluatedKey)
		if err != nil {
			return ListAccountsOutput{}, err
		}
	}

	return ListAccountsOutput{
		Accounts:         accounts,
		LastEvaluatedKey: lastEvaluatedKey,
	}, nil
}
