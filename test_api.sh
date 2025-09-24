#!/bin/bash

# Quivra Backend API テストスクリプト

BASE_URL="http://localhost:8080/api"

echo "=== Quivra Backend API テスト ==="

# 1. ルーム作成
echo "1. ルーム作成テスト"
ROOM_RESPONSE=$(curl -s -X POST "$BASE_URL/rooms" \
  -H "Content-Type: application/json" \
  -d '{"name": "テストルーム"}')

echo "ルーム作成レスポンス: $ROOM_RESPONSE"

# ルームIDを抽出
ROOM_ID=$(echo $ROOM_RESPONSE | grep -o '"roomId":"[^"]*"' | cut -d'"' -f4)
echo "作成されたルームID: $ROOM_ID"

if [ -z "$ROOM_ID" ]; then
  echo "エラー: ルームIDが取得できませんでした"
  exit 1
fi

# 2. ルーム情報取得
echo -e "\n2. ルーム情報取得テスト"
curl -s -X GET "$BASE_URL/rooms/$ROOM_ID" | jq '.'

# 3. 問題作成
echo -e "\n3. 問題作成テスト"
curl -s -X POST "$BASE_URL/questions" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Go言語の作者は誰ですか？",
    "answer": "ロブ・パイク",
    "category": "programming",
    "difficulty": "medium"
  }' | jq '.'

# 4. 問題一覧取得
echo -e "\n4. 問題一覧取得テスト"
curl -s -X GET "$BASE_URL/questions" | jq '.'

# 5. ルーム参加（HTTP API経由）
echo -e "\n5. ルーム参加テスト"
curl -s -X POST "$BASE_URL/rooms/join" \
  -H "Content-Type: application/json" \
  -d "{
    \"roomId\": \"$ROOM_ID\",
    \"playerName\": \"テストプレイヤー1\"
  }" | jq '.'

echo -e "\n=== テスト完了 ==="
echo "WebSocketテストは手動で行ってください:"
echo "ws://localhost:8080/ws"
