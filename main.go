package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

func printPolicies(iamSvc *iam.IAM, attachedPolicies []*iam.AttachedPolicy) {
	for _, pol := range attachedPolicies {
		getPolOut, _ := iamSvc.GetPolicy(&iam.GetPolicyInput{PolicyArn: pol.PolicyArn})
		polOut, _ := iamSvc.GetPolicyVersion(&iam.GetPolicyVersionInput{
			PolicyArn: getPolOut.Policy.Arn,
			VersionId: getPolOut.Policy.DefaultVersionId,
		})
		data, _ := url.QueryUnescape(*polOut.PolicyVersion.Document)
		fmt.Printf("-------- %s --------\n", *pol.PolicyName)
		fmt.Println(data)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: aws-info <profile-name>")
		return
	}
	prof := os.Args[1]
	cfg := &aws.Config{
		Credentials: credentials.NewSharedCredentials("", prof),
	}
	session := session.Must(session.NewSession(cfg))
	iamSvc := iam.New(session)
	userOutput, err := iamSvc.GetUser(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("==== User =====")
	fmt.Printf("Username: %s\nID: %s\nARN: %s\n", *userOutput.User.UserName, *userOutput.User.UserId, *userOutput.User.Arn)
	groupsOutput, err := iamSvc.ListGroupsForUser(&iam.ListGroupsForUserInput{
		UserName: userOutput.User.UserName,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("==== Groups =====")
	for _, grp := range groupsOutput.Groups {
		fmt.Println("Name: ", *grp.GroupName)
		fmt.Println("Arn: ", *grp.Arn)
		fmt.Println("ID: ", *grp.GroupId)
		attachedGroupPol, _ := iamSvc.ListAttachedGroupPolicies(&iam.ListAttachedGroupPoliciesInput{
			GroupName: grp.GroupName,
		})

		if len(attachedGroupPol.AttachedPolicies) != 0 {
			fmt.Println("==== Attached Group Policies =====")
			printPolicies(iamSvc, attachedGroupPol.AttachedPolicies)
		}
	}
	userPolicies, err := iamSvc.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
		UserName: userOutput.User.UserName,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(userPolicies.AttachedPolicies) != 0 {
		fmt.Println("==== Attached User Policies =====")
		printPolicies(iamSvc, userPolicies.AttachedPolicies)
	}
}
