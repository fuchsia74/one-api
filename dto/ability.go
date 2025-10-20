package dto

// EnabledAbility represents channel metadata exposed to API consumers for an ability lookup.
type EnabledAbility struct {
	Model       string `json:"model" gorm:"model"`
	ChannelType int    `json:"channel_type" gorm:"channel_type"`
	ChannelId   int    `json:"channel_id" gorm:"channel_id"`
}
