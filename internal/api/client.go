package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.prod.whoop.com/developer"

type Client struct {
	http *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{http: httpClient}
}

type PaginatedResponse[T any] struct {
	Records   []T    `json:"records"`
	NextToken string `json:"next_token"`
}

type UserProfile struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type BodyMeasurement struct {
	HeightMeter    float64 `json:"height_meter"`
	WeightKilogram float64 `json:"weight_kilogram"`
	MaxHeartRate   int     `json:"max_heart_rate"`
}

type CycleScore struct {
	Strain           float64 `json:"strain"`
	Kilojoule        float64 `json:"kilojoule"`
	AverageHeartRate int     `json:"average_heart_rate"`
	MaxHeartRate     int     `json:"max_heart_rate"`
}

type Cycle struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"user_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Start          time.Time   `json:"start"`
	End            *time.Time  `json:"end"`
	TimezoneOffset string      `json:"timezone_offset"`
	ScoreState     string      `json:"score_state"`
	Score          *CycleScore `json:"score"`
}

type RecoveryScore struct {
	UserCalibrating  bool     `json:"user_calibrating"`
	RecoveryScore    float64  `json:"recovery_score"`
	RestingHeartRate float64  `json:"resting_heart_rate"`
	HrvRmssdMilli    float64  `json:"hrv_rmssd_milli"`
	Spo2Percentage   *float64 `json:"spo2_percentage"`
	SkinTempCelsius  *float64 `json:"skin_temp_celsius"`
}

type Recovery struct {
	CycleID    int64          `json:"cycle_id"`
	SleepID    string         `json:"sleep_id"`
	UserID     int64          `json:"user_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	ScoreState string         `json:"score_state"`
	Score      *RecoveryScore `json:"score"`
}

type SleepStageSummary struct {
	TotalInBedTimeMilli        int `json:"total_in_bed_time_milli"`
	TotalAwakeTimeMilli        int `json:"total_awake_time_milli"`
	TotalNoDataTimeMilli       int `json:"total_no_data_time_milli"`
	TotalLightSleepTimeMilli   int `json:"total_light_sleep_time_milli"`
	TotalSlowWaveSleepTimeMilli int `json:"total_slow_wave_sleep_time_milli"`
	TotalRemSleepTimeMilli     int `json:"total_rem_sleep_time_milli"`
	SleepCycleCount            int `json:"sleep_cycle_count"`
	DisturbanceCount           int `json:"disturbance_count"`
}

type SleepNeeded struct {
	BaselineMilli             int64 `json:"baseline_milli"`
	NeedFromSleepDebtMilli    int64 `json:"need_from_sleep_debt_milli"`
	NeedFromRecentStrainMilli int64 `json:"need_from_recent_strain_milli"`
	NeedFromRecentNapMilli    int64 `json:"need_from_recent_nap_milli"`
}

type SleepScore struct {
	StageSummary                SleepStageSummary `json:"stage_summary"`
	SleepNeeded                 SleepNeeded       `json:"sleep_needed"`
	RespiratoryRate             *float64          `json:"respiratory_rate"`
	SleepPerformancePercentage  *float64          `json:"sleep_performance_percentage"`
	SleepConsistencyPercentage  *float64          `json:"sleep_consistency_percentage"`
	SleepEfficiencyPercentage   *float64          `json:"sleep_efficiency_percentage"`
}

type Sleep struct {
	ID             string      `json:"id"`
	CycleID        int64       `json:"cycle_id"`
	UserID         int64       `json:"user_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Start          time.Time   `json:"start"`
	End            time.Time   `json:"end"`
	TimezoneOffset string      `json:"timezone_offset"`
	Nap            bool        `json:"nap"`
	ScoreState     string      `json:"score_state"`
	Score          *SleepScore `json:"score"`
}

type ZoneDurations struct {
	ZoneZeroMilli  int64 `json:"zone_zero_milli"`
	ZoneOneMilli   int64 `json:"zone_one_milli"`
	ZoneTwoMilli   int64 `json:"zone_two_milli"`
	ZoneThreeMilli int64 `json:"zone_three_milli"`
	ZoneFourMilli  int64 `json:"zone_four_milli"`
	ZoneFiveMilli  int64 `json:"zone_five_milli"`
}

type WorkoutScore struct {
	Strain             float64       `json:"strain"`
	AverageHeartRate   int           `json:"average_heart_rate"`
	MaxHeartRate       int           `json:"max_heart_rate"`
	Kilojoule          float64       `json:"kilojoule"`
	PercentRecorded    float64       `json:"percent_recorded"`
	DistanceMeter      *float64      `json:"distance_meter"`
	AltitudeGainMeter  *float64      `json:"altitude_gain_meter"`
	AltitudeChangeMeter *float64     `json:"altitude_change_meter"`
	ZoneDurations      ZoneDurations `json:"zone_duration"`
}

type Workout struct {
	ID             string        `json:"id"`
	UserID         int64         `json:"user_id"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Start          time.Time     `json:"start"`
	End            time.Time     `json:"end"`
	TimezoneOffset string        `json:"timezone_offset"`
	SportName      string        `json:"sport_name"`
	SportID        *int          `json:"sport_id"`
	ScoreState     string        `json:"score_state"`
	Score          *WorkoutScore `json:"score"`
}

func (c *Client) get(path string, params url.Values, out any) error {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	resp, err := c.http.Get(u)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) GetProfile() (*UserProfile, error) {
	var p UserProfile
	return &p, c.get("/v2/user/profile/basic", nil, &p)
}

func (c *Client) GetBodyMeasurement() (*BodyMeasurement, error) {
	var b BodyMeasurement
	return &b, c.get("/v2/user/measurement/body", nil, &b)
}

func (c *Client) GetCycles(start, end string, limit int) (*PaginatedResponse[Cycle], error) {
	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if end != "" {
		params.Set("end", end)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	var resp PaginatedResponse[Cycle]
	return &resp, c.get("/v2/cycle", params, &resp)
}

func (c *Client) GetRecoveries(start, end string, limit int) (*PaginatedResponse[Recovery], error) {
	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if end != "" {
		params.Set("end", end)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	var resp PaginatedResponse[Recovery]
	return &resp, c.get("/v2/recovery", params, &resp)
}

func (c *Client) GetSleeps(start, end string, limit int) (*PaginatedResponse[Sleep], error) {
	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if end != "" {
		params.Set("end", end)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	var resp PaginatedResponse[Sleep]
	return &resp, c.get("/v2/activity/sleep", params, &resp)
}

func (c *Client) GetWorkouts(start, end string, limit int) (*PaginatedResponse[Workout], error) {
	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if end != "" {
		params.Set("end", end)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	var resp PaginatedResponse[Workout]
	return &resp, c.get("/v2/activity/workout", params, &resp)
}
