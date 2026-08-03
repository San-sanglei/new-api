package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"
)

// SES 发信超时时间，避免 Render 等环境下接口挂起无响应
const sesSendTimeout = 10 * time.Second

var (
	sesClientOnce sync.Once
	sesClient     *sesv2.Client
	sesClientErr  error
)

// getSESClient 懒加载 SESv2 客户端（仅初始化一次）。
// 凭证从环境变量读取：
//   - AWS_REGION
//   - AWS_ACCESS_KEY_ID
//   - AWS_SECRET_ACCESS_KEY
//
// 任一缺失则返回错误，客户端不初始化。
func getSESClient() (*sesv2.Client, error) {
	sesClientOnce.Do(func() {
		region := os.Getenv("AWS_REGION")
		if region == "" {
			sesClientErr = errors.New("AWS_REGION environment variable is not set")
			return
		}
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey == "" || secretKey == "" {
			sesClientErr = errors.New("AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY environment variable is not set")
			return
		}
		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			sesClientErr = fmt.Errorf("load AWS config failed: %w", err)
			return
		}
		sesClient = sesv2.NewFromConfig(cfg)
	})
	return sesClient, sesClientErr
}

// SendEmail 通过 AWS SESv2 HTTPS API 发送邮件。
//
// 设计说明：
//   - 原实现基于 net/smtp，Render 等平台封锁出站 SMTP 端口（25/465/587）会导致接口挂死。
//   - 改用 SESv2 HTTPS API，规避 SMTP 端口限制。
//   - 发件人地址从 SES_FROM_EMAIL 环境变量读取，显示名使用 SystemName。
//   - 保持原签名 (subject, receiver, content string) error，所有调用方无需修改。
//   - 10 秒超时控制，避免接口挂起。
//   - receiver 支持分号分隔多个收件人（与原实现兼容）。
func SendEmail(subject string, receiver string, content string) error {
	client, err := getSESClient()
	if err != nil {
		SysError(fmt.Sprintf("SES client init failed: %v", err))
		return err
	}

	fromEmail := os.Getenv("SES_FROM_EMAIL")
	if fromEmail == "" {
		return errors.New("SES_FROM_EMAIL environment variable is not set")
	}
	// 发件人格式：Display Name <email>
	fromAddress := fmt.Sprintf("%s <%s>", SystemName, fromEmail)

	// 兼容原实现：receiver 用分号分隔多个收件人
	toAddresses := strings.Split(receiver, ";")
	for i := range toAddresses {
		toAddresses[i] = strings.TrimSpace(toAddresses[i])
	}

	ctx, cancel := context.WithTimeout(context.Background(), sesSendTimeout)
	defer cancel()

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(fromAddress),
		Destination: &types.Destination{
			ToAddresses: toAddresses,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data:    aws.String(content),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	}

	_, err = client.SendEmail(ctx, input)
	if err != nil {
		// 提取 AWS 返回的错误码和消息，便于排查
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			SysError(fmt.Sprintf("SES SendEmail failed (code=%s, msg=%s): %v",
				apiErr.ErrorCode(), apiErr.ErrorMessage(), err))
		} else {
			SysError(fmt.Sprintf("SES SendEmail failed: %v", err))
		}
		return fmt.Errorf("send email via SES failed: %w", err)
	}
	return nil
}
