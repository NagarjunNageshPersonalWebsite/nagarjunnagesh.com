package main

import (
	"fmt"
	"website-infra/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type WebsiteInfraStackProps struct {
	awscdk.StackProps
	DomainName     string
	Environment    string
	Project        string
	CertificateArn string
	HostedZoneId   string
	HomePage       string
}

func NewWebsiteInfraStack(scope constructs.Construct, id string, props *WebsiteInfraStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Creation of the stack goes here

	domainName, bucket := createS3Bucket(props, stack)
	// CloudFront Origin Access Identity
	cloudfrontOAI, cloudfrontOAIPrincipal := createCloudfrontOAI(stack, props)

	bucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:    jsii.Strings("s3:GetObject"),
		Effect:     awsiam.Effect_ALLOW,
		Principals: &[]awsiam.IPrincipal{awsiam.NewCanonicalUserPrincipal(cloudfrontOAIPrincipal)},
		Resources: &[]*string{
			bucket.ArnForObjects(jsii.String("*")),
		},
	}))

	bucket.AddLifecycleRule(&s3.LifecycleRule{
		Id:                                  jsii.String("delete-old-versioned-objects"),
		NoncurrentVersionExpiration:         awscdk.Duration_Days(jsii.Number(3)),
		Enabled:                             jsii.Bool(true),
		ExpiredObjectDeleteMarker:           jsii.Bool(true),                      // Delete expired object delete markers
		AbortIncompleteMultipartUploadAfter: awscdk.Duration_Days(jsii.Number(3)), // Delete incomplete multipart uploads after 3 days
	})

	myCertificate := awscertificatemanager.Certificate_FromCertificateArn(stack, jsii.String(fmt.Sprintf("%s-HTTPSCertificate", props.DomainName)), jsii.String(props.CertificateArn))
	myResponseHeadersPolicy := createResponseHeaderPolicy(stack, props)
	distribution := createCloudfrontDistribution(stack, props, bucket, cloudfrontOAI, myResponseHeadersPolicy, myCertificate)
	createHostedZones(stack, props, distribution)
	cdkOutput(stack, distribution, domainName, bucket)

	return stack
}

func createCloudfrontOAI(stack awscdk.Stack, props *WebsiteInfraStackProps) (awscloudfront.OriginAccessIdentity, *string) {
	cloudfrontOAI := awscloudfront.NewOriginAccessIdentity(stack, jsii.String(fmt.Sprintf("%s-cloudfront-OAI", props.DomainName)), &awscloudfront.OriginAccessIdentityProps{
		Comment: jsii.String(fmt.Sprintf("[%s] %s Static Resources OAI", props.Environment, props.Project)),
	})

	cloudfrontOAIPrincipal := cloudfrontOAI.CloudFrontOriginAccessIdentityS3CanonicalUserId()
	return cloudfrontOAI, cloudfrontOAIPrincipal
}

func createS3Bucket(props *WebsiteInfraStackProps, stack awscdk.Stack) (string, s3.Bucket) {
	domainName := props.DomainName
	var bucket s3.Bucket
	if domainName == NAGARJUNNAGESH_COM_DOMAIN_NAME {
		bucket = s3.NewBucket(stack, jsii.String(domainName+"Bucket"), &s3.BucketProps{
			BucketName:        jsii.String(domainName),
			PublicReadAccess:  jsii.Bool(false),
			BlockPublicAccess: s3.BlockPublicAccess_BLOCK_ALL(),
			Versioned:         jsii.Bool(true),
			ObjectOwnership:   s3.ObjectOwnership_OBJECT_WRITER,
		})
	} else {
		bucket = s3.NewBucket(stack, jsii.String(domainName+"Bucket"), &s3.BucketProps{
			BucketName:        jsii.String(domainName),
			PublicReadAccess:  jsii.Bool(false),
			BlockPublicAccess: s3.BlockPublicAccess_BLOCK_ALL(),
			Versioned:         jsii.Bool(true),
			ObjectOwnership:   s3.ObjectOwnership_OBJECT_WRITER,
			WebsiteRedirect: &s3.RedirectTarget{
				HostName: jsii.String(NAGARJUNNAGESH_COM_DOMAIN_NAME),
			},
		})
	}
	return domainName, bucket
}

