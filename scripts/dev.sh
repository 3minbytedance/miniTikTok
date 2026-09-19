#!/usr/bin/env bash
# 宿主机开发环境一键管理脚本（基础设施 MySQL/Redis/Kafka/Mongo/etcd 需已在本机运行）。
#
# 用法:
#   ./scripts/dev.sh start              编译并后台启动全部服务
#   ./scripts/dev.sh stop               停止全部服务
#   ./scripts/dev.sh restart            重启全部服务
#   ./scripts/dev.sh status             查看运行状态
#   ./scripts/dev.sh logs [服务名]      查看全部或指定服务日志（social|video|message|api）
#
# 服务端口:
#   api      :8080   HTTP 网关（/douyin/*、/metrics、/static/*）
#   social   :4002   用户注册登录 + 关系链
#   video    :4003   视频流/发布 + 评论 + 点赞
#   message  :4005   私信消息

set -euo pipefail

# 切到仓库根目录，保证 config/app.yaml 等相对路径可用
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

BIN_DIR="build/bin"
RUN_DIR="run"
LOG_DIR="log"
# 启动顺序：先 RPC 服务，最后网关
SERVICES=(social video message api)

mkdir -p "$RUN_DIR" "$LOG_DIR"

is_running() {
  local svc=$1 pidfile="$RUN_DIR/$svc.pid"
  [[ -f "$pidfile" ]] || return 1
  local pid
  pid=$(cat "$pidfile" 2>/dev/null || true)
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

start_one() {
  local svc=$1
  if is_running "$svc"; then
    echo "[skip] $svc 已在运行 (pid $(cat "$RUN_DIR/$svc.pid"))"
    return 0
  fi
  nohup "$BIN_DIR/$svc" >> "$LOG_DIR/$svc.out.log" 2>&1 &
  echo $! > "$RUN_DIR/$svc.pid"
  echo "[ok]   启动 $svc (pid $!)"
}

stop_one() {
  local svc=$1
  if ! is_running "$svc"; then
    rm -f "$RUN_DIR/$svc.pid"
    echo "[skip] $svc 未运行"
    return 0
  fi
  local pid
  pid=$(cat "$RUN_DIR/$svc.pid")
  kill "$pid" 2>/dev/null || true
  for _ in $(seq 1 10); do
    kill -0 "$pid" 2>/dev/null || break
    sleep 0.5
  done
  if kill -0 "$pid" 2>/dev/null; then
    kill -9 "$pid" 2>/dev/null || true
  fi
  rm -f "$RUN_DIR/$svc.pid"
  echo "[ok]   停止 $svc"
}

cmd_start() {
  echo "==> 编译服务..."
  make build-all
  echo "==> 启动服务..."
  for svc in "${SERVICES[@]}"; do start_one "$svc"; done
  echo "==> 完成。api 网关: http://127.0.0.1:8080"
}

cmd_stop() {
  echo "==> 停止服务..."
  for ((i=${#SERVICES[@]}-1; i>=0; i--)); do stop_one "${SERVICES[$i]}"; done
}

cmd_status() {
  printf "%-10s %-8s %-8s\n" "SERVICE" "PORT" "STATUS"
  declare -A ports=([social]=4002 [video]=4003 [message]=4005 [api]=8080)
  for svc in "${SERVICES[@]}"; do
    if is_running "$svc"; then
      printf "%-10s %-8s %-8s\n" "$svc" "${ports[$svc]}" "running (pid $(cat "$RUN_DIR/$svc.pid"))"
    else
      printf "%-10s %-8s %-8s\n" "$svc" "${ports[$svc]}" "stopped"
    fi
  done
}

cmd_logs() {
  local svc=${1:-}
  if [[ -n "$svc" ]]; then
    tail -f "$LOG_DIR/$svc.out.log"
  else
    tail -f "$LOG_DIR"/*.out.log
  fi
}

case "${1:-}" in
  start)   cmd_start ;;
  stop)    cmd_stop ;;
  restart) cmd_stop; cmd_start ;;
  status)  cmd_status ;;
  logs)    cmd_logs "${2:-}" ;;
  *)
    grep '^#' "$0" | sed 's/^# \{0,1\}//'
    exit 1
    ;;
esac
