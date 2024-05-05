package controllers

import (
	"fmt"
	models "go-server-template/internal/models"
	response "go-server-template/internal/responses"
	cronJobHandler "go-server-template/pkg/cron"
	"go-server-template/pkg/db"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func unblockKey(db *gorm.DB, keyID string) {
	db.Exec("UPDATE keys SET busy = false WHERE key_id = ?;", keyID)
}

func GenerateKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	c := cronJobHandler.GetCronJobHandler()
	key.Init(dbconn)
	err := dbconn.Create(&key).Error
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	entries := c.Entries()
	fmt.Println("All entries in cron scheduler:")
	for _, entry := range entries {
		fmt.Printf("Entry ID: %d, Next scheduled time: %s, Job function: %v\n", entry.ID, entry.Next, entry.Job)
	}
	response.JSON(w, 200, map[string]interface{}{"x-api-key": key.KeyID, "key-expiry-time": c.Entry(cron.EntryID(key.KeyLifecycleEntry)).Next}, nil)
}

func GetAvailableKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	var cronEntryID cron.EntryID
	var expiryTime time.Time
	dbconn := db.GetDB()
	c := cronJobHandler.GetCronJobHandler()
	err := dbconn.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("busy = false").First(&key).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("no key available")
			}
			return err
		}
		expiryTime = time.Now().Truncate(time.Minute).Add(time.Minute)
		cronExpr := fmt.Sprintf("%d %d %d %d %d", expiryTime.Minute(), expiryTime.Hour(), expiryTime.Day(), expiryTime.Month(), expiryTime.Weekday())
		cronEntryID, err = c.AddFunc(cronExpr, func() {
			fmt.Println("User Access expired:", key.KeyID)
			unblockKey(dbconn, key.KeyID)
		})
		fmt.Println(cronEntryID)
		if err != nil {
			return err
		}
		key.Busy = true
		key.UserLifecycleEntry = int(cronEntryID)
		err = tx.Save(&key).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.Error(w, 400, fmt.Errorf("Error in getting key: ", err))
		return
	}
	response.JSON(w, 200, map[string]interface{}{"x-api-key": key.KeyID, "access-expiry-time": expiryTime, "key-expiry-time": c.Entry(cron.EntryID(key.KeyLifecycleEntry)).Next}, nil)
}

func GetKeyInfoHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	c := cronJobHandler.GetCronJobHandler()
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		err = tx.First(&key, "key_id = ?", id).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("key does not exist or has been deleted")
			}
			return err
		}
		return nil
	})
	if err != nil {
		w.Header().Add("error", err.Error())
		response.Error(w, 400, nil)
		return
	}
	w.Header().Add("x-api-key", key.KeyID)
	w.Header().Add("key-expiry-time", c.Entry(cron.EntryID(key.KeyLifecycleEntry)).Next.String())
	w.Header().Add("access-expiry-time", c.Entry(cron.EntryID(key.UserLifecycleEntry)).Next.String())
	w.Header().Add("busy", strconv.FormatBool(key.Busy))
	w.WriteHeader(200)
	return
}

func DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	c := cronJobHandler.GetCronJobHandler()
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		err = dbconn.Delete(&key, "key_id = ?", id).Error
		if err != nil {
			return err
		}
		c.Remove(cron.EntryID(key.KeyLifecycleEntry))
		c.Remove(cron.EntryID(key.UserLifecycleEntry))
		return nil
	})
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	response.JSON(w, 200, "Key deleted", nil)
}

func UnblockKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	c := cronJobHandler.GetCronJobHandler()
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		err = dbconn.First(&key, "key_id = ?", id).Error
		if err != nil {
			return err
		}
		key.Busy = false
		c.Remove(cron.EntryID(key.UserLifecycleEntry))
		err = dbconn.Save(&key).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	response.JSON(w, 200, "Key unblocked", nil)
}

func KeepAliveHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	c := cronJobHandler.GetCronJobHandler()
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		err = dbconn.First(&key, "key_id = ?", id).Error
		if err != nil {
			return err
		}
		if key.KeyID == "" {
			return err
		}
		c.Remove(cron.EntryID(key.KeyLifecycleEntry))
		expiryTime := time.Now().Add(5 * time.Minute)
		cronExpr := fmt.Sprintf("%d %d %d %d %d", expiryTime.Minute(), expiryTime.Hour(), expiryTime.Day(), expiryTime.Month(), expiryTime.Weekday())
		cronEntryID, err := c.AddFunc(cronExpr, func() {
			fmt.Println("Key expired:", key.KeyID)
			models.DeleteKey(dbconn, key.KeyID)
		})
		if err != nil {
			return err
		}
		key.KeyLifecycleEntry = int(cronEntryID)
		err = dbconn.Save(&key).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	response.JSON(w, 200, map[string]interface{}{"x-api-key": key.KeyID, "access-expiry-time": c.Entry(cron.EntryID(key.UserLifecycleEntry)).Next, "key-expiry-time": c.Entry(cron.EntryID(key.KeyLifecycleEntry)).Next}, nil)
}
