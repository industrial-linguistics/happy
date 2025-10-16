# Student Exercises - Positive Social Network

Progressive exercises for building a client for the Happy API using Claude Code.

## Setup

**API Base URL:** `https://happy.industrial-linguistics.com/v1`

**Your session ID:** Use `bitmex_java_YYYYMMDD` (e.g., `bitmex_java_20251014`)

---

## Exercise 1: Simple Command-Line Client (10 minutes)

**Goal:** Create a Java program that fetches and displays a positive message.

**Prompt for Claude:**
```
Create a Java program that:
1. Makes an HTTP GET request to https://happy.industrial-linguistics.com/v1/message
2. Passes my name as the 'name' parameter and 'bitmex_java_20251014' as session_id
3. Parses the JSON response
4. Prints the message to the console

Use standard Java libraries (no external dependencies).
Put it in a file called MessageClient.java
```

**Success criteria:**
- Program compiles without errors
- Running it displays a positive message with your name
- Each run increments the sequence number

---

## Exercise 2: Add ONE Test (8 minutes)

**Goal:** Learn to work incrementally with TDD.

**Prompt for Claude:**
```
I want to add testing to MessageClient.java.

First, write ONLY ONE test that verifies we can parse the JSON response correctly.
Use JUnit 5.

Don't refactor anything else. Just add the test class and one test method.
```

**Then:**
```
Now make that test pass by extracting the JSON parsing into a testable method.
Only change what's necessary to make the test pass.
```

**Key lesson:** Control Claude to make small, incremental changes instead of rewriting everything.

**Success criteria:**
- One test file created
- One test method
- Test passes

---

## Exercise 3: Web Interface (10 minutes)

**Goal:** Build a simple HTML/CSS/JavaScript interface.

**Prompt for Claude:**
```
Create a simple landing page for the Positive Social Network:
1. One HTML file with a form asking for the user's name
2. One CSS file for styling (make it cheerful and colorful)
3. JavaScript to fetch from the API and display the message
4. Show the message with a fade-in animation
5. Put files in a 'website' directory

Keep it simple - no frameworks, no build tools.
```

**Success criteria:**
- Can open index.html in a browser
- Entering your name and clicking button shows a message
- Page looks pleasant (not default browser styling)

---

## Exercise 4: Incremental Improvements (12 minutes)

**Goal:** Practice controlling scope with specific prompts.

**Prompt 1:**
```
The button needs to be more prominent. Make ONLY the button styling better.
Don't change anything else.
```

**Prompt 2:**
```
Add a loading indicator that shows while waiting for the API response.
Just the loading indicator - don't refactor other code.
```

**Prompt 3:**
```
Show the sequence number next to the message.
Only add this feature, don't reorganize the code.
```

**Key lesson:** Use precise prompts to prevent Claude from over-engineering.

**Success criteria:**
- Each change is small and focused
- No unrelated files modified
- Features work as requested

---

## Exercise 5: Add Local Storage (15 minutes)

**Goal:** Build incrementally - database first, then save, then retrieve.

**Prompt 1:**
```
Add an SQLite database to MessageClient.java to store received messages.
Just create the database schema and connection code.
Don't implement saving yet.
Write a test that verifies we can connect to the database.
```

**Prompt 2:**
```
Now add a method to save a message to the database.
Write ONE test for the save method first, then implement it.
```

**Prompt 3:**
```
Add a method to retrieve the last 5 messages from the database.
Write a test first, then implement.
```

**Prompt 4:**
```
Update MessageClient to save each message it receives and display the history.
```

**Key lesson:** Breaking down complex features into testable steps.

**Success criteria:**
- Database created
- Messages saved after each API call
- Can retrieve and display message history
- Tests pass for each component

---

## Exercise 6: Send Messages to Others (Advanced, 15 minutes)

**Goal:** Implement POST functionality.

