package main

import (
	"time"

	"gorm.io/gorm"
)

type CommonModelPrimaryKey struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	CreatedAt *time.Time `json:"created_at" time_format:"2006-01-02 15:04:05"`
	UpdatedAt *time.Time `json:"updated_at" time_format:"2006-01-02 15:04:05"`
}

type Configurations struct {
	CommonModelPrimaryKey
	OrganizationID    int            `json:"organization_id"`
	Name              string         `json:"name" filter:"true" display_name:"Name" can_be_hidden:"false" sortable:"true" default:"true"`
	Platform          string         `json:"platform" filter:"true" display_name:"Platform" can_be_hidden:"false" sortable:"true" default:"true" options:"Desktop,Mobile"`
	IsKaneSupported   bool           `json:"is_kane_supported" filter:"true" display_name:"Kane AI Supported" can_be_hidden:"true" sortable:"false" default:"false"`
	IsManualSupported bool           `json:"is_manual_supported" filter:"true" display_name:"Manual Supported" can_be_hidden:"true" sortable:"false" default:"false"`
	CreatedBy         int            `json:"created_by" filter:"true" display_name:"Created By" can_be_hidden:"false" sortable:"false" default:"true"`
	UpdatedBy         int            `json:"updated_by" filter:"true" display_name:"Updated By" can_be_hidden:"false" sortable:"false" default:"false"`
	IsDefault         bool           `json:"is_default" filter:"true" display_name:"Default" can_be_hidden:"true" sortable:"false" default:"false"`
	IsCustom          bool           `json:"is_custom"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" filter:"false" display_name:"Deleted At" can_be_hidden:"true" sortable:"false" default:"false"`
	IsComplete        bool           `json:"is_complete" filter:"true" display_name:"Complete" can_be_hidden:"true" sortable:"false" default:"false"`

	TestEnvironments []TestEnvironment `json:"test_environments" gorm:"foreignKey:ConfigurationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (m *Configurations) GetTableName() string {
	return "configurations"
}

type TestEnvironment struct {
	CommonModelPrimaryKey
	ConfigurationID  int64          `json:"configuration_id" filter:"true" display_name:"Configurations ID" can_be_hidden:"false" sortable:"false" default:"false"`
	UDID             string         `gorm:"column:udid" json:"udid" filter:"true" display_name:"UDID" can_be_hidden:"false" sortable:"false" default:"false"`
	PlatformType     string         `json:"platform_type" filter:"true" display_name:"Platform Type" can_be_hidden:"false" sortable:"false" default:"false"`
	OrganizationID   int            `json:"organization_id"`
	Name             string         `json:"name" filter:"true" display_name:"Name" can_be_hidden:"false" sortable:"true" default:"true"`
	Brand            *string        `json:"brand"`
	OSName           *string        `json:"os_name" filter:"true" display_name:"OS" can_be_hidden:"false" sortable:"false" default:"true" options:"Windows,macOS,Android,iOS"`
	OS               *string        `json:"os" filter:"true" display_name:"OS" can_be_hidden:"false" sortable:"false" default:"true"`
	BrowserVersion   *string        `json:"browser_version"`
	Resolution       *string        `json:"resolution"`
	Browser          *string        `json:"browser" filter:"true" display_name:"Browser" can_be_hidden:"false" sortable:"false" default:"true" options:"Firefox,Edge,Chrome,Brave,Opera,IE,Yandex,Safari,Edge,Chromium"`
	OSVersion        *string        `json:"os_version"`
	Device           *string        `json:"device"`
	Platform         *string        `json:"platform" filter:"true" display_name:"Platform" can_be_hidden:"false" sortable:"true" default:"true" options:"Desktop,Mobile"`
	OSVersionID      *string        `json:"os_version_id"`
	OSID             *string        `json:"os_id"`
	BrowserID        *string        `json:"browser_id"`
	BrowserVersionID *string        `json:"browser_version_id"`
	ResolutionID     *string        `json:"resolution_id"`
	DeviceID         *string        `json:"device_id"`
	ManufacturerID   *string        `json:"manufacturer_id"` //ManufacturerID is same as BrandID
	AppID            *string        `json:"app_id"`
	URL              *string        `json:"url"`
	IsCustom         bool           `json:"is_custom"`
	IsComplete       bool           `json:"is_complete"`
	IsKaneSupported  bool           `json:"is_kane_supported" filter:"true" display_name:"Kane AI Supported" can_be_hidden:"true" sortable:"false" default:"false"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at" filter:"false" display_name:"Deleted At" can_be_hidden:"true" sortable:"false" default:"false"`
	PrivateCloud     bool           `json:"private_cloud" filter:"true" display_name:"Private Cloud" can_be_hidden:"true" sortable:"false" default:"false"`
	IsDefault        bool           `json:"is_default" filter:"true" display_name:"Default" can_be_hidden:"true" sortable:"false" default:"false"`
	MetaData         *string        `gorm:"column:metadata" json:"metadata"`
}

func (m *TestEnvironment) GetTableName() string {
	return "test_environments"
}
