# DEVELOPMENT.md

## テストに関して
1. テスト結果は `test/logs` `test/reports` 以下に出力する
2. テストには `test/env/llm-info.yaml` を使用する. 使うURLやキーが記述されている。
3. テストの最後には `Bash(./bin/llm-info --url https://openrouter.ai/api)` をして、通常動作を必ず確認する。

