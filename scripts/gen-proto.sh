#!/usr/bin/env bash
# 由 proto/repgame/v1/*.proto 生成 Go 与 C# 代码。
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

"$PROTOC" -I "$ROOT/proto" \
  --go_out="$GO_OUT" --go_opt=paths=source_relative \
  --go_opt=Mrepgame/v1/game.proto="GoServer/tcpgameserver/pb;pb" \
  --csharp_out="$CS_OUT" \
  "$ROOT/proto/repgame/v1/game.proto"

# paths=source_relative 会保留 repgame/v1 目录，拍平到 pb/ 下
if [ -f "$GO_OUT/repgame/v1/game.pb.go" ]; then
  mv "$GO_OUT/repgame/v1/game.pb.go" "$GO_OUT/game.pb.go"
  rm -rf "$GO_OUT/repgame"
fi

echo "Generated: $GO_OUT/game.pb.go"
echo "Generated: $CS_OUT/Game.cs"
