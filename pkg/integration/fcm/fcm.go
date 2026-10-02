package fcm

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type PushSender struct {
	client *messaging.Client
}

func findCredentialsFile(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func NewPushSender(credentialsSource string) *PushSender {
	var opt option.ClientOption

	// 1. If environment variable FCM_SERVICE_ACCOUNT_JSON is set, use raw JSON credentials
	fcmJson := os.Getenv("FCM_SERVICE_ACCOUNT_JSON")
	if fcmJson != "" {
		opt = option.WithCredentialsJSON([]byte(fcmJson))
	} else if strings.HasPrefix(strings.TrimSpace(credentialsSource), "{") {
		// 2. Raw JSON string passed directly
		opt = option.WithCredentialsJSON([]byte(credentialsSource))
	} else {
		// 3. Search candidate credential file paths
		filePath := findCredentialsFile(
			credentialsSource,
			os.Getenv("FCM_CREDENTIALS_FILE"),
			"configs/secrets/yolyolakay-3310f-firebase-adminsdk-fbsvc-b4053bed72.json",
			"../configs/secrets/yolyolakay-3310f-firebase-adminsdk-fbsvc-b4053bed72.json",
			"serviceAccountKey.json",
			"/root/serviceAccountKey.json",
		)
		if filePath != "" {
			slog.Info("Found Firebase credentials file", "path", filePath)
			opt = option.WithCredentialsFile(filePath)
		} else {
			slog.Warn("Firebase credentials file not found on disk, running in mock mode")
			return nil
		}
	}

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		slog.Error("Failed to initialize Firebase App", "error", err)
		return nil
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		slog.Error("Failed to initialize Firebase Messaging client", "error", err)
		return nil
	}

	slog.Info("Firebase Messaging Client initialized successfully")
	return &PushSender{client: client}
}

func (f *PushSender) SendPush(ctx context.Context, token string, title, body string, data map[string]string) error {
	if f == nil || f.client == nil {
		slog.Warn("[FCM MOCK] PushSender client not initialized, skipping push send")
		return nil
	}

	// Ensure all data payload values are strings (required by FCM)
	cleanData := make(map[string]string)
	for k, v := range data {
		cleanData[k] = v
	}
	if title != "" && cleanData["title"] == "" {
		cleanData["title"] = title
	}
	if body != "" && cleanData["body"] == "" {
		cleanData["body"] = body
	}

	isChat := cleanData["type"] == "new_message" || cleanData["type"] == "chat_message"
	androidChannelID := "yolyolakay_default_v3"
	iosSound := "yol_ovoz.wav"
	if isChat {
		androidChannelID = "yolyolakay_chat_v3"
		iosSound = "default"
	}

	badge := 1
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: cleanData,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID:    androidChannelID,
				Sound:        "default",
				DefaultSound: true,
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority":  "10",
				"apns-push-type": "alert",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: iosSound,
					Badge: &badge,
				},
			},
		},
	}

	_, err := f.client.Send(ctx, msg)
	if err != nil {
		if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
			slog.Warn("FCM token is invalid or unregistered", "token", token, "error", err)
		}
	}
	return err
}
