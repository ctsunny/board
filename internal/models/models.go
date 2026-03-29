package models

import "time"

type Customer struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `json:"name"`
	Contact     string    `json:"contact"`
	RegionName  string    `json:"region_name"`
	RouteName   string    `json:"route_name"`
	ServerName  string    `json:"server_name"`
	NodeName    string    `json:"node_name"`
	ExpiresAt   time.Time `json:"expires_at"`
	Amount      float64   `json:"amount"`
	BillingType string    `json:"billing_type"`
	TrafficGB   float64   `json:"traffic_gb"`
	UsedGB      float64   `json:"used_gb"`
	Tags        string    `json:"tags"`
	Status      string    `json:"status"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Region struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

type Server struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Name       string    `json:"name"`
	IP         string    `json:"ip"`
	Location   string    `json:"location"`
	Remark     string    `json:"remark"`
	Status     string    `json:"status"`
	Latency    int       `json:"latency"`
	LastPingAt time.Time `json:"last_ping_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Route struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Protocol  string    `json:"protocol"`
	RegionID  uint      `json:"region_id"`
	Region    Region    `gorm:"foreignKey:RegionID" json:"region"`
	ServerID  uint      `json:"server_id"`
	Server    Server    `gorm:"foreignKey:ServerID" json:"server"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Node struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	RouteID   uint      `json:"route_id"`
	Route     Route     `gorm:"foreignKey:RouteID" json:"route"`
	ServerID  uint      `json:"server_id"`
	Server    Server    `gorm:"foreignKey:ServerID" json:"server"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID uint      `json:"resource_id"`
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	Operator   string    `json:"operator"`
	CreatedAt  time.Time `json:"created_at"`
}

type APIToken struct {
	ID         uint       `gorm:"primarykey" json:"id"`
	Name       string     `json:"name"`
	Token      string     `json:"token" gorm:"uniqueIndex"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type NotificationLog struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	Type           string    `json:"type"`
	RecipientEmail string    `json:"recipient_email"`
	Subject        string    `json:"subject"`
	Status         string    `json:"status"`
	Error          string    `json:"error"`
	CreatedAt      time.Time `json:"created_at"`
}
