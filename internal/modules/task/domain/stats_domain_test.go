package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateCompletionRate(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		total     int
		expected  float64
	}{
		{
			name:      "full completion",
			completed: 10,
			total:     10,
			expected:  100.0,
		},
		{
			name:      "half completion",
			completed: 5,
			total:     10,
			expected:  50.0,
		},
		{
			name:      "no completion",
			completed: 0,
			total:     10,
			expected:  0.0,
		},
		{
			name:      "zero total",
			completed: 0,
			total:     0,
			expected:  0.0,
		},
		{
			name:      "partial completion",
			completed: 1,
			total:     3,
			expected:  33.333333333333336,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCompletionRate(tt.completed, tt.total)
			assert.InDelta(t, tt.expected, result, 0.000001, "Completion rate should match expected value")
		})
	}
}

func TestGetProgressColor(t *testing.T) {
	tests := []struct {
		name           string
		completionRate float64
		expected       ProgressColor
	}{
		{
			name:           "100% completion",
			completionRate: 100.0,
			expected:       ColorDarkGreen,
		},
		{
			name:           "90% completion",
			completionRate: 90.0,
			expected:       ColorGreen,
		},
		{
			name:           "80% completion",
			completionRate: 80.0,
			expected:       ColorGreen,
		},
		{
			name:           "70% completion",
			completionRate: 70.0,
			expected:       ColorYellow,
		},
		{
			name:           "60% completion",
			completionRate: 60.0,
			expected:       ColorYellow,
		},
		{
			name:           "50% completion",
			completionRate: 50.0,
			expected:       ColorOrange,
		},
		{
			name:           "40% completion",
			completionRate: 40.0,
			expected:       ColorOrange,
		},
		{
			name:           "30% completion",
			completionRate: 30.0,
			expected:       ColorLightRed,
		},
		{
			name:           "20% completion",
			completionRate: 20.0,
			expected:       ColorLightRed,
		},
		{
			name:           "10% completion",
			completionRate: 10.0,
			expected:       ColorRed,
		},
		{
			name:           "1% completion",
			completionRate: 1.0,
			expected:       ColorRed,
		},
		{
			name:           "0% completion",
			completionRate: 0.0,
			expected:       ColorGray,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProgressColor(tt.completionRate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProgressLevel(t *testing.T) {
	tests := []struct {
		name           string
		completionRate float64
		expectedLevel  ProgressLevel
	}{
		{
			name:           "100% completion",
			completionRate: 100.0,
			expectedLevel: ProgressLevel{
				Percentage: 100,
				Color:      ColorDarkGreen,
				Label:      "完了",
			},
		},
		{
			name:           "85% completion",
			completionRate: 85.0,
			expectedLevel: ProgressLevel{
				Percentage: 85,
				Color:      ColorGreen,
				Label:      "優秀",
			},
		},
		{
			name:           "65% completion",
			completionRate: 65.0,
			expectedLevel: ProgressLevel{
				Percentage: 65,
				Color:      ColorYellow,
				Label:      "良好",
			},
		},
		{
			name:           "45% completion",
			completionRate: 45.0,
			expectedLevel: ProgressLevel{
				Percentage: 45,
				Color:      ColorOrange,
				Label:      "普通",
			},
		},
		{
			name:           "25% completion",
			completionRate: 25.0,
			expectedLevel: ProgressLevel{
				Percentage: 25,
				Color:      ColorLightRed,
				Label:      "要改善",
			},
		},
		{
			name:           "5% completion",
			completionRate: 5.0,
			expectedLevel: ProgressLevel{
				Percentage: 5,
				Color:      ColorRed,
				Label:      "低調",
			},
		},
		{
			name:           "0% completion",
			completionRate: 0.0,
			expectedLevel: ProgressLevel{
				Percentage: 0,
				Color:      ColorGray,
				Label:      "未着手",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProgressLevel(tt.completionRate)
			assert.Equal(t, tt.expectedLevel, result)
		})
	}
}

func TestGetWeekdayName(t *testing.T) {
	tests := []struct {
		weekday  time.Weekday
		expected string
	}{
		{time.Monday, "Monday"},
		{time.Tuesday, "Tuesday"},
		{time.Wednesday, "Wednesday"},
		{time.Thursday, "Thursday"},
		{time.Friday, "Friday"},
		{time.Saturday, "Saturday"},
		{time.Sunday, "Sunday"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GetWeekdayName(tt.weekday)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetWeekdayNameJP(t *testing.T) {
	tests := []struct {
		weekday  time.Weekday
		expected string
	}{
		{time.Monday, "月"},
		{time.Tuesday, "火"},
		{time.Wednesday, "水"},
		{time.Thursday, "木"},
		{time.Friday, "金"},
		{time.Saturday, "土"},
		{time.Sunday, "日"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GetWeekdayNameJP(tt.weekday)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetWeekStartEnd(t *testing.T) {
	tests := []struct {
		name          string
		date          time.Time
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			name:          "Monday - start of week",
			date:          time.Date(2024, 1, 1, 15, 30, 45, 0, time.UTC), // Monday
			expectedStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:          "Wednesday - middle of week",
			date:          time.Date(2024, 1, 3, 10, 15, 20, 0, time.UTC),         // Wednesday
			expectedStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),            // Previous Monday
			expectedEnd:   time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC), // Sunday
		},
		{
			name:          "Sunday - end of week",
			date:          time.Date(2024, 1, 7, 20, 30, 0, 0, time.UTC),          // Sunday
			expectedStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),            // Monday
			expectedEnd:   time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC), // Same Sunday
		},
		{
			name:          "Different month boundary",
			date:          time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),           // Thursday
			expectedStart: time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC),           // Previous Monday
			expectedEnd:   time.Date(2024, 2, 4, 23, 59, 59, 999999999, time.UTC), // Sunday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := GetWeekStartEnd(tt.date)

			assert.Equal(t, tt.expectedStart, start)
			assert.Equal(t, tt.expectedEnd, end)

			// Verify that start is Monday and end is Sunday
			assert.Equal(t, time.Monday, start.Weekday())
			assert.Equal(t, time.Sunday, end.Weekday())

			// Verify that the input date is within the range
			assert.True(t, tt.date.After(start) || tt.date.Equal(start))
			assert.True(t, tt.date.Before(end) || tt.date.Equal(end))
		})
	}
}

func TestGetDayStartEnd(t *testing.T) {
	tests := []struct {
		name          string
		date          time.Time
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			name:          "normal day",
			date:          time.Date(2024, 1, 15, 14, 30, 45, 123456789, time.UTC),
			expectedStart: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:          "start of day",
			date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedStart: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:          "end of day",
			date:          time.Date(2024, 1, 15, 23, 59, 59, 999999999, time.UTC),
			expectedStart: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 15, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := GetDayStartEnd(tt.date)

			assert.Equal(t, tt.expectedStart, start)
			assert.Equal(t, tt.expectedEnd, end)

			// Verify that the input date is within the range
			assert.True(t, tt.date.After(start) || tt.date.Equal(start))
			assert.True(t, tt.date.Before(end) || tt.date.Equal(end))
		})
	}
}

