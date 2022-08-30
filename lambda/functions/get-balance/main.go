package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/jakepatzer/banking-service/lambda/functions"
	"github.com/jakepatzer/banking-service/lambda/internal"
	"log"
	"os"
)

var accountManager internal.AccountManager
var inputValidator *validator.Validate
var translator ut.Translator

func init() {
	cfg, _ := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("AWS_REGION")))
	ddb := dynamodb.NewFromConfig(cfg)
	accountManager = internal.NewAccountManager(ddb)

	inputValidator = validator.New()

	english := en.New()
	uni := ut.New(english, english)
	var ok bool
	translator, ok = uni.GetTranslator("en")
	if !ok {
		panic("Failed to initialize translator!")
	}
	err := enTranslations.RegisterDefaultTranslations(inputValidator, translator)
	if err != nil {
		panic(err)
	}
}

func handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// TODO: Gracefully handle timeouts based on Lambda function deadline
	accountID := request.RequestContext.Authorizer.IAM.AccountID

	log.Printf("Recieved request from account ID %s: %s", accountID, request.Body)

	var input internal.GetBalanceInput
	err := json.Unmarshal([]byte(request.Body), &input)
	if err != nil {
		requestErr := functions.RequestError{
			AccountID:   accountID,
			RequestBody: request.Body,
			Err:         err.Error(),
		}
		log.Print(requestErr)
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       "Error parsing the provided request",
		}, nil
	}

	err = inputValidator.Struct(input)
	if err != nil {
		requestErr := functions.RequestError{
			AccountID:   accountID,
			RequestBody: request.Body,
			Err:         err.Error(),
		}
		log.Print(requestErr)
		return processError(err), nil
	}

	output, err := accountManager.GetBalance(ctx, accountID, input)
	if err != nil {
		requestErr := functions.RequestError{
			AccountID:   accountID,
			RequestBody: request.Body,
			Err:         err.Error(),
		}
		log.Print(requestErr)
		return processError(err), nil
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Body:       functions.MarshalOutput(output),
	}, nil
}

func processError(err error) events.LambdaFunctionURLResponse {
	var accountDoesNotExistErr internal.AccountDoesNotExistError
	var validationErrs validator.ValidationErrors
	if errors.As(err, &accountDoesNotExistErr) {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       accountDoesNotExistErr.Error(),
		}
	} else if errors.As(err, &validationErrs) {
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("Invalid request: %v", validationErrs.Translate(translator)),
		}
	} else {
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       "Internal error",
		}
	}
}

func main() {
	lambda.Start(handler)
}
