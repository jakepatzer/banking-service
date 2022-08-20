import * as cdk from 'aws-cdk-lib';
import {Construct} from 'constructs';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import {AttributeType, BillingMode} from 'aws-cdk-lib/aws-dynamodb';

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

  }
}
