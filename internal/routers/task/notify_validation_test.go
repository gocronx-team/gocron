package task

import (
	"testing"

	"github.com/gocronx-team/gocron/internal/models"
)

func TestValidateNotifyKeywordRegex(t *testing.T) {
	tests := []struct {
		name    string
		task    models.Task
		wantKey string
	}{
		{
			"keyword condition off skips validation",
			models.Task{NotifyStatus: 3, NotifyKeywordRegex: 1, NotifyKeyword: "[bad"},
			"",
		},
		{
			"substring mode skips validation",
			models.Task{NotifyStatus: 4, NotifyKeywordRegex: 0, NotifyKeyword: "[bad", NotifyKeywordExclude: "[bad"},
			"",
		},
		{
			"valid keyword and exclude regex",
			models.Task{NotifyStatus: 4, NotifyKeywordRegex: 1, NotifyKeyword: "ERR[0-9]+", NotifyKeywordExclude: "ERR0(ignored)?"},
			"",
		},
		{
			"invalid keyword regex",
			models.Task{NotifyStatus: 4, NotifyKeywordRegex: 1, NotifyKeyword: "ERR[0-9"},
			"notify_keyword_regex_invalid",
		},
		{
			"invalid exclude regex",
			models.Task{NotifyStatus: 4, NotifyKeywordRegex: 1, NotifyKeyword: "ERROR", NotifyKeywordExclude: "ERROR: [unclosed"},
			"notify_keyword_exclude_regex_invalid",
		},
		{
			"empty exclude passes",
			models.Task{NotifyStatus: 4, NotifyKeywordRegex: 1, NotifyKeyword: "ERROR"},
			"",
		},
		{
			"invalid keyword reported before exclude",
			models.Task{NotifyStatus: 4, NotifyKeywordRegex: 1, NotifyKeyword: "[bad", NotifyKeywordExclude: "[bad"},
			"notify_keyword_regex_invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateNotifyKeywordRegex(tt.task); got != tt.wantKey {
				t.Errorf("validateNotifyKeywordRegex() = %q, want %q", got, tt.wantKey)
			}
		})
	}
}