func createResponseHeaderPolicy(stack awscdk.Stack, props *WebsiteInfraStackProps) awscloudfront.ResponseHeadersPolicy {
	cspPolicy := `default-src 'self' https://www.google.com/recaptcha/api2/; script-src 'self' 'unsafe-inline' https://nagarjunnagesh.com/  https://cdn.jsdelivr.net/ https://www.google.com/recaptcha/ https://cdn.tailwindcss.com/ https://cdnjs.cloudflare.com/ https://www.gstatic.com/recaptcha/; object-src 'none'; font-src data: https://use.fontawesome.com/ https://fonts.gstatic.com/ https://cdnjs.cloudflare.com/ 'self'; style-src-elem https://cdn.jsdelivr.net/ https://use.fontawesome.com/ https://cdnjs.cloudflare.com/ https://fonts.googleapis.com/ 'self' 'unsafe-inline'; style-src-attr 'unsafe-hashes' 'sha256-X+zrZv/IbzjZUnhsbWlsecLbwjndTpG0ZynXOif7V+k=' 'sha256-a4ayc/80/OGda4BO/1o/V0etpOqiLx1JwB5S3beHW0s='`

	rspName := "NAGARJUNNAGESHResponseHeadersPolicy"
	if props.DomainName != NAGARJUNNAGESH_COM_DOMAIN_NAME {
		rspName = "WWWNAGARJUNNAGESHResponseHeaderPolicy"
	}

	myResponseHeadersPolicy := awscloudfront.NewResponseHeadersPolicy(stack, jsii.String(rspName), &awscloudfront.ResponseHeadersPolicyProps{
		ResponseHeadersPolicyName: jsii.String(rspName),
		Comment:                   jsii.String(fmt.Sprintf("%s ResponseHeadersPolicy", props.DomainName)),
		SecurityHeadersBehavior: &awscloudfront.ResponseSecurityHeadersBehavior{
			ContentSecurityPolicy: &awscloudfront.ResponseHeadersContentSecurityPolicy{
				ContentSecurityPolicy: jsii.String(cspPolicy),
				Override:              jsii.Bool(true),
			},
			ContentTypeOptions: &awscloudfront.ResponseHeadersContentTypeOptions{
				Override: jsii.Bool(true),
			},
			FrameOptions: &awscloudfront.ResponseHeadersFrameOptions{
				FrameOption: awscloudfront.HeadersFrameOption_SAMEORIGIN,
				Override:    jsii.Bool(true),
			},
			StrictTransportSecurity: &awscloudfront.ResponseHeadersStrictTransportSecurity{
				AccessControlMaxAge: awscdk.Duration_Seconds(jsii.Number(600)),
				IncludeSubdomains:   jsii.Bool(true),
				Override:            jsii.Bool(true),
			},
			ReferrerPolicy: &awscloudfront.ResponseHeadersReferrerPolicy{
				ReferrerPolicy: awscloudfront.HeadersReferrerPolicy_NO_REFERRER,
				Override:       jsii.Bool(true),
			},
			XssProtection: &awscloudfront.ResponseHeadersXSSProtection{
				Override:   jsii.Bool(true),
				Protection: jsii.Bool(true),
				ModeBlock:  jsii.Bool(true),
			},
		},
		RemoveHeaders: &[]*string{
			jsii.String("Server"),
		},
		CorsBehavior: &awscloudfront.ResponseHeadersCorsBehavior{
			AccessControlAllowCredentials: jsii.Bool(false),
			AccessControlAllowHeaders:     jsii.Strings("*"),
			AccessControlAllowMethods:     jsii.Strings("GET", "POST", "OPTIONS"),
			AccessControlAllowOrigins:     jsii.Strings(props.DomainName),
			AccessControlExposeHeaders:    jsii.Strings("*"),
			OriginOverride:                jsii.Bool(true),
			AccessControlMaxAge:           awscdk.Duration_Seconds(jsii.Number(600)),
		},
		CustomHeadersBehavior: &awscloudfront.ResponseCustomHeadersBehavior{
			CustomHeaders: &[]*awscloudfront.ResponseCustomHeader{
				{
					Header:   jsii.String("Cache-Control"),
					Value:    jsii.String("max-age=3600, must-revalidate"),
					Override: jsii.Bool(true),
				},
			},
		},
	})
	return myResponseHeadersPolicy
}

func createCloudfrontDistribution(stack awscdk.Stack, props *WebsiteInfraStackProps, bucket s3.Bucket, cloudfrontOAI awscloudfront.OriginAccessIdentity, myResponseHeadersPolicy awscloudfront.ResponseHeadersPolicy, myCertificate awscertificatemanager.ICertificate) awscloudfront.Distribution {
	distribution := awscloudfront.NewDistribution(stack, jsii.String(fmt.Sprintf("%sCDN", props.DomainName)), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:                awscloudfrontorigins.NewS3Origin(bucket, &awscloudfrontorigins.S3OriginProps{OriginAccessIdentity: cloudfrontOAI}),
			Compress:              jsii.Bool(true),
			CachePolicy:           awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
			AllowedMethods:        awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS(),
			ViewerProtocolPolicy:  awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			ResponseHeadersPolicy: myResponseHeadersPolicy,
		},
		ErrorResponses: &[]*awscloudfront.ErrorResponse{
			{
				HttpStatus:         jsii.Number(403),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String(props.HomePage),
			},
			{
				HttpStatus:         jsii.Number(404),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String(props.HomePage),
			},
		},
		MinimumProtocolVersion: awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021,
		DefaultRootObject:      jsii.String(props.HomePage),
		HttpVersion:            awscloudfront.HttpVersion_HTTP2_AND_3,
		Certificate:            myCertificate,
		DomainNames:            jsii.Strings(props.DomainName),
		Comment:                jsii.String(fmt.Sprintf("[%s] %s Static Resources", props.Environment, props.Project)),
	})
	return distribution
}

