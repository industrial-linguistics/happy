package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "/var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db"

func main() {
	// Command-line flags
	modeFlag := flag.String("mode", "live", "Mode: live, summary, students, export")
	tailFlag := flag.Int("tail", 20, "Number of recent entries to show")
	sinceFlag := flag.String("since", "", "Show activity since timestamp (RFC3339)")
	studentFlag := flag.String("student", "", "Filter by student name")

	flag.Parse()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch *modeFlag {
	case "live":
		runLiveMode(db, *tailFlag)
	case "summary":
		runSummary(db, *sinceFlag)
	case "students":
		runStudentProgress(db)
	case "export":
		runExport(db, *sinceFlag, *studentFlag)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *modeFlag)
		flag.Usage()
		os.Exit(1)
	}
}

func runLiveMode(db *sql.DB, tail int) {
	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	for {
		// Clear screen and move to top
		fmt.Print("\033[2J\033[H")

		fmt.Println("=== Happy API Activity Monitor ===")
		fmt.Printf("Last updated: %s (Press Ctrl+C to exit)\n\n", time.Now().Format("15:04:05"))

		// Get active users with their latest activity
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
			fmt.Printf("Error querying database: %v\n", err)
			time.Sleep(3 * time.Second)
			continue
		}

		type UserActivity struct {
			name       string
			lastSeen   time.Time
			endpoint   string
			totalCount int
			errorCount int
		}

		var users []UserActivity
		for rows.Next() {
			var u UserActivity
			rows.Scan(&u.name, &u.lastSeen, &u.endpoint, &u.totalCount, &u.errorCount)
			users = append(users, u)
		}
		rows.Close()

		if len(users) == 0 {
			fmt.Println("No activity in the last hour.")
		} else {
			// Print header
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintf(w, "Name\tLast Seen\tLast Activity\tRequests\tErrors\n")
			fmt.Fprintf(w, "----\t---------\t-------------\t--------\t------\n")

			// Print each user
			for _, u := range users {
				ago := time.Since(u.lastSeen).Round(time.Second)
				agoStr := formatDuration(ago)

				errorStr := ""
				if u.errorCount > 0 {
					errorStr = fmt.Sprintf("%d", u.errorCount)
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
					truncate(u.name, 20),
					agoStr,
					u.endpoint,
					u.totalCount,
					errorStr)
			}
			w.Flush()

			// Show summary
			fmt.Printf("\nTotal active users: %d\n", len(users))
		}

		time.Sleep(3 * time.Second)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	} else {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
}

func runSummary(db *sql.DB, since string) {
	whereClause := ""
	args := []interface{}{}

	if since != "" {
		whereClause = "WHERE timestamp >= ?"
		args = append(args, since)
	} else {
		// Default: last 2 hours
		whereClause = "WHERE timestamp >= datetime('now', '-2 hours')"
	}

	fmt.Println("=== Activity Summary ===\n")

	// Total requests
	var totalRequests int
	db.QueryRow("SELECT COUNT(*) FROM activity_log "+whereClause, args...).Scan(&totalRequests)
	fmt.Printf("Total Requests: %d\n", totalRequests)

	// Unique students
	var uniqueStudents int
	db.QueryRow(`
        SELECT COUNT(DISTINCT name)
        FROM activity_log
        `+whereClause+` AND name IS NOT NULL
    `, args...).Scan(&uniqueStudents)
	fmt.Printf("Active Students: %d\n\n", uniqueStudents)

	// Requests by endpoint
	fmt.Println("Requests by Endpoint:")
	rows, _ := db.Query(`
        SELECT endpoint, COUNT(*) as count
        FROM activity_log
        `+whereClause+`
        GROUP BY endpoint
        ORDER BY count DESC
    `, args...)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for rows.Next() {
		var endpoint string
		var count int
		rows.Scan(&endpoint, &count)
		fmt.Fprintf(w, "  %s\t%d\n", endpoint, count)
	}
	w.Flush()
	rows.Close()

	fmt.Println()

	// Error rate
	var errorCount int
	db.QueryRow(`
        SELECT COUNT(*) FROM activity_log
        `+whereClause+` AND response_code >= 400
    `, args...).Scan(&errorCount)

	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(errorCount) / float64(totalRequests) * 100
	}
	fmt.Printf("Error Rate: %.1f%% (%d errors)\n", errorRate, errorCount)
}

