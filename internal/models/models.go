package models

type Usage struct {
	UnitNumber                int     `json:"unit_number"`
	CalendarDate              string  `json:"calendar_date"`
	LeftCookTime              float64 `json:"left_stove_cooktime"`
	RightCookTime             float64 `json:"right_stove_cooktime"`
	DailyCookingTime          float64 `json:"daily_cooking_time"`
	DailyPowerConsumption     float64 `json:"daily_power_consumption"`
	StoveOnOffCount           float64 `json:"stove_on_off_count"`
	AvgCookingTimePerUse      float64 `json:"average_cooking_time_per_use"`
	AvgPowerConsumptionPerUse float64 `json:"average_power_consumption_per_use"`
}

type Stats struct {
	UnitNumber            int     `json:"unit_number"`
	TotalPowerConsumption float64 `json:"total_power_consumption"`
	AvgPowerConsumption   float64 `json:"avg_power_consumption"`
}

type StatsUser struct {
	TotalPowerConsumption float64 `json:"total_power_consumption"`
	AvgPowerConsumption   float64 `json:"avg_power_consumption"`
}

type DeviceUsage struct {
	DeviceID string  `json:"device_id"`
	Date     string  `json:"date"`
	EGPDY    float64 `json:"EG_p_d_y"` // Amount of electricity used by the project device in kWh
}