func TestNewDailyStats(t *testing.T) {
	date := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Create sample tasks
	tasks := []*Task{
		{
			ID:      "task1",
			Status:  TaskStatusDone,
			DueDate: &[]time.Time{time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC)}[0], // Overdue
		},
		{
			ID:     "task2",
			Status: TaskStatusInProgress,
		},
		{
			ID:     "task3",
			Status: TaskStatusTodo,
		},
		{
			ID:     "task4",
			Status: TaskStatusDone,
		},
	}

	stats := NewDailyStats(date, tasks)

	require.NotNil(t, stats)
	assert.Equal(t, date, stats.Date)
	assert.Equal(t, 4, stats.TotalTasks)
	assert.Equal(t, 2, stats.CompletedTasks)
	assert.Equal(t, 1, stats.InProgressTasks)
	assert.Equal(t, 1, stats.TodoTasks)
	assert.Equal(t, 0, stats.OverdueTasks)
	assert.Equal(t, 50.0, stats.CompletionRate) // 2/4 * 100
}

func TestNewDailyStats_EmptyTasks(t *testing.T) {
	date := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	tasks := []*Task{}

	stats := NewDailyStats(date, tasks)

	require.NotNil(t, stats)
	assert.Equal(t, date, stats.Date)
	assert.Equal(t, 0, stats.TotalTasks)
	assert.Equal(t, 0, stats.CompletedTasks)
	assert.Equal(t, 0, stats.InProgressTasks)
	assert.Equal(t, 0, stats.TodoTasks)
	assert.Equal(t, 0, stats.OverdueTasks)
	assert.Equal(t, 0.0, stats.CompletionRate)
}