func runStudentProgress(db *sql.DB) {
	fmt.Println("=== Student Progress ===\n")

	rows, _ := db.Query(`
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "Student\tRequests\tFirst Seen\tLast Seen\tSessions\n")
	fmt.Fprintf(w, "-------\t--------\t----------\t---------\t--------\n")

	for rows.Next() {
		var name string
		var totalRequests, sessions int
		var firstSeen, lastSeen time.Time

		rows.Scan(&name, &totalRequests, &firstSeen, &lastSeen, &sessions)

		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%d\n",
			truncate(name, 20),
			totalRequests,
			firstSeen.Format("15:04"),
			lastSeen.Format("15:04"),
			sessions)
	}

	w.Flush()
	rows.Close()

	// Show who hasn't been seen recently
	fmt.Println("\n=== Inactive Students (>15 min) ===\n")

	rows, _ = db.Query(`
        SELECT name, MAX(timestamp) as last_seen
        FROM activity_log
        WHERE name IS NOT NULL
        GROUP BY name
        HAVING last_seen < datetime('now', '-15 minutes')
        ORDER BY last_seen DESC
    `)

	for rows.Next() {
		var name string
		var lastSeen time.Time
		rows.Scan(&name, &lastSeen)

		ago := time.Since(lastSeen).Round(time.Minute)
		fmt.Printf("  %s (last seen %v ago)\n", name, ago)
	}
	rows.Close()
}

func runExport(db *sql.DB, since, student string) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if since != "" {
		whereClause += " AND timestamp >= ?"
		args = append(args, since)
	}

	if student != "" {
		whereClause += " AND name = ?"
		args = append(args, student)
	}

	rows, err := db.Query(`
        SELECT timestamp, name, endpoint, session_id, ip_address,
               response_code, response_time_ms
        FROM activity_log
        `+whereClause+`
        ORDER BY timestamp
    `, args...)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// CSV output
	fmt.Println("timestamp,name,endpoint,session_id,ip_address,response_code,response_time_ms")

	for rows.Next() {
		var ts time.Time
		var name, endpoint, sessionID, ip sql.NullString
		var responseCode sql.NullInt64
		var responseTime sql.NullInt64

		rows.Scan(&ts, &name, &endpoint, &sessionID, &ip, &responseCode, &responseTime)

		fmt.Printf("%s,%s,%s,%s,%s,%d,%d\n",
			ts.Format(time.RFC3339),
			nullStringOr(name, ""),
			nullStringOr(endpoint, ""),
			nullStringOr(sessionID, ""),
			nullStringOr(ip, ""),
			nullInt64Or(responseCode, 0),
			nullInt64Or(responseTime, 0))
	}
	rows.Close()
}

func showRecentActivity(db *sql.DB, fromID, toID int64) {
	rows, _ := db.Query(`
        SELECT id, timestamp, name, endpoint, response_code, session_id
        FROM activity_log
        WHERE id > ? AND id <= ?
        ORDER BY id
    `, fromID, toID)

	for rows.Next() {
		var id int64
		var ts time.Time
		var name, endpoint, sessionID sql.NullString
		var responseCode int

		rows.Scan(&id, &ts, &name, &endpoint, &responseCode, &sessionID)

		status := "✓"
		if responseCode >= 400 {
			status = "✗"
		}

		nameStr := "anonymous"
		if name.Valid && name.String != "" {
			nameStr = name.String
		}

		fmt.Printf("%s [%s] %-15s %s\n",
			status,
			ts.Format("15:04:05"),
			truncate(nameStr, 15),
			endpoint.String)
	}
	rows.Close()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func nullStringOr(ns sql.NullString, def string) string {
	if ns.Valid {
		return ns.String
	}
	return def
}

func nullInt64Or(ni sql.NullInt64, def int64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return def
}
