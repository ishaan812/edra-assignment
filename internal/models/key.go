package models

import (
	"fmt"
	cronJobHandler "go-server-template/pkg/cron"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Key struct {
	gorm.Model
	KeyID              string `gorm:"primaryKey;" json:"key_id,omitempty"`
	Busy               bool   `json:"busy,omitempty" gorm:"default:false;index"`
	KeyLifecycleEntry  int    `json:"key_lifecycle_entry,omitempty"`
	UserLifecycleEntry int    `json:"user_lifecycle_entry,omitempty"`
}

func DeleteKey(db *gorm.DB, keyID string) {
	db.Exec("UPDATE keys SET deleted_at = NOW() AND key_lifecyle_entry = 0 WHERE key_id = ?;", keyID)
}

func (k *Key) Init(dbconn *gorm.DB) (cron.EntryID, error) {
	c := cronJobHandler.GetCronJobHandler()
	var err error
	var cronEntryID cron.EntryID
	k.Busy = false
	k.KeyID = uuid.New().String()
	expiryTime := time.Now().Truncate(time.Minute).Add(5 * time.Minute)
	cronExpr := fmt.Sprintf("%d %d %d %d %d", expiryTime.Minute(), expiryTime.Hour(), expiryTime.Day(), expiryTime.Month(), expiryTime.Weekday())
	cronEntryID, err = c.AddFunc(cronExpr, func() {
		fmt.Println("Key expired:", k.KeyID)
		DeleteKey(dbconn, k.KeyID)
		fmt.Println("Hwo herer", cronEntryID)
		c.Remove(cronEntryID)
	})
	if err != nil {
		fmt.Println("Error adding cron job:", err)
		return 0, err
	}
	k.KeyLifecycleEntry = int(cronEntryID)
	return cronEntryID, nil
}