func TestNewWeeklyStats(t *testing.T) {
	weekStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)          // Monday
	weekEnd := time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC) // Sunday

	dailyStats := map[string]*DailyStats{
		"Monday": {
			Date:           time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			TotalTasks:     5,
			CompletedTasks: 3,
		},
		"Tuesday": {
			Date:           time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			TotalTasks:     3,
			CompletedTasks: 2,
		},
		"Wednesday": {
			Date:           time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			TotalTasks:     4,
			CompletedTasks: 4,
		},
	}

	stats := NewWeeklyStats(weekStart, weekEnd, dailyStats)

	require.NotNil(t, stats)
	assert.Equal(t, weekStart, stats.WeekStart)
	assert.Equal(t, weekEnd, stats.WeekEnd)
	assert.Equal(t, 12, stats.TotalTasks)       // 5+3+4
	assert.Equal(t, 9, stats.CompletedTasks)    // 3+2+4
	assert.Equal(t, 75.0, stats.CompletionRate) // 9/12 * 100
	assert.Equal(t, dailyStats, stats.DailyStats)
}

func TestNewWeeklyStats_EmptyDailyStats(t *testing.T) {
	weekStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC)
	dailyStats := map[string]*DailyStats{}

	stats := NewWeeklyStats(weekStart, weekEnd, dailyStats)

	require.NotNil(t, stats)
	assert.Equal(t, weekStart, stats.WeekStart)
	assert.Equal(t, weekEnd, stats.WeekEnd)
	assert.Equal(t, 0, stats.TotalTasks)
	assert.Equal(t, 0, stats.CompletedTasks)
	assert.Equal(t, 0.0, stats.CompletionRate)
	assert.Equal(t, dailyStats, stats.DailyStats)
}

// Test edge cases and integration scenarios

func TestProgressColor_Boundaries(t *testing.T) {
	// Test exact boundary values
	tests := []struct {
		rate     float64
		expected ProgressColor
	}{
		{99.99, ColorGreen},
		{100.0, ColorDarkGreen},
		{100.01, ColorDarkGreen},
		{79.99, ColorYellow},
		{80.0, ColorGreen},
		{59.99, ColorOrange},
		{60.0, ColorYellow},
		{39.99, ColorLightRed},
		{40.0, ColorOrange},
		{19.99, ColorRed},
		{20.0, ColorLightRed},
		{0.99, ColorGray},
		{1.0, ColorRed},
		{-1.0, ColorGray},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2f", tt.rate), func(t *testing.T) {
			result := GetProgressColor(tt.rate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWeekCalculation_AcrossYearBoundary(t *testing.T) {
	// Test week calculation across year boundary
	date := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // January 1, 2024 (Monday)

	start, end := GetWeekStartEnd(date)

	// Should be the same week since Jan 1, 2024 is a Monday
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), start)
	assert.Equal(t, time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC), end)
}

func TestWeekCalculation_PreviousYear(t *testing.T) {
	// Test when Monday falls in previous year
	date := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC) // January 3, 2024 (Wednesday)

	start, end := GetWeekStartEnd(date)

	// Monday should be January 1, 2024
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), start)
	assert.Equal(t, time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC), end)
}

func TestTaskStats_RealWorldScenario(t *testing.T) {
	// Simulate a real-world daily stats scenario
	date := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)

	tasks := []*Task{
		// Completed tasks
		{ID: "1", Status: TaskStatusDone},
		{ID: "2", Status: TaskStatusDone},
		{ID: "3", Status: TaskStatusDone},
		// In progress
		{ID: "4", Status: TaskStatusInProgress},
		{ID: "5", Status: TaskStatusInProgress},
		// Todo
		{ID: "6", Status: TaskStatusTodo},
		// Overdue task
		{
			ID:      "7",
			Status:  TaskStatusTodo,
			DueDate: &[]time.Time{time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC)}[0],
		},
	}

	stats := NewDailyStats(date, tasks)

	assert.Equal(t, 7, stats.TotalTasks)
	assert.Equal(t, 3, stats.CompletedTasks)
	assert.Equal(t, 2, stats.InProgressTasks)
	assert.Equal(t, 2, stats.TodoTasks) // Including overdue
	assert.Equal(t, 1, stats.OverdueTasks)

	expectedRate := CalculateCompletionRate(3, 7) // ~42.86%
	assert.Equal(t, expectedRate, stats.CompletionRate)

	// Check progress level
	level := GetProgressLevel(stats.CompletionRate)
	assert.Equal(t, ColorOrange, level.Color)
	assert.Equal(t, "普通", level.Label)
}
