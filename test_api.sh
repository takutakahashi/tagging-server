#!/bin/bash

# サーバー設定
BASE_URL="https://taggingservice-production-26607266306.asia-northeast1.run.app"

# 新しい鍵を生成
echo "=== Generating New Key ==="
NEW_KEY=$(curl -s -X GET "$BASE_URL/new-key" | tr -d '"')
if [ -z "$NEW_KEY" ]; then
    echo "Failed to generate new key."
    exit 1
fi
echo "Generated Key: $NEW_KEY"

# ヘッダー設定
AUTH_HEADER="Authorization: v0hm85H9DdySTu3B-DkrgS3ci34fdWVx"

# テスト用データ
TARGET="myTarget"
TAG="myTag"

# タグの追加
echo "=== Adding Tag ==="
curl -s -X POST "$BASE_URL/add-tag" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{\"target\": \"$TARGET\", \"tag\": \"$TAG\"}"
echo -e "\nTag added successfully."

# タグの取得
echo "=== Getting Tags for Target ==="
curl -s -X GET "$BASE_URL/get-tags?target=$TARGET" \
    -H "$AUTH_HEADER"
echo -e "\nTags fetched successfully."

# ターゲットの取得
echo "=== Getting Targets for Tag ==="
curl -s -X GET "$BASE_URL/get-targets?tag=$TAG" \
    -H "$AUTH_HEADER"
echo -e "\nTargets fetched successfully."

# Likeを追加
echo "=== Liking Target ==="
curl -s -X POST "$BASE_URL/like-target" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{\"target\": \"$TARGET\"}"
echo -e "\nTarget liked successfully."

# Like数の取得
echo "=== Getting Like Count for Target ==="
curl -s -X GET "$BASE_URL/get-likes?target=$TARGET" \
    -H "$AUTH_HEADER"
echo -e "\nLike count fetched successfully."

