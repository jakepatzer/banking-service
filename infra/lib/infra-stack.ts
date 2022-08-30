import * as cdk from 'aws-cdk-lib';
import {Construct} from 'constructs';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import {AttributeType, BillingMode} from 'aws-cdk-lib/aws-dynamodb';
import * as iam from "aws-cdk-lib/aws-iam";
import {AccountPrincipal} from "aws-cdk-lib/aws-iam";
import * as lambdago from "@aws-cdk/aws-lambda-go-alpha";
import * as lambda from "aws-cdk-lib/aws-lambda";
import {FunctionUrlAuthType} from "aws-cdk-lib/aws-lambda";
import * as path from "path";

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
      super(scope, id, props);

      const accountsTable = new dynamodb.Table(this, 'AccountsTable', {
          tableName: 'accounts-table',
          partitionKey: {
              name: 'AccountId',
              type: AttributeType.STRING
          },
          sortKey: {
              name: 'AccountType',
              type: AttributeType.STRING
          },
          billingMode: BillingMode.PAY_PER_REQUEST
      });

      const dynamoDBAccessPolicy = new iam.PolicyStatement({
          actions: [
              'dynamodb:DeleteItem',
              'dynamodb:GetItem',
              'dynamodb:PutItem',
              'dynamodb:Query',
              'dynamodb:Scan',
              'dynamodb:UpdateItem'
          ],
          effect: iam.Effect.ALLOW,
          resources: [accountsTable.tableArn]
      })

      const createAccountLambda = new lambdago.GoFunction(this, 'create-account-function', {
          entry: path.join(__dirname, '../../lambda/functions/create-account'),
          functionName: 'create-account',
          initialPolicy: [
              new iam.PolicyStatement(dynamoDBAccessPolicy)
          ]
      })
      createAccountLambda.addPermission('resource-policy', {
          action: 'lambda:InvokeFunctionUrl',
          principal: new AccountPrincipal('*'),
          functionUrlAuthType: FunctionUrlAuthType.AWS_IAM
      })
      new lambda.FunctionUrl(this, 'create-account-url', {
          function: createAccountLambda,
          authType: lambda.FunctionUrlAuthType.AWS_IAM
      })

      const deleteAccountLambda = new lambdago.GoFunction(this, 'delete-account-function', {
          entry: path.join(__dirname, '../../lambda/functions/delete-account'),
          functionName: 'delete-account',
          initialPolicy: [
              new iam.PolicyStatement(dynamoDBAccessPolicy)
          ]
      })
      deleteAccountLambda.addPermission('resource-policy', {
          action: 'lambda:InvokeFunctionUrl',
          principal: new AccountPrincipal('*'),
          functionUrlAuthType: FunctionUrlAuthType.AWS_IAM
      })
      new lambda.FunctionUrl(this, 'delete-account-url', {
          function: deleteAccountLambda,
          authType: lambda.FunctionUrlAuthType.AWS_IAM
      })

      const getBalanceLambda = new lambdago.GoFunction(this, 'get-balance-function', {
          entry: path.join(__dirname, '../../lambda/functions/get-balance'),
          functionName: 'get-balance',
          initialPolicy: [
              new iam.PolicyStatement(dynamoDBAccessPolicy)
          ]
      })
      getBalanceLambda.addPermission('resource-policy', {
          action: 'lambda:InvokeFunctionUrl',
          principal: new AccountPrincipal('*'),
          functionUrlAuthType: FunctionUrlAuthType.AWS_IAM
      })
      new lambda.FunctionUrl(this, 'get-balance-url', {
          function: getBalanceLambda,
          authType: lambda.FunctionUrlAuthType.AWS_IAM
      })

      const listAccountsLambda = new lambdago.GoFunction(this, 'list-accounts-function', {
          entry: path.join(__dirname, '../../lambda/functions/list-accounts'),
          functionName: 'list-accounts',
          initialPolicy: [
              new iam.PolicyStatement(dynamoDBAccessPolicy)
          ]
      })
      listAccountsLambda.addPermission('resource-policy', {
          action: 'lambda:InvokeFunctionUrl',
          principal: new AccountPrincipal('*'),
          functionUrlAuthType: FunctionUrlAuthType.AWS_IAM
      })
      new lambda.FunctionUrl(this, 'list-accounts-url', {
          function: listAccountsLambda,
          authType: lambda.FunctionUrlAuthType.AWS_IAM
      })

      const transferLambda = new lambdago.GoFunction(this, 'transfer-function', {
          entry: path.join(__dirname, '../../lambda/functions/transfer'),
          functionName: 'transfer',
          initialPolicy: [
              new iam.PolicyStatement(dynamoDBAccessPolicy)
          ]
      })
      transferLambda.addPermission('resource-policy', {
          action: 'lambda:InvokeFunctionUrl',
          principal: new AccountPrincipal('*'),
          functionUrlAuthType: FunctionUrlAuthType.AWS_IAM
      })
      new lambda.FunctionUrl(this, 'transfer-url', {
          function: transferLambda,
          authType: lambda.FunctionUrlAuthType.AWS_IAM
      })

      // TODO: Add CloudTrail to log failed API calls, or use API Gateway which features CloudWatch logging

  }
}
