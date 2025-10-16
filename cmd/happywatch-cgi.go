package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "/vhosts/happy.industrial-linguistics.com/data/positive-social.db"

type liveUser struct {
	Name       string
	LastSeen   time.Time
	Endpoint   string
	TotalCount int
	ErrorCount int
}

type endpointCount struct {
	Endpoint string
	Count    int
}

type studentProgress struct {
	Name          string
	TotalRequests int
	FirstSeen     time.Time
	LastSeen      time.Time
	Sessions      int
}

type inactiveStudent struct {
	Name     string
	LastSeen time.Time
}

type pageData struct {
	GeneratedAt       time.Time
	LiveUsers         []liveUser
	SummaryTotal      int
	SummaryStudents   int
	SummaryEndpoints  []endpointCount
	SummaryErrorRate  float64
	SummaryErrorCount int
	StudentProgress   []studentProgress
	InactiveStudents  []inactiveStudent
}

func main() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		sendError(http.StatusInternalServerError, fmt.Sprintf("failed to open database: %v", err))
		return
	}
	defer db.Close()

	data, err := gatherPageData(db)
	if err != nil {
		sendError(http.StatusInternalServerError, err.Error())
		return
	}

	if err := pageTemplate.Execute(os.Stdout, data); err != nil {
		sendError(http.StatusInternalServerError, fmt.Sprintf("failed to render template: %v", err))
		return
	}
}

func gatherPageData(db *sql.DB) (pageData, error) {
	liveUsers, err := loadLiveUsers(db)
	if err != nil {
		return pageData{}, fmt.Errorf("failed to load live users: %w", err)
	}

	totalRequests, uniqueStudents, endpointCounts, errorRate, errorCount, err := loadSummary(db)
	if err != nil {
		return pageData{}, fmt.Errorf("failed to load summary: %w", err)
	}

	students, err := loadStudentProgress(db)
	if err != nil {
		return pageData{}, fmt.Errorf("failed to load student progress: %w", err)
	}

	inactive, err := loadInactiveStudents(db)
	if err != nil {
		return pageData{}, fmt.Errorf("failed to load inactive students: %w", err)
	}

	return pageData{
		GeneratedAt:       time.Now(),
		LiveUsers:         liveUsers,
		SummaryTotal:      totalRequests,
		SummaryStudents:   uniqueStudents,
		SummaryEndpoints:  endpointCounts,
		SummaryErrorRate:  errorRate,
		SummaryErrorCount: errorCount,
		StudentProgress:   students,
		InactiveStudents:  inactive,
	}, nil
}

