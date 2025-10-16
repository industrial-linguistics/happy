#!/bin/sh
# Test script for happy-api

BASE_URL="${BASE_URL:-https://happy.industrial-linguistics.com/v1}"

echo "Testing happy-api at $BASE_URL"
echo "========================================"
echo ""

# Test 1: Health check
echo "Test 1: GET /status"
echo "-------------------"
curl -s "$BASE_URL/status" | jq . || echo "FAILED"
echo ""

# Test 2: Get a message
echo "Test 2: GET /message?name=TestUser"
echo "-----------------------------------"
curl -s "$BASE_URL/message?name=TestUser&session_id=test_session" | jq . || echo "FAILED"
echo ""

# Test 3: Get another message (check sequence)
echo "Test 3: GET /message again (check sequence)"
echo "-------------------------------------------"
curl -s "$BASE_URL/message?name=TestUser&session_id=test_session" | jq . || echo "FAILED"
echo ""

# Test 4: Post a message
echo "Test 4: POST /message"
echo "---------------------"
curl -s -X POST "$BASE_URL/message" \
  -H "Content-Type: application/json" \
  -d '{
    "from": "Alice",
    "to": "Bob",
    "message": "Great work on that feature!",
    "session_id": "test_session"
  }' | jq . || echo "FAILED"
echo ""

# Test 5: Get messages for recipient
echo "Test 5: GET /messages?recipient=Bob"
echo "------------------------------------"
curl -s "$BASE_URL/messages?recipient=Bob&limit=10" | jq . || echo "FAILED"
echo ""

# Test 6: Error handling - missing name
echo "Test 6: Error handling (missing name)"
echo "--------------------------------------"
curl -s "$BASE_URL/message" | jq . || echo "FAILED"
echo ""

# Test 7: Error handling - negative message
echo "Test 7: Error handling (negative message)"
echo "-----------------------------------------"
curl -s -X POST "$BASE_URL/message" \
  -H "Content-Type: application/json" \
  -d '{
    "from": "Eve",
    "to": "Carol",
    "message": "This is terrible and awful",
    "session_id": "test_session"
  }' | jq . || echo "FAILED"
echo ""

echo "========================================"
echo "Tests complete!"
echo ""
echo "To view activity:"
echo "  logwatch -mode summary"
