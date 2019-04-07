package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	ec2Instances := []string{}
	ec2InstancesASG := []string{}

	autoscalingGroupName := flag.String("asg", "", "name of autoscaling group")
	flag.Parse()

	if *autoscalingGroupName == "" {
		fmt.Println("Autoscaling Group name required, please set --asg=your-asg-name")
		return
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := ec2.New(sess)

	// Get all EC2 instances matching the autoscaling:GroupName tag
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:aws:autoscaling:groupName"),
				Values: []*string{aws.String(*autoscalingGroupName)},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running"), aws.String("pending")},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Pull out the instanceIds
	for res := range resp.Reservations {
		for i := 0; i < len(resp.Reservations[res].Instances); i++ {
			for _, inst := range resp.Reservations[res].Instances {
				ec2Instances = append(ec2Instances, *inst.InstanceId)
			}
		}
	}

	fmt.Println("The following EC2 Instances are tagged: ", ec2Instances)

	// Get Instances from ASG
	asgsvc := autoscaling.New(sess)
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(*autoscalingGroupName),
		},
	}

	result, err := asgsvc.DescribeAutoScalingGroups(input)
	if err != nil {
		log.Fatal(err.Error())
	}

	for i := 0; i < len(result.AutoScalingGroups[0].Instances); i++ {
		ec2InstancesASG = append(ec2InstancesASG, *result.AutoScalingGroups[0].Instances[i].InstanceId)
	}

	fmt.Println("The following instances are registered to the ASG: ", ec2InstancesASG)
}