func loadLiveUsers(db *sql.DB) ([]liveUser, error) {
	rows, err := db.Query(`
        SELECT
            a.name,
            a.timestamp as last_seen,
            a.endpoint,
            stats.total_count,
            stats.error_count
        FROM activity_log a
        INNER JOIN (
            SELECT
                name,
                MAX(timestamp) as max_timestamp,
                COUNT(*) as total_count,
                SUM(CASE WHEN response_code >= 400 THEN 1 ELSE 0 END) as error_count
            FROM activity_log
            WHERE name IS NOT NULL AND name != ''
              AND timestamp >= datetime('now', '-1 hour')
            GROUP BY name
        ) stats ON a.name = stats.name AND a.timestamp = stats.max_timestamp
        WHERE a.name IS NOT NULL AND a.name != ''
        ORDER BY a.timestamp DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []liveUser
	for rows.Next() {
		var u liveUser
		if err := rows.Scan(&u.Name, &u.LastSeen, &u.Endpoint, &u.TotalCount, &u.ErrorCount); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func loadSummary(db *sql.DB) (totalRequests int, uniqueStudents int, endpointCounts []endpointCount, errorRate float64, errorCount int, err error) {
	whereClause := "WHERE timestamp >= datetime('now', '-2 hours')"

	if err = db.QueryRow("SELECT COUNT(*) FROM activity_log " + whereClause).Scan(&totalRequests); err != nil {
		return
	}

	if err = db.QueryRow(`
        SELECT COUNT(DISTINCT name)
        FROM activity_log
        ` + whereClause + ` AND name IS NOT NULL
    `).Scan(&uniqueStudents); err != nil {
		return
	}

	rows, err := db.Query(`
        SELECT endpoint, COUNT(*) as count
        FROM activity_log
        ` + whereClause + `
        GROUP BY endpoint
        ORDER BY count DESC
    `)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var ec endpointCount
		if err = rows.Scan(&ec.Endpoint, &ec.Count); err != nil {
			return
		}
		endpointCounts = append(endpointCounts, ec)
	}
	if err = rows.Err(); err != nil {
		return
	}

	if err = db.QueryRow(`
        SELECT COUNT(*) FROM activity_log
        ` + whereClause + ` AND response_code >= 400
    `).Scan(&errorCount); err != nil {
		return
	}

	if totalRequests > 0 {
		errorRate = float64(errorCount) / float64(totalRequests) * 100
	}

	return
}

func loadStudentProgress(db *sql.DB) ([]studentProgress, error) {
	rows, err := db.Query(`
        SELECT
            name,
            COUNT(*) as total_requests,
            MIN(timestamp) as first_seen,
            MAX(timestamp) as last_seen,
            COUNT(DISTINCT session_id) as sessions
        FROM activity_log
        WHERE name IS NOT NULL
          AND timestamp >= datetime('now', '-4 hours')
        GROUP BY name
        ORDER BY total_requests DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []studentProgress
	for rows.Next() {
		var s studentProgress
		if err := rows.Scan(&s.Name, &s.TotalRequests, &s.FirstSeen, &s.LastSeen, &s.Sessions); err != nil {
			return nil, err
		}
		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func loadInactiveStudents(db *sql.DB) ([]inactiveStudent, error) {
	rows, err := db.Query(`
        SELECT name, MAX(timestamp) as last_seen
        FROM activity_log
        WHERE name IS NOT NULL
        GROUP BY name
        HAVING last_seen < datetime('now', '-15 minutes')
        ORDER BY last_seen DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []inactiveStudent
	for rows.Next() {
		var s inactiveStudent
		if err := rows.Scan(&s.Name, &s.LastSeen); err != nil {
			return nil, err
		}
		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func sendError(status int, message string) {
	fmt.Printf("Status: %d %s\r\n", status, http.StatusText(status))
	fmt.Printf("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	fmt.Fprintf(os.Stdout, "%s\n", message)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("15:04")
}

func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format(time.RFC3339)
}

var pageTemplate = template.Must(template.New("page").Funcs(template.FuncMap{
	"formatAgo": func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		return formatDuration(time.Since(t))
	},
	"formatTime":      formatTime,
	"formatTimestamp": formatTimestamp,
}).Parse(`Status: 200 OK
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>happywatch</title>
    <style>
        body { font-family: system-ui, sans-serif; margin: 2rem; background: #f4f6f8; color: #222; }
        h1 { margin-bottom: 0.2rem; }
        h2 { margin-top: 2rem; }
        table { border-collapse: collapse; width: 100%; background: #fff; box-shadow: 0 1px 3px rgba(0,0,0,0.1); margin-top: 1rem; }
        th, td { border: 1px solid #d8dee4; padding: 0.5rem 0.75rem; text-align: left; }
        th { background: #e9eef3; }
        tbody tr:nth-child(even) { background: #f7f9fb; }
        .muted { color: #555; font-size: 0.9rem; }
        .error { color: #b42323; font-weight: bold; }
        .badge { display: inline-block; padding: 0.15rem 0.5rem; border-radius: 999px; background: #e1ecf4; color: #085fa2; font-size: 0.8rem; margin-left: 0.5rem; }
        .card { background: #fff; padding: 1rem; border-radius: 0.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); margin-top: 1rem; }
        ul { padding-left: 1.25rem; }
    </style>
</head>
<body>
    <h1>happywatch</h1>
    <p class="muted">Snapshot generated at {{ .GeneratedAt.Local | formatTimestamp }}</p>

    <h2>Live Activity</h2>
    {{ if .LiveUsers }}
    <table>
        <thead>
            <tr>
                <th>Name</th>
                <th>Last Seen</th>
                <th>Last Activity</th>
                <th>Requests</th>
                <th>Errors</th>
            </tr>
        </thead>
        <tbody>
        {{ range .LiveUsers }}
            <tr>
                <td>{{ .Name }}</td>
                <td>{{ .LastSeen | formatAgo }}</td>
                <td>{{ .Endpoint }}</td>
                <td>{{ .TotalCount }}</td>
                <td>{{ if gt .ErrorCount 0 }}<span class="error">{{ .ErrorCount }}</span>{{ else }}0{{ end }}</td>
            </tr>
        {{ end }}
        </tbody>
    </table>
    <p class="muted">Total active users: {{ len .LiveUsers }}</p>
    {{ else }}
    <div class="card">No activity in the last hour.</div>
    {{ end }}

    <h2>Activity Summary (last 2 hours)</h2>
    <div class="card">
        <p><strong>Total Requests:</strong> {{ .SummaryTotal }}</p>
        <p><strong>Active Students:</strong> {{ .SummaryStudents }}</p>
        <p><strong>Error Rate:</strong> {{ printf "%.1f" .SummaryErrorRate }}% ({{ .SummaryErrorCount }} errors)</p>
    </div>
    {{ if .SummaryEndpoints }}
    <table>
        <thead>
            <tr>
                <th>Endpoint</th>
                <th>Requests</th>
            </tr>
        </thead>
        <tbody>
        {{ range .SummaryEndpoints }}
            <tr>
                <td>{{ .Endpoint }}</td>
                <td>{{ .Count }}</td>
            </tr>
        {{ end }}
        </tbody>
    </table>
    {{ end }}

    <h2>Student Progress (last 4 hours)</h2>
    {{ if .StudentProgress }}
    <table>
        <thead>
            <tr>
                <th>Student</th>
                <th>Requests</th>
                <th>First Seen</th>
                <th>Last Seen</th>
                <th>Sessions</th>
            </tr>
        </thead>
        <tbody>
        {{ range .StudentProgress }}
            <tr>
                <td>{{ .Name }}</td>
                <td>{{ .TotalRequests }}</td>
                <td>{{ .FirstSeen | formatTime }}</td>
                <td>{{ .LastSeen | formatTime }}</td>
                <td>{{ .Sessions }}</td>
            </tr>
        {{ end }}
        </tbody>
    </table>
    {{ else }}
    <div class="card">No recent student activity.</div>
    {{ end }}

    <h2>Inactive Students (&gt;15 minutes)</h2>
    {{ if .InactiveStudents }}
    <ul>
        {{ range .InactiveStudents }}
        <li>{{ .Name }} <span class="badge">last seen {{ .LastSeen | formatAgo }}</span></li>
        {{ end }}
    </ul>
    {{ else }}
    <div class="card">No inactive students in the last period.</div>
    {{ end }}
</body>
</html>
`))
