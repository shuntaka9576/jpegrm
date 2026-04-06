# jpegrm

JPEG ファイルを EXIF 撮影日時に基づいてリネームする CLI ツール。

## 概要

JPEG ファイルの EXIF メタデータから撮影日時を読み取り、統一されたファイル名にリネームする。Go で実装しており、macOS / Windows / Linux のバイナリをクロスコンパイルで生成できる。

## リネーム形式

```
YYYY_MM_DD_HHMM_NN.jpg
```

- `YYYY` - 年 (4桁)
- `MM` - 月 (2桁ゼロ埋め)
- `DD` - 日 (2桁ゼロ埋め)
- `HHMM` - 時分 (4桁、秒は含まない)
- `NN` - 連番 (2桁ゼロ埋め、`_00` から開始)
- 拡張子は常に小文字 `.jpg` に正規化

例: `IMG_0001.jpg` → `2024_03_15_1430_00.jpg`

## EXIF タグ優先順位

以下の順に最初に見つかったタグを使用する。

1. `DateTimeOriginal` - シャッターを切った瞬間
2. `DateTimeDigitized` - デジタル化した日時
3. `DateTime` - 最終更新日時（フォールバック）

タイムゾーン情報は無視し、EXIF に記録されたローカル時刻をそのまま使用する。

## 対象ファイル

- 拡張子: `.jpg`, `.jpeg`, `.JPG`, `.JPEG` (大文字小文字不問)
- EXIF データが存在しないファイルはスキップ
- EXIF に日時タグが存在しないファイルはスキップ

## 重複ファイル名の処理

- リネーム前に全ファイルの計画を構築してから実行する
- ファイルはパスの辞書順でソートし、決定論的な結果を保証する
- 同一タイムスタンプのファイルには `_00`, `_01`, `_02` ... と連番を付与する
- ディスク上の既存ファイルとの衝突もチェックし、衝突する場合は連番をインクリメントする
- 既にリネーム済み（ファイル名が変わらない）の場合はスキップする

## CLI オプション

```
Usage: jpegrm [options] [directory] [pattern]

Arguments:
  directory  対象ディレクトリ (省略時: カレントディレクトリ)
  pattern    ファイル名フィルタ (glob形式, 省略時: 全JPEGファイル)

Options:
  -n    プレビューのみ（実際にはリネームしない）
  -r    サブディレクトリも走査
  -v    スキップしたファイル等の詳細表示
```

- `directory` は省略時カレントディレクトリ
- `-r` なしの場合は指定ディレクトリ直下のみ走査
- `pattern` は glob 形式でファイル名をフィルタする（大文字小文字不問）
  - `*.*` または省略: 全JPEG ファイル
  - `DSC*`: DSC で始まるファイル
  - `DSC1234`: `DSC1234.jpg` のみ（ワイルドカード・拡張子なしの場合は自動で `.*` を付与）

## エラーハンドリング

| 状況 | 動作 |
|---|---|
| EXIF なし / 日時タグなし / 日時形式不正 | スキップ (`-v` で表示) |
| ファイルを開けない | スキップ (`-v` で表示) |
| ディレクトリが存在しない | stderr にエラー出力、exit code 1 |
| JPEG ファイルが見つからない | `No JPEG files found.` を表示して正常終了 |
| リネーム対象がない | `No files to rename.` を表示して正常終了 |
| リネーム時の OS エラー | stderr にエラー出力、残りのファイルは続行 |

## 出力形式

通常モード:
```
IMG_0001.jpg -> 2024_03_15_1430_00.jpg
IMG_0002.jpg -> 2024_03_15_1430_01.jpg

Renamed 2 files.
```

ドライランモード (`-n`):
```
[DRY RUN] IMG_0001.jpg -> 2024_03_15_1430_00.jpg
[DRY RUN] IMG_0002.jpg -> 2024_03_15_1430_01.jpg

Dry run complete. 2 files would be renamed.
```

同一ディレクトリ内のリネームはファイル名のみ表示。再帰モードでディレクトリが異なる場合はフルパスを表示する。

## ビルド

```bash
# macOS (Apple Silicon)
go build -o jpegrm .

# Windows
GOOS=windows GOARCH=amd64 go build -o jpegrm.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o jpegrm .
```

## Makefile コマンド

| コマンド | 説明 |
|---|---|
| `make build` | 現在の OS 向けにビルド |
| `make build-all` | 4プラットフォーム一括ビルド (`dist/` に出力) |
| `make package-windows` | Windows バイナリ + README を zip にパッケージング |
| `make installer` | `package-windows` 実行後、Inno Setup でのインストーラー作成手順を表示 |
| `make clean` | ビルド成果物を削除 |

## Windows インストーラー (Inno Setup)

`installer.iss` は Inno Setup 用のスクリプト。インストーラーは macOS からは作成できないため、Windows 側でビルドする。

### 手順

1. macOS 側で `make package-windows` を実行し `dist/jpegrm.exe` を生成
2. リポジトリを Windows に持っていく
3. [Inno Setup](https://jrsoftware.org/isdl.php) をインストール
4. コマンドプロンプトで `iscc installer.iss` を実行
5. `dist/jpegrm-setup.exe` が生成される

### インストーラーの動作

- `C:\Program Files\jpegrm\` に `jpegrm.exe` と `README-windows.txt` を配置
- ユーザーの PATH 環境変数に自動追加
- 管理者権限不要 (ユーザースコープでインストール)

## 依存ライブラリ

- `github.com/rwcarlsen/goexif` - EXIF 読み取り
