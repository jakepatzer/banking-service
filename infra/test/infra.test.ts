 import * as cdk from 'aws-cdk-lib';
 import { Template } from 'aws-cdk-lib/assertions';
 import * as Infra from '../lib/infra-stack';

test('DynamoDB Table Created', () => {
   const app = new cdk.App();
   const stack = new Infra.InfraStack(app, 'MyTestStack');
   const template = Template.fromStack(stack);

   template.hasResourceProperties('AWS::DynamoDB::Table', {
       TableName: 'accounts-table',
       KeySchema: [
           {
               AttributeName: "AccountId",
               KeyType: "HASH"
           },
           {
               AttributeName: "AccountType",
               KeyType: "RANGE"
           }
       ],
       AttributeDefinitions: [
           {
               AttributeName: "AccountId",
               AttributeType: "S"
           },
           {
               AttributeName: "AccountType",
               AttributeType: "S"
           }
       ],
       BillingMode: "PAY_PER_REQUEST"
   });

});

// TODO: Add additional unit tests
