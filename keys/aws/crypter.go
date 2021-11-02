package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

type AWSKMSCrypter struct {
	keyARN string
}

func NewAWSKMSCrypter(key []byte) *AWSKMSCrypter {
	return &AWSKMSCrypter{keyARN: string(key)}
}

func (c *AWSKMSCrypter) Encrypt(message []byte) (encrypted []byte, err error) {
	ctx := context.Background()
	client := createKMSClient(ctx)

	encryptOutput, err := client.Encrypt(ctx, &kms.EncryptInput{
		KeyId:               aws.String(c.keyARN),
		Plaintext:           message,
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
	})
	if err != nil {
		return encrypted, err
	}

	encrypted = encryptOutput.CiphertextBlob

	return encrypted, err
}

func (c *AWSKMSCrypter) Decrypt(encrypted []byte) (message []byte, err error) {
	ctx := context.Background()
	client := createKMSClient(ctx)

	decryptOutput, err := client.Decrypt(ctx, &kms.DecryptInput{
		KeyId:               aws.String(c.keyARN),
		CiphertextBlob:      encrypted,
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
	})
	if err != nil {
		return message, err
	}

	message = decryptOutput.Plaintext

	return message, err
}
