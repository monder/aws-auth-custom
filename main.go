package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/kubernetes-sigs/aws-iam-authenticator/pkg/token"
)

func main() {
	var roleARN, roleSessionName, clusterID string
	flag.StringVar(&roleARN, "r", "", "Role to assume")
	flag.StringVar(&roleSessionName, "s", "", "session name. defaults to instance private dns name")
	flag.StringVar(&clusterID, "i", "", "cluster id")
	flag.Parse()

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get token: %v\n", err)
		os.Exit(1)
	}

	if roleSessionName == "" {
		roleSessionName, err = ec2metadata.New(sess).GetMetadata("local-hostname")
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get token: %v\n", err)
			os.Exit(1)
		}
	}
	if roleARN == "" || clusterID == "" {
		fmt.Fprintf(os.Stderr, "you must specify at least role and cluster id")
		os.Exit(1)
	}

	gen, err := token.NewGenerator(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get token: %v\n", err)
		os.Exit(1)
	}

	sessionSetter := func(provider *stscreds.AssumeRoleProvider) {
		provider.RoleSessionName = roleSessionName
	}

	creds := stscreds.NewCredentials(sess, roleARN, sessionSetter)

	stsAPI := sts.New(sess, &aws.Config{Credentials: creds})
	token, err := gen.GetWithSTS(clusterID, stsAPI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get token: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(gen.FormatJSON(token))
}
