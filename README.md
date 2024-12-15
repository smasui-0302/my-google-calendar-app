# my-google-calendar-app

Oauth認証を勉強するために実装した、Google Calender連携を行い予定を表示するサンプルコード

## 使い方

1. Google Cloud Consoleで認証情報を取得
    - Google Cloud Consoleにアクセスし、プロジェクトを作成
    - OAuth同意画面を設定
    - 認証情報を作成し、OAuth 2.0 クライアントIDを取得
    - 認証情報をJSONファイルとしてダウンロード

2. 認証情報の配置
    - ダウンロードしたJSONファイルを `.credentials/calendar_credentials.json` として配置
    ```
    mv {ダウンロードしたファイル} .credentials/calendar_credentials.json
    ```

3. アプリケーションのビルドと実行
    ```
    # ビルド
    make build

    # 実行
    make run
    ```

4. アプリケーションへのアクセス
    - ブラウザで http://localhost:8080 にアクセス
    - 「Login with Google Calendar」ボタンをクリックしてGoogleアカウントでログイン
    - 承認後、今後1ヶ月分の予定が表示されます
