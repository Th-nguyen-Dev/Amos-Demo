#!/bin/bash
set -e

echo "=== Comprehensive API Test ==="
echo ""

# Test 1: Create multiple Q&A pairs
echo "1. Creating Q&A pairs..."
ID1=$(curl -s -X POST http://localhost:8080/api/qa-pairs \
  -H "Content-Type: application/json" \
  -d '{"question": "What are your business hours?", "answer": "We are open 9 AM - 5 PM EST."}' | jq -r '.qa_pair.id')
echo "   Created Q&A 1: $ID1"

ID2=$(curl -s -X POST http://localhost:8080/api/qa-pairs \
  -H "Content-Type: application/json" \
  -d '{"question": "Do you offer gift wrapping?", "answer": "Yes, gift wrapping is available for $5."}' | jq -r '.qa_pair.id')
echo "   Created Q&A 2: $ID2"

# Test 2: List Q&A pairs
echo ""
echo "2. Listing Q&A pairs..."
COUNT=$(curl -s http://localhost:8080/api/qa-pairs | jq '.data | length')
echo "   Total Q&A pairs: $COUNT"

# Test 3: Search
echo ""
echo "3. Testing search..."
RESULTS=$(curl -s 'http://localhost:8080/api/qa-pairs?search=gift' | jq '.data | length')
echo "   Search results for 'gift': $RESULTS"

# Test 4: Create conversation
echo ""
echo "4. Creating conversation..."
CONV_ID=$(curl -s -X POST http://localhost:8080/api/conversations \
  -H "Content-Type: application/json" \
  -d '{"title": "Product Inquiry"}' | jq -r '.conversation.id')
echo "   Created conversation: $CONV_ID"

# Test 5: Add messages
echo ""
echo "5. Adding messages to conversation..."
MSG1=$(curl -s -X POST http://localhost:8080/api/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "Do you have gift wrapping?",
    "raw_message": {"role": "user", "content": "Do you have gift wrapping?"}
  }' | jq -r '.message.id')
echo "   Added user message: $MSG1"

MSG2=$(curl -s -X POST http://localhost:8080/api/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d '{
    "role": "assistant",
    "content": null,
    "raw_message": {
      "role": "assistant",
      "content": null,
      "tool_calls": [{
        "id": "call_search_123",
        "type": "function",
        "function": {"name": "search_qa", "arguments": "{\"query\": \"gift wrapping\"}"}
      }]
    }
  }' | jq -r '.message.id')
echo "   Added assistant message with tool call: $MSG2"

MSG3=$(curl -s -X POST http://localhost:8080/api/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d "{
    \"role\": \"tool\",
    \"tool_call_id\": \"call_search_123\",
    \"content\": \"Yes, gift wrapping is available for \$5.\",
    \"raw_message\": {
      \"role\": \"tool\",
      \"tool_call_id\": \"call_search_123\",
      \"content\": \"Yes, gift wrapping is available for \$5.\"
    }
  }" | jq -r '.message.id')
echo "   Added tool response: $MSG3"

# Test 6: Get messages
echo ""
echo "6. Retrieving conversation messages..."
MSG_COUNT=$(curl -s http://localhost:8080/api/conversations/$CONV_ID/messages | jq '.data | length')
echo "   Messages in conversation: $MSG_COUNT"

# Test 7: Verify in database
echo ""
echo "7. Verifying data in database..."
sqlite3 smart_discovery.db "SELECT COUNT(*) FROM qa_pairs;" | sed 's/^/   Q&A pairs in DB: /'
sqlite3 smart_discovery.db "SELECT COUNT(*) FROM conversations;" | sed 's/^/   Conversations in DB: /'
sqlite3 smart_discovery.db "SELECT COUNT(*) FROM messages;" | sed 's/^/   Messages in DB: /'

# Test 8: Delete Q&A
echo ""
echo "8. Testing deletion..."
curl -s -X DELETE http://localhost:8080/api/qa-pairs/$ID1 | jq -r '.success' | sed 's/^/   Deleted Q&A: /'
REMAINING=$(curl -s http://localhost:8080/api/qa-pairs | jq '.data | length')
echo "   Remaining Q&A pairs: $REMAINING"

# Test 9: Tool API
echo ""
echo "9. Testing Tool API..."
SEARCH_RESULTS=$(curl -s -X POST http://localhost:8080/tools/search-qa \
  -H "Content-Type: application/json" \
  -d '{"query": "gift", "limit": 10}' | jq '.count')
echo "   Tool search results: $SEARCH_RESULTS"

echo ""
echo "=== All API Tests Completed Successfully! ==="
