package notify

import (
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/httpclient"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
)

type WebHook struct{}

func (webHook *WebHook) Send(msg Message) {
	model := new(models.Setting)
	webHookSetting, err := model.Webhook()
	if err != nil {
		logger.Error("#webHook#从数据库获取webHook配置失败", err)
		return
	}
	if len(webHookSetting.WebhookUrls) == 0 {
		logger.Error("#webHook#webhook地址列表为空")
		return
	}
	logger.Debugf("%+v", webHookSetting)
	msg["content"] = renderWebhookTemplate(webHookSetting.Template, msg)

	// 获取任务配置的接收者ID列表
	activeUrls := webHook.getActiveWebhookUrls(webHookSetting, msg)

	// 向所有激活的webhook地址发送
	for _, webhookUrl := range activeUrls {
		go webHook.send(msg, webhookUrl.Url)
	}
}

func (webHook *WebHook) getActiveWebhookUrls(webHookSetting models.WebHook, msg Message) []models.WebhookUrl {
	taskReceiverIds := strings.Split(msg["task_receiver_id"].(string), ",")
	urls := []models.WebhookUrl{}
	for _, v := range webHookSetting.WebhookUrls {
		if utils.InStringSlice(taskReceiverIds, strconv.Itoa(v.Id)) {
			urls = append(urls, v)
		}
	}
	return urls
}

func (webHook *WebHook) send(msg Message, url string) {
	content := msg["content"].(string)
	timeout := 30
	maxTimes := 3
	i := 0
	for i < maxTimes {
		resp := httpclient.PostJson(url, content, timeout)
		if resp.StatusCode == 200 {
			break
		}
		i += 1
		time.Sleep(2 * time.Second)
		if i < maxTimes {
			logger.Errorf("webHook#发送消息失败#%s#消息内容-%s", resp.Body, msg["content"])
		}
	}
}

func renderWebhookTemplate(tmpl string, msg Message) string {
	msg["name"] = utils.EscapeJson(msg["name"].(string))
	msg["output"] = utils.EscapeJson(msg["output"].(string))
	if statusText, ok := msg["status_text"].(string); ok {
		msg["status_text"] = utils.EscapeJson(statusText)
	}
	return html.UnescapeString(parseNotifyTemplate(tmpl, msg))
}
