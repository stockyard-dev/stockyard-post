package server

import "github.com/stockyard-dev/stockyard-post/internal/license"

type Limits struct {
	MaxForms            int  // 0 = unlimited
	MaxSubmissionsMonth int  // per form
	RetentionDays       int
	WebhookNotify       bool // Pro
	FileUploads         bool // Pro
	CustomThankYou      bool // Pro
	ExportCSV           bool // Pro
}

var freeLimits = Limits{
	MaxForms:            3,
	MaxSubmissionsMonth: 100,
	RetentionDays:       7,
	WebhookNotify:       false,
	FileUploads:         false,
	CustomThankYou:      false,
	ExportCSV:           false,
}

var proLimits = Limits{
	MaxForms:            0,
	MaxSubmissionsMonth: 0,
	RetentionDays:       90,
	WebhookNotify:       true,
	FileUploads:         true,
	CustomThankYou:      true,
	ExportCSV:           true,
}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() {
		return proLimits
	}
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}
