#!/usr/bin/env sh
set -eu

COMPOSE="docker compose -f docker/docker-compose.yml -f docker/docker-compose.debug.yml"
REDIS_CLI="$COMPOSE exec -T redis redis-cli"

CURSOR=0
while :; do
  SCAN_OUT="$($REDIS_CLI SCAN "$CURSOR")"
  CURSOR="$(printf '%s\n' "$SCAN_OUT" | sed -n '1p')"
  KEYS="$(printf '%s\n' "$SCAN_OUT" | sed -n '2,$p')"

  for key in $KEYS; do
    type="$($REDIS_CLI TYPE "$key")"
    echo "Key: $key, Type: $type"
    case "$type" in
      string)
        $REDIS_CLI GET "$key"
        ;;
      hash)
        $REDIS_CLI HGETALL "$key"
        ;;
      list)
        $REDIS_CLI LRANGE "$key" 0 -1
        ;;
      set)
        $REDIS_CLI SMEMBERS "$key"
        ;;
      zset)
        $REDIS_CLI ZRANGE "$key" 0 -1 WITHSCORES
        ;;
    esac
  done

  if [ "$CURSOR" = "0" ]; then
    break
  fi
done
