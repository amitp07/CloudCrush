import { StringParameter } from 'aws-cdk-lib/aws-ssm';
import { SubnetType, Vpc } from 'aws-cdk-lib/aws-ec2';
import {Bucket} from 'aws-cdk-lib/aws-s3';
import * as cdk from 'aws-cdk-lib/core';
import { Construct } from 'constructs';

export class K8SInfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    cdk.Tags.of(this).add("project", "cloudcrush")
    cdk.Tags.of(this).add("env", "lab")


    const vpc = new Vpc(this, "cloudcrush-vpc", {
      maxAzs: 2,
      natGateways: 0,
      subnetConfiguration: [
        {
          name: "public",
          subnetType: SubnetType.PUBLIC,
          mapPublicIpOnLaunch: true
        }
      ]
    }) 

    const bucket = new Bucket(this, "cloudcrush-bucket", {
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      versioned: true,
      autoDeleteObjects: true,
    }) 

    const paramPrefix = "cloudcrush/sandbox/";
    new StringParameter(this, "BucketNameParam", {
      parameterName: `${paramPrefix}s3/bucketName`,
      stringValue: bucket.bucketName
    }) 

    new StringParameter(this, "VpcIdParam", {
      parameterName: `${paramPrefix}vpc/id`,
      stringValue: vpc.vpcId
    })

  }
}
