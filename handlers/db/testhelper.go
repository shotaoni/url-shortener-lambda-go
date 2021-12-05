package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

func TestNew() DB {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String(Region),
		Endpoint: aws.String("http://localhost:8000")}),
	)

	return DB{Instance: dynamodb.New(sess)}
}

func (d DB) CreateLinkTable() error {
	cti := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("shorten_resource"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("shorten_resource"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(LinkTableName),
	}

	_, err := d.Instance.CreateTable(cti)
	if err != nil {
		return err
	}

	desti := &dynamodb.DescribeTableInput{
		TableName: aws.String(LinkTableName),
	}
	if err := d.Instance.WaitUntilTableExists(desti); err != nil {
		return err
	}
	return nil
}

func (d DB) DeleteLinkTable() error {
	delti := &dynamodb.DeleteTableInput{
		TableName: aws.String(LinkTableName),
	}
	_, err := d.Instance.DeleteTable(delti)
	if err != nil {
		return err
	}

	desti := &dynamodb.DescribeTableInput{
		TableName: aws.String(LinkTableName),
	}
	if err := d.Instance.WaitUntilTableNotExists(desti); err != nil {
		return err
	}

	return nil
}

func (d DB) LinkTableExists() (bool, error) {
	input := &dynamodb.ListTablesInput{}
	output, err := d.Instance.ListTables(input)
	if err != nil {
		return false, errors.Wrap(err, "failed to list tables")
	}
	if contains(output.TableNames, LinkTableName) {
		return true, nil
	}
	return false, nil
}

func contains(s []*string, e string) bool {
	for _, a := range s {
		if a == nil {
			continue
		}
		if *a == e {
			return true
		}
	}
	return false
}
