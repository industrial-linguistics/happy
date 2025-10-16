package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "/var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db"

var positiveMessages = []string{
	"You're doing an amazing job!",
	"Your code is getting better every day!",
	"Keep up the excellent work!",
	"You're making great progress!",
	"Your debugging skills are impressive!",
	"That was a clever solution!",
	"You're really getting the hang of this!",
	"Your attention to detail is fantastic!",
	"You're asking all the right questions!",
	"Great job thinking through that problem!",
	"Your code is clean and well-organized!",
	"You're becoming a strong developer!",
	"That refactoring was spot-on!",
	"Your test coverage is excellent!",
	"You're a natural at this!",
	"Your problem-solving skills shine!",
	"You write very readable code!",
	"Your commit messages are clear and helpful!",
	"You're mastering these concepts quickly!",
	"Your architecture decisions are sound!",
	"You're great at breaking down complex problems!",
	"Your API design is intuitive!",
	"You're thinking like a senior developer!",
	"Your code reviews are thoughtful!",
	"You're building something impressive!",
	"Your persistence is paying off!",
	"You're learning at an amazing pace!",
	"Your error handling is robust!",
	"You write elegant solutions!",
	"Your documentation is clear and helpful!",
	"You're making this look easy!",
	"Your variable names are descriptive!",
	"You're following best practices perfectly!",
	"Your curiosity drives great code!",
	"You're building confidence with every line!",
	"Your code is production-ready!",
	"You understand the fundamentals deeply!",
	"Your incremental approach is smart!",
	"You're collaborating effectively!",
	"Your testing strategy is solid!",
	"You're thinking about edge cases!",
	"Your code is maintainable!",
	"You're writing self-documenting code!",
	"Your logic is clear and correct!",
	"You're balancing speed and quality well!",
	"Your debugging process is methodical!",
	"You're learning from every mistake!",
	"Your git workflow is professional!",
	"You're asking for help at the right times!",
	"Your code reflects deep understanding!",
}

func main() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Error creating schema: %v", err)
	}

	// Populate messages
	for i, msg := range positiveMessages {
		category := "encouragement"
		if i%3 == 0 {
			category = "achievement"
		} else if i%3 == 1 {
			category = "persistence"
		}

		_, err := db.Exec(`
            INSERT INTO messages (message, category) VALUES (?, ?)
        `, msg, category)

		if err != nil {
			log.Printf("Error inserting message: %v", err)
		}
	}

	fmt.Println("Database initialized successfully!")
	fmt.Printf("Created %d positive messages\n", len(positiveMessages))
}

const schema = `
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY,
    message TEXT NOT NULL,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_category ON messages(category);

CREATE TABLE IF NOT EXISTS user_messages (
    message_id TEXT PRIMARY KEY,
    from_user TEXT NOT NULL,
    to_user TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_messages_recipient ON user_messages(to_user, created_at);

CREATE TABLE IF NOT EXISTS activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    endpoint TEXT NOT NULL,
    name TEXT,
    session_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    response_code INTEGER,
    response_time_ms INTEGER
);

CREATE INDEX IF NOT EXISTS idx_activity_timestamp ON activity_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_activity_session ON activity_log(session_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_activity_name ON activity_log(name, timestamp);

CREATE TABLE IF NOT EXISTS request_stats (
    ip_address TEXT NOT NULL,
    minute_bucket TEXT NOT NULL,
    request_count INTEGER DEFAULT 1,
    PRIMARY KEY (ip_address, minute_bucket)
);

CREATE INDEX IF NOT EXISTS idx_stats_bucket ON request_stats(minute_bucket);
`