func createHostedZones(stack awscdk.Stack, props *WebsiteInfraStackProps, distribution awscloudfront.Distribution) {
	hostedZone := awsroute53.HostedZone_FromHostedZoneAttributes(stack, jsii.String("HostedZone"), &awsroute53.HostedZoneAttributes{
		HostedZoneId: jsii.String(props.HostedZoneId),
		ZoneName:     jsii.String(props.DomainName),
	})

	awsroute53.NewARecord(stack, jsii.String(fmt.Sprintf("%sApiARecord", props.DomainName)), &awsroute53.ARecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(props.DomainName),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(distribution)),
	})

	awsroute53.NewAaaaRecord(stack, jsii.String(fmt.Sprintf("%sApiAaaaRecord", props.DomainName)), &awsroute53.AaaaRecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(props.DomainName),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(distribution)),
	})
}

func cdkOutput(stack awscdk.Stack, distribution awscloudfront.Distribution, domainName string, bucket s3.Bucket) {
	awscdk.NewCfnOutput(stack, jsii.String("DistributionDomainName"), &awscdk.CfnOutputProps{
		Value: distribution.DomainName(),
	})

	awscdk.NewCfnOutput(stack, jsii.String(domainName+"BucketName"), &awscdk.CfnOutputProps{
		Value: bucket.BucketName(),
	})
}

type UsEast1StackProps struct {
	awscdk.StackProps
	HostedZoneID string
	DomainName   string
	Environment  string
}

func NewUsEast1Stack(scope constructs.Construct, id string, props *UsEast1StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	// Create a hosted zone object using the hosted zone ID
	hostedZone := awsroute53.HostedZone_FromHostedZoneAttributes(stack, jsii.String("HostedZone"), &awsroute53.HostedZoneAttributes{
		HostedZoneId: jsii.String(props.HostedZoneID),
		ZoneName:     jsii.String(props.DomainName),
	})

	// Define the ACM certificate
	certificate := awscertificatemanager.NewCertificate(stack, jsii.String("Certificate"), &awscertificatemanager.CertificateProps{
		DomainName:              jsii.String(props.DomainName),
		SubjectAlternativeNames: jsii.Strings("*."+props.DomainName, props.DomainName),
		CertificateName:         jsii.String(props.DomainName),
		Validation:              awscertificatemanager.CertificateValidation_FromDns(hostedZone),
	})

	// Define the outputs
	awscdk.NewCfnOutput(stack, jsii.String(props.Environment+"CertificateArn"), &awscdk.CfnOutputProps{
		Description: jsii.String("ARN of the API Gateway Certificate"),
		Value:       certificate.CertificateArn(),
		ExportName:  jsii.String(props.Environment + "CertificateArn"),
	})

	return stack
}

var (
	NAGARJUNNAGESH_COM_DOMAIN_NAME = "nagarjunnagesh.com"
	ENVIRONMENT         = "prod"
	PROJECT             = "NAGARJUNNAGESH"
	HOSTED_ZONE_ID      = "Z03896663UJBQA3QD1F4X"
	HOME_PAGE           = "/index.html"
	CERTIFICATE_ARN     = config.CertificateARN
	WWW_HOME_PAGE       = "/"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Stack for ACM certificate in us-east-1
	NewUsEast1Stack(app, "UsEast1Stack", &UsEast1StackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Region: jsii.String("us-east-1"),
			},
		},
		HOSTED_ZONE_ID,
		NAGARJUNNAGESH_COM_DOMAIN_NAME,
		ENVIRONMENT,
	})

	// Create nagarjunnagesh.com infrastructure
	NewWebsiteInfraStack(app, "WebsiteInfraStack", &WebsiteInfraStackProps{
		awscdk.StackProps{
			Env: env(),
		},
		NAGARJUNNAGESH_COM_DOMAIN_NAME,
		ENVIRONMENT,
		PROJECT,
		CERTIFICATE_ARN,
		HOSTED_ZONE_ID,
		HOME_PAGE,
	})

	// Create www.nagarjunnagesh.com infrastructure
	NewWebsiteInfraStack(app, "WWWWebsiteInfraStack", &WebsiteInfraStackProps{
		awscdk.StackProps{
			Env: env(),
		},
		fmt.Sprintf("www.%s", NAGARJUNNAGESH_COM_DOMAIN_NAME),
		ENVIRONMENT,
		PROJECT,
		CERTIFICATE_ARN,
		HOSTED_ZONE_ID,
		WWW_HOME_PAGE,
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	// return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(config.AWSAccount),
		Region:  jsii.String(config.AWSRegion),
	}

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
