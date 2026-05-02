import { StringParameter } from 'aws-cdk-lib/aws-ssm';
import { SubnetType, Vpc } from 'aws-cdk-lib/aws-ec2';
import {Bucket} from 'aws-cdk-lib/aws-s3';
import * as cdk from 'aws-cdk-lib/core';
import { Construct } from 'constructs';
import { aws_ec2, aws_eks, aws_iam } from 'aws-cdk-lib';
import {KubectlV35Layer} from '@aws-cdk/lambda-layer-kubectl-v35';

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

    const paramPrefix = "/cloudcrush/sandbox/";
    new StringParameter(this, "BucketNameParam", {
      parameterName: `${paramPrefix}s3/bucketName`,
      stringValue: bucket.bucketName
    }) 

    new StringParameter(this, "VpcIdParam", {
      parameterName: `${paramPrefix}vpc/id`,
      stringValue: vpc.vpcId
    })

    const cluster = new aws_eks.Cluster(this, "cloudcrush-cluster", {
      version: aws_eks.KubernetesVersion.V1_35,
      kubectlLayer: new KubectlV35Layer(this, "kubectl"),
      defaultCapacity: 0,
      clusterName: "cloudcrush-lab",
      endpointAccess: aws_eks.EndpointAccess.PUBLIC
    })

    cluster.addAutoScalingGroupCapacity("nodes", {
      instanceType: new aws_ec2.InstanceType("t3.small"),
      minCapacity: 1,
      maxCapacity: 2,
      vpcSubnets: {subnetType: aws_ec2.SubnetType.PUBLIC},
      associatePublicIpAddress: true
    })

    const ghPat = process.env.GH_PACKAGE_READ_PAT;
    const ghUsername = process.env.GH_USERNAME;

    if (ghPat && ghUsername) {
      const authString = Buffer.from(`${ghUsername}:${ghPat}`).toString("base64");
      const dockerConfig = {
        auths: {
          "ghcr.io": {
            auth: authString
          }
        }
      }

      cluster.addManifest("ghcr-auth-secret", {
        apiVersion: "v1",
        kind: "Secret",
        metadata: {name : "ghcr-auth", namespace: "default"},
        type: 'kubernetes.io/dockerconfigjson',
        data: {
          '.dockerconfigjson': Buffer.from(JSON.stringify(dockerConfig)).toString("base64")
        }
      })
    }

    const podRole = new aws_iam.Role(this, "cloudcrush-pod-role", {
      assumedBy: new aws_iam.WebIdentityPrincipal(cluster.openIdConnectProvider.openIdConnectProviderArn, {
        StringEquals: new cdk.CfnJson(this, "Condition", {
          value: {
            [`${cluster.openIdConnectProvider.openIdConnectProviderIssuer}:sub`]: 'system:serviceaccount:default:cloudcrush-sa'
          }
        }) 
      })
    })

    bucket.grantReadWrite(podRole)
    podRole.addManagedPolicy(aws_iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonSSMReadOnlyAccess'));

    cluster.addManifest('cloudcrush-sa', {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: {
        name: 'cloudcrush-sa',
        namespace: 'default',
        annotations: { 'eks.amazonaws.com/role-arn': podRole.roleArn }
      }
    });
  }
}
