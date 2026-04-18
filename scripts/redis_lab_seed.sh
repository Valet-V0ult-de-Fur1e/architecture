#!/usr/bin/env sh
set -eu

COMPOSE="docker compose -f docker/docker-compose.yml -f docker/docker-compose.debug.yml"
REDIS_CLI="$COMPOSE exec -T redis redis-cli"

GROUP="${GROUP:-your_group}"
LIST_NUM="${LIST_NUM:-00}"
FIO="${FIO:-Your Full Name}"
AGE="${AGE:-20}"
EMAIL="${EMAIL:-your.email@misis.edu}"

KEY_BASE="student:${GROUP}:${LIST_NUM}"

# String
$REDIS_CLI SET "$KEY_BASE" "$FIO" >/dev/null

# Hash
$REDIS_CLI HSET "${KEY_BASE}:info" name "$FIO" age "$AGE" email "$EMAIL" >/dev/null

# List (more than one item)
$REDIS_CLI DEL "${KEY_BASE}:timetable" >/dev/null
$REDIS_CLI RPUSH "${KEY_BASE}:timetable" "Math" "Physics" "Databases" >/dev/null

# Set
$REDIS_CLI DEL "${KEY_BASE}:skills" >/dev/null
$REDIS_CLI SADD "${KEY_BASE}:skills" "Docker" "Go" "PostgreSQL" "Redis" >/dev/null

# ZSet
$REDIS_CLI DEL "${KEY_BASE}:tasks_w_priority" >/dev/null
$REDIS_CLI ZADD "${KEY_BASE}:tasks_w_priority" 100 "Finish lab 1" 150 "Finish lab 2" 200 "Prepare defense" >/dev/null

echo "Seeded lab keys for ${KEY_BASE}"
