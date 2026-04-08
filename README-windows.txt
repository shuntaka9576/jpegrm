============================================================
jpegrm - JPEG EXIF撮影日時リネームツール (Windows版)
============================================================

■ 概要
JPEG ファイルの EXIF 撮影日時を読み取り、以下の形式にリネームします。

  元ファイル名.jpg → YYYY_MM_DD_HHMM_NN.jpg

  例: IMG_0001.jpg → 2024_03_15_1430_00.jpg

同じ分に複数ファイルがある場合は _00, _01, _02 ... と連番が付きます。


■ セットアップ
1. jpegrm-setup.exe を実行してインストール（管理者権限不要）
2. jpegrm.exe が C:\Program Files\jpegrm\ に配置され、PATH に自動追加されます
3. インストール後、コマンドプロンプトまたは PowerShell から使えます


■ 使い方

  jpegrm [options] [path]

コマンドプロンプト (cmd) または PowerShell を開いて実行します。

  ドライラン（プレビュー、実際にはリネームしない）:
    jpegrm -n C:\Users\gaola\OneDrive\画像\G16インポート

  全JPEGをリネーム:
    jpegrm C:\Users\gaola\OneDrive\画像\G16インポート

  *.* で全JPEG（上と同じ）:
    jpegrm C:\Users\gaola\OneDrive\画像\G16インポート\*.*
    ※ 依頼仕様では G16インポート*.* でしたが、フォルダ名との境界を判別できないため \ 区切りに変更

  DSC1234.jpg だけリネーム:
    jpegrm C:\Users\gaola\OneDrive\画像\G16インポート\DSC1234

  DSC で始まるファイルだけリネーム:
    jpegrm C:\Users\gaola\OneDrive\画像\G16インポート\DSC*

  サブフォルダも含めて処理:
    jpegrm -r C:\Users\gaola\OneDrive\画像\G16インポート

  詳細表示（スキップしたファイルも表示）:
    jpegrm -v C:\Users\gaola\OneDrive\画像\G16インポート

  カメラ識別サフィックス付きでリネーム:
    jpegrm -s EOS C:\Users\gaola\OneDrive\画像\G16インポート
    → 2024_03_15_1430_00_EOS.jpg のようにリネーム

  EXIFのカメラ機種名を自動でサフィックスに使う:
    jpegrm -s %model C:\Users\gaola\OneDrive\画像\G16インポート
    → 2024_03_15_1430_00_ZV-E10M2.jpg のようにリネーム

  メーカー名と機種名の組み合わせ:
    jpegrm -s %make_%model C:\Users\gaola\OneDrive\画像\G16インポート
    → 2024_03_15_1430_00_SONY_ZV-E10M2.jpg のようにリネーム


■ オプション一覧
  -n              プレビューのみ（実際にはリネームしない）
  -r              サブフォルダも走査
  -s / -suffix    サフィックス（%make=メーカー, %model=機種名 でEXIF値に置換）
  -v              スキップしたファイル等の詳細表示


■ 対象ファイル
  .jpg / .jpeg / .JPG / .JPEG（大文字小文字不問）


■ 注意事項
  - EXIF データがないファイルはスキップされます
  - リネーム前に必ず -n オプションでプレビューすることを推奨します
  - 元に戻す機能はありません。心配な場合は事前にバックアップしてください
