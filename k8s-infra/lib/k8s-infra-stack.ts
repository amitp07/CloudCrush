import { StringParameter } from 'aws-cdk-lib/aws-ssm';
import { SubnetType, Vpc } from 'aws-cdk-lib/aws-ec2';
import { Bucket } from 'aws-cdk-lib/aws-s3';
import * as cdk from 'aws-cdk-lib/core';
import { Construct } from 'constructs';
import { aws_ec2, aws_eks, aws_iam } from 'aws-cdk-lib';
import { KubectlV34Layer } from '@aws-cdk/lambda-layer-kubectl-v34';
import { AccessPolicy, AccessScopeType } from 'aws-cdk-lib/aws-eks';

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
      vpc: vpc,
      vpcSubnets: [{ subnetType: aws_ec2.SubnetType.PUBLIC }],
      version: aws_eks.KubernetesVersion.V1_34,
      kubectlLayer: new KubectlV34Layer(this, "kubectl"),
      defaultCapacity: 0,
      clusterName: "cloudcrush-lab",
      endpointAccess: aws_eks.EndpointAccess.PUBLIC,
      authenticationMode: aws_eks.AuthenticationMode.API_AND_CONFIG_MAP,
      bootstrapClusterCreatorAdminPermissions: true
    })

    new aws_eks.CfnAddon(this, "ebs-csi-addon", {
      addonName: "aws-ebs-csi-driver",
      clusterName: cluster.clusterName
    })

    const awsAccountId = process.env.AWS_ACCOUNT_ID;
    cluster.grantAccess("AdminAccess", `arn:aws:iam::${awsAccountId}:root`, [
      AccessPolicy.fromAccessPolicyName("AmazonEKSClusterAdminPolicy", {
        accessScopeType: AccessScopeType.CLUSTER
      })
    ])

    // cluster.addAutoScalingGroupCapacity("nodes", {
    //   instanceType: new aws_ec2.InstanceType("t3.small"),
    //   minCapacity: 1,
    //   maxCapacity: 2,
    //   vpcSubnets: { subnetType: aws_ec2.SubnetType.PUBLIC },
    //   associatePublicIpAddress: true
    // })
    cluster.addNodegroupCapacity("managed-nodes", {
      instanceTypes: [new aws_ec2.InstanceType("t3.small")],
      minSize: 1,
      maxSize: 2,
      subnets: { subnetType: aws_ec2.SubnetType.PUBLIC },
      amiType: aws_eks.NodegroupAmiType.AL2023_X86_64_STANDARD
    });

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
        metadata: { name: "ghcr-auth", namespace: "default" },
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
    podRole.addManagedPolicy(aws_iam.ManagedPolicy.fromAwsManagedPolicyName('AmazonS3FullAccess'))

    // Create the IAM Role for the EBS CSI Driver
    const ebsCsiRole = new aws_iam.Role(this, 'EbsCsiRole', {
      assumedBy: new aws_iam.WebIdentityPrincipal(cluster.openIdConnectProvider.openIdConnectProviderArn, {
        StringEquals: new cdk.CfnJson(this, 'EbsCsiCondition', {
          value: {
            [`${cluster.openIdConnectProvider.openIdConnectProviderIssuer}:sub`]: 'system:serviceaccount:kube-system:ebs-csi-controller-sa',
          },
        }),
      }),
    });

    // Attach the AWS-managed policy for EBS
    ebsCsiRole.addManagedPolicy(aws_iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AmazonEBSCSIDriverPolicy'));

    // Update the Addon to use this Role
    new aws_eks.CfnAddon(this, 'ebs-csi-addon', {
      addonName: 'aws-ebs-csi-driver',
      clusterName: cluster.clusterName,
      serviceAccountRoleArn: ebsCsiRole.roleArn, // Link the role here
    });

    cluster.addManifest('cloudcrush-sa', {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: {
        name: 'cloudcrush-sa',
        namespace: 'default',
        annotations: { 'eks.amazonaws.com/role-arn': podRole.roleArn }
      }
    });

    new aws_eks.AccessEntry(this, 'LocalAdminAccess', {
      cluster: cluster,
      principal: `arn:aws:iam::${awsAccountId}:user/admin_user`, // Replace with your IAM User ARN
      accessPolicies: [
        aws_eks.AccessPolicy.fromAccessPolicyName('AmazonEKSClusterAdminPolicy', {
          accessScopeType: aws_eks.AccessScopeType.CLUSTER,
        }),
      ],
    });
  }
}