**Prompt:**
```
Add functionality to send a positive message to another user.

Build this incrementally:
1. First, add a method that makes a POST request to /v1/message
2. Write a test for it (you can mock the HTTP response)
3. Add a simple command-line option to send vs receive messages
4. Test it by sending a message to "Bob"
```

**Then test:**
```bash
# Send a message
java MessageClient send Bob "Great work on your code!"

# Check Bob received it
curl "https://happy.industrial-linguistics.com/v1/messages?recipient=Bob"
```

**Success criteria:**
- Can send messages via POST
- Messages appear in recipient's message list
- Test coverage for the new feature

---

## Exercise 7: Error Handling (10 minutes)

**Goal:** Make the code robust.

**Prompt:**
```
Add proper error handling to MessageClient:
1. Handle network errors gracefully
2. Handle invalid API responses
3. Handle database errors
4. Print user-friendly error messages

Don't refactor the whole program - just add error handling where needed.
```

**Test by:**
- Disconnecting network
- Passing invalid name
- Corrupting database file

**Success criteria:**
- Program doesn't crash on errors
- User sees helpful error messages
- Original functionality still works

---

## Exercise 8: Code Review (5 minutes)

**Goal:** Use Claude to review your work.

**Prompt:**
```
Review MessageClient.java for:
1. Code quality and best practices
2. Missing error handling
3. Potential bugs
4. Test coverage gaps

Give me a prioritized list of improvements.
```

**Then pick ONE improvement:**
```
Fix the [highest priority issue from review].
Make only that change - don't refactor everything.
```

**Success criteria:**
- Receive actionable feedback
- Implement one improvement
- Understand why it matters

---

## Bonus Exercises

### Bonus 1: Web Interface with Message History
Add the local storage from Exercise 5 to the web interface from Exercise 3.

### Bonus 2: Send Messages via Web UI
Add a form to send messages to other users on the web interface.

### Bonus 3: Auto-refresh
Make the web interface poll for new messages every 30 seconds.

### Bonus 4: Message Categories
The API returns messages in categories (achievement, encouragement, persistence).
Display them with different colors or icons.

---

## Tips for Working with Claude

### DO:
✓ Give specific, focused prompts
✓ Build incrementally (database → save → retrieve)
✓ Write tests before implementation
✓ Tell Claude what NOT to change
✓ Use phrases like "ONLY add...", "Don't refactor..."

### DON'T:
✗ Ask Claude to "build the whole thing"
✗ Let it generate 10+ files at once
✗ Accept massive refactorings when you asked for a small change
✗ Skip testing

### Recovery Techniques:
- If Claude over-engineers: `/rewind` and try a more specific prompt
- If you're stuck: Start a new Claude session for that subtask
- If tests fail: Ask Claude to fix ONLY the failing test
- If output is wrong: "The [X] is wrong - fix only that, don't change [Y]"

---

## Monitoring Your Progress

The instructor can see your activity at:
```bash
happywatch -mode students
```

Make sure your code is actually hitting the API! The instructor uses this to know:
- Who's working on which exercise
- Who might be stuck
- When to move to the next topic

Include the session ID in all requests so activity is properly tracked.

---

## Common Issues

**"My program compiles but nothing happens"**
- Add print statements to debug
- Check you're actually calling the API
- Verify the URL is correct

**"I get JSON parsing errors"**
- Print the raw response to see what you received
- Check your parameter names match the API docs
- Try the API with `curl` first

**"Claude generated way too much code"**
- Use `/rewind` to go back
- Give a more specific prompt
- Say "Just add X, don't change anything else"

**"My tests aren't running"**
- Verify JUnit is in your classpath
- Check test method names start with `test`
- Look at the test output for specific errors

---

## What You're Learning

Beyond the API and Java:

1. **Prompt Engineering**: How to control AI scope with specific language
2. **Incremental Development**: Building features step-by-step
3. **TDD with AI**: Writing tests before implementation
4. **Error Recovery**: Using `/rewind` and new sessions strategically
5. **Code Review**: Using AI as a reviewer, not just a writer

These skills transfer to any AI coding tool and any programming language.
