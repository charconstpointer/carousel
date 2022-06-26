package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

type Handler struct {
	c *slack.Client
}

func NewHandler(apiKey string) *Handler {
	c := slack.New(apiKey)
	return &Handler{c: c}
}

type ReadinessRequest struct {
	UserID     string
	CallbackID string
}

func (s *Handler) RequestUserReadiness(ctx context.Context, userID string) (*ReadinessRequest, error) {
	conversation, _, _, err := s.c.OpenConversation(&slack.OpenConversationParameters{
		Users: []string{userID},
	})
	if err != nil {
		return nil, fmt.Errorf("could open private conversation with user %s : %w", userID, err)
	}
	callback := fmt.Sprintf("readiness_%s", time.Now().UTC())
	attr := slack.Attachment{
		Title:      "Daily reminder",
		Text:       "Are you able to direct today's daily meeting? Respond below",
		CallbackID: callback,
		Actions: []slack.AttachmentAction{
			{
				Name:  "OK button",
				Text:  "Yes",
				Type:  "button",
				Value: "OK",
			},
			{
				Name:  "Nope button",
				Text:  "Nope",
				Type:  "button",
				Value: "Nope",
			},
		},
	}
	_, _, err = s.c.PostMessageContext(ctx, conversation.ID, slack.MsgOptionAttachments(attr))
	if err != nil {
		return nil, fmt.Errorf("could not post a message to a private conversation %s : %w", conversation.ID, err)
	}
	return &ReadinessRequest{
		UserID:     userID,
		CallbackID: callback,
	}, nil
}

type MessageResponse struct {
	Enterprise interface{} `json:"enterprise"`
	Team       struct {
		Id     string `json:"id"`
		Domain string `json:"domain"`
	} `json:"team"`
	Channel struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	User struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	AttachmentId string `json:"attachment_id"`
	CallbackId   string `json:"callback_id"`
	ActionTs     string `json:"action_ts"`
	MessageTs    string `json:"message_ts"`
	ResponseUrl  string `json:"response_url"`
	Token        string `json:"token"`
	Type         string `json:"type"`
	TriggerId    string `json:"trigger_id"`
	Actions      []struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"actions"`
	OriginalMessage struct {
		BotId       string `json:"bot_id"`
		Type        string `json:"type"`
		Text        string `json:"text"`
		User        string `json:"user"`
		Ts          string `json:"ts"`
		AppId       string `json:"app_id"`
		Team        string `json:"team"`
		Attachments []struct {
			Fallback   string `json:"fallback"`
			Text       string `json:"text"`
			CallbackId string `json:"callback_id"`
			Actions    []struct {
				Id    string `json:"id"`
				Name  string `json:"name"`
				Text  string `json:"text"`
				Type  string `json:"type"`
				Value string `json:"value"`
				Style string `json:"style"`
			} `json:"actions"`
			Id int `json:"id"`
		} `json:"attachments"`
		BotProfile struct {
			Icons struct {
				Image36 string `json:"image_36"`
				Image48 string `json:"image_48"`
				Image72 string `json:"image_72"`
			} `json:"icons"`
			Id      string `json:"id"`
			Name    string `json:"name"`
			AppId   string `json:"app_id"`
			TeamId  string `json:"team_id"`
			Updated int    `json:"updated"`
			Deleted bool   `json:"deleted"`
		} `json:"bot_profile"`
	} `json:"original_message"`
	IsAppUnfurl         bool `json:"is_app_unfurl"`
	IsEnterpriseInstall bool `json:"is_enterprise_install"`
}
