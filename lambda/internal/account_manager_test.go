package internal

/*
import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
)

func TestSuite(t *testing.T) {
	suite.Run(t, new(accountManagerTestSuite))
}

type accountManagerTestSuite struct {
	suite.Suite
	ddb          *dynamodb.Client
	ddbContainer testcontainers.Container
}

func (suite *accountManagerTestSuite) SetupSuite() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local:latest",
		Cmd:          []string{"-jar", "DynamoDBLocal.jar", "-sharedDb"},
		ExposedPorts: []string{"8000/tcp"},
		WaitingFor:   wait.NewHostPortStrategy("8000"),
	}

	d, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		panic(err)
	}

	suite.ddbContainer = d

	resolver := aws.EndpointResolverWithOptionsFunc(func(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           "http://localhost:8000",
			SigningRegion: "localhost",
		}, nil
	})
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(resolver),
		config.WithRegion("localhost"))

	suite.ddb = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Credentials = credentials.NewStaticCredentialsProvider("fakekey", "fakesecretkey", "")
	})
}

func (suite *accountManagerTestSuite) TearDownSuite() {
	suite.ddbContainer.Terminate(context.Background())
}

func (suite *accountManagerTestSuite) TestTable() {
	attrDefinitions := []types.AttributeDefinition{
		{
			AttributeName: aws.String(accountIDAttr),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: aws.String(accountTypeAttr),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: aws.String(accountIDAttr),
			KeyType:       types.KeyTypeHash,
		},
		{
			AttributeName: aws.String(accountTypeAttr),
			KeyType:       types.KeyTypeRange,
		},
	}
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: attrDefinitions,
		KeySchema:            keySchema,
		TableName:            aws.String(tableName),
	}
	_, err := suite.ddb.CreateTable(context.Background(), input)
	assert.NoError(suite.T(), err)

	accountManager := NewAccountManager(suite.ddb)
	err = accountManager.CreateAccount(context.Background(), "test-account-id", "account-type", 5)
	assert.NoError(suite.T(), err)

	bal, err := accountManager.GetBalance(context.Background(), "test-account-id", "account-type")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), bal, 5)
}

*/
