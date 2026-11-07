#!/usr/bin/env bash
# 由 proto/repgame/v1/*.proto（按模块拆分）生成 Go 与 C# 代码。
#
# 依赖：
#   protoc               3.29.x（NuGet 包 Google.Protobuf.Tools 自带，或官方发布版）
#   protoc-gen-go        go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.5
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

"$PROTOC" -I "$ROOT/proto"   --go_out="$GO_OUT" --go_opt=paths=source_relative "${GO_MAPS[@]}"   --csharp_out="$CS_OUT"   "${FILES[@]}"

# paths=source_relative 会保留 repgame/v1 目录，拍平到 pb/ 下
if [ -d "$GO_OUT/repgame/v1" ]; then
  mv "$GO_OUT"/repgame/v1/*.pb.go "$GO_OUT"/
  rm -rf "$GO_OUT/repgame"
fi
