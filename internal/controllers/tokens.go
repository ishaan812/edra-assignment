package controllers

import (
	"fmt"
	models "go-server-template/internal/models"
	response "go-server-template/internal/responses"
	"go-server-template/pkg/db"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func deleteOldKeys(db *gorm.DB) {
	threshold := time.Now().Add(-5 * time.Minute)
	db.Exec("UPDATE keys SET deleted_at = NOW() WHERE last_accessed < ?;", threshold)
}

func unblockBlockedKeys(db *gorm.DB) {
	threshold := time.Now().Add(-1 * time.Minute)
	db.Exec("UPDATE keys SET busy = false AND user_access_time = NULL WHERE user_access_time < ? AND user_access_time != NULL;", threshold)
}

func GenerateKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	key.Init()
	err := dbconn.Create(&key).Error
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	response.JSON(w, 200, map[string]interface{}{"x-api-key": key.KeyID, "last_accessed": key.LastAccessed, "user_access_time": key.UserAccessTime, "busy": key.Busy}, nil)
}

func GetAvailableKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	err := dbconn.Transaction(func(tx *gorm.DB) error {
		deleteOldKeys(tx)
		unblockBlockedKeys(tx)
		err := tx.Where("busy = false").Find(&key).Error
		if err != nil {
			return err
		}
		if key.KeyID == "" {
			return err
		}
		key.Busy = true
		now := time.Now()
		key.UserAccessTime = &now
		err = tx.Save(&key).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.Error(w, 400, fmt.Errorf("Error in getting key: ", err))
	}
	response.JSON(w, 200, map[string]interface{}{"x-api-key": key.KeyID}, nil)
}

func GetKeyInfoHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		deleteOldKeys(tx)
		unblockBlockedKeys(tx)
		err = tx.First(&key, "key_id = ?", id).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	w.Header().Add("x-api-key", key.KeyID)
	w.Header().Add("last_accessed", key.LastAccessed.String())
	w.Header().Add("user_access_time", key.UserAccessTime.String())
	w.Header().Add("busy", strconv.FormatBool(key.Busy))
	w.WriteHeader(200)
}

func DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	dbconn := db.GetDB()
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		deleteOldKeys(tx)
		unblockBlockedKeys(tx)
		err = dbconn.Delete(&key, "key_id = ?", id).Error
		if err != nil {
			return err
		}
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
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		deleteOldKeys(tx)
		unblockBlockedKeys(tx)
		err = dbconn.First(&key, "key_id = ?", id).Error
		if err != nil {
			return err
		}
		key.Busy = false
		key.UserAccessTime = nil
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
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		response.Error(w, 400, err)
		return
	}
	err = dbconn.Transaction(func(tx *gorm.DB) error {
		deleteOldKeys(tx)
		unblockBlockedKeys(tx)
		err = dbconn.First(&key, "key_id = ?", id).Error
		if err != nil {
			return err
		}
		if key.KeyID == "" {
			return err
		}
		key.LastAccessed = time.Now()
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
	response.JSON(w, 200, map[string]interface{}{"x-api-key": key.KeyID, "last_accessed": key.LastAccessed, "user_access_time": key.UserAccessTime, "busy": key.Busy}, nil)
}
