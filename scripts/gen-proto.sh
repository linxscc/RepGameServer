#!/usr/bin/env bash
# 由 proto/repgame/v1/*.proto（按模块拆分）生成 Go 与 C# 代码。
#
# 依赖：
#   protoc               3.29.x（NuGet 包 Google.Protobuf.Tools 自带，或官方发布版）
#   protoc-gen-go        go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5
#   python3 + protobuf   pip install protobuf（gen-wire-tags.py 读取描述符用）
#
# 用法：
#   PROTOC=/path/to/protoc ./scripts/gen-proto.sh [前端 RepGame 项目路径]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CLIENT="${1:-$ROOT/../RepGame}"
PROTOC="${PROTOC:-protoc}"

GO_OUT="$ROOT/GoServer/tcpgameserver/pb"
CS_OUT="$CLIENT/Assets/Scripts/Protocol/Generated"
mkdir -p "$GO_OUT" "$CS_OUT"

# 按模块拆分的协议文件，同一个 package（repgame.v1），线上格式与拆分前一致
PROTOS=(envelope common auth lobby match_action match_config match_state)

GO_MAPS=()
FILES=()
for p in "${PROTOS[@]}"; do
  GO_MAPS+=(--go_opt=Mrepgame/v1/$p.proto="GoServer/tcpgameserver/pb;pb")
  FILES+=("$ROOT/proto/repgame/v1/$p.proto")
done

# 拆分前的单文件产物
rm -f "$GO_OUT/game.pb.go" "$CS_OUT/Game.cs" "$CS_OUT/Game.cs.meta"

"$PROTOC" -I "$ROOT/proto" \
  --go_out="$GO_OUT" --go_opt=paths=source_relative "${GO_MAPS[@]}" \
  --csharp_out="$CS_OUT" \
  "${FILES[@]}"

# paths=source_relative 会保留 repgame/v1 目录，拍平到 pb/ 下
if [ -d "$GO_OUT/repgame/v1" ]; then
  mv "$GO_OUT"/repgame/v1/*.pb.go "$GO_OUT"/
  rm -rf "$GO_OUT/repgame"
fi

# 选第一个能 import protobuf 的解释器（Windows 上的 python3 可能是应用商店占位程序）
if [ -z "${PYTHON:-}" ]; then
  for p in python3 python; do
    if command -v "$p" >/dev/null && "$p" -c "import google.protobuf" 2>/dev/null; then PYTHON="$p"; break; fi
  done
fi
if [ -z "${PYTHON:-}" ]; then
  echo "gen-proto.sh: 需要装有 protobuf 的 python（pip install protobuf），或用 PYTHON=... 指定" >&2
  exit 1
fi

# 描述符数据（schema）移到单独的文件：C# <Module>Descriptor.cs，Go <module>_desc.pb.go
"$PYTHON" "$ROOT/scripts/split-descriptors.py" "$GO_OUT" "$CS_OUT" "${PROTOS[@]}"

# C# 解析代码里的 case 数字换成具名的线上标签枚举（<Module>Tags.cs），见 gen-wire-tags.py
DESC="$(mktemp)"
trap 'rm -f "$DESC"' EXIT
"$PROTOC" -I "$ROOT/proto" --include_source_info --descriptor_set_out="$DESC" "${FILES[@]}"
"$PYTHON" "$ROOT/scripts/gen-wire-tags.py" "$DESC" "$CS_OUT"
