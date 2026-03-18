# cocotola-auth

認証・認可を管理するマイクロサービスです

- Webアプリケーションです。

## 構成

複数の組織を持ちます。組織はテナントであり、このサービスはマルチテナントをサポートします。

組織にはグループとユーザーが所属します。
グループには0名以上のユーザーが所属します。
グループには0個以上のグループが所属します。グループの所属階層はループできません。
ユーザーは0以上のグループに所属することができます。

## 認証方式

Cookie認証とToken認証の2種類をサポートする。

### Cookie認証 (Webブラウザ向け)

- SessionToken (opaque token) を使用。JWTではない
- RefreshTokenなし
- 操作ごとにExpiresAtを30分延長する (sliding window)
- ただしCreatedAtから最大24時間を超えない (absolute timeout)
- Cookie名: `session_token`
- Cookie属性: HttpOnly=true, Secure/SameSite/Pathは環境変数で設定可能
- DBに保存する。ユーザーあたり最新10件を保持する (ホワイトリスト)

### Token認証 (API/モバイル向け)

- AccessToken (JWT) + RefreshToken (opaque token) を使用
- AccessTokenの有効期限: デフォルト1時間 (環境変数で設定可能)
- RefreshTokenの有効期限: デフォルト30日 (環境変数で設定可能)
- AccessTokenの自動延長はしない
- JWT署名検証 + インメモリキャッシュでホワイトリスト照合
- DBにホワイトリスト保存。ユーザーあたりRefreshToken最新10件、AccessToken最新10件を保持
- RefreshTokenまたはAccessTokenを指定してRevokeできる
- RefreshTokenをRevokeすると紐づくAccessTokenも全てRevokeされる

## ドメイン

### Organization (集約)

組織（テナント）。ユーザーとグループの有効数の上限を持つ。

| フィールド | 型 | 説明 |
|---|---|---|
| ID | int | 識別子 |
| Name | string | 組織名 (一意) |
| MaxActiveUsers | int | 有効にできるユーザー数の上限 |
| MaxActiveGroups | int | 有効にできるグループ数の上限 |

### AppUser (集約)

組織に所属するユーザー。1つの組織にのみ所属できる。

| フィールド | 型 | 説明 |
|---|---|---|
| ID | int | 識別子 |
| OrganizationID | int | 所属する組織のID |
| LoginID | string | ログインID |
| Enabled | bool | 有効/無効 |

### Group (集約)

組織に所属するグループ。1つの組織にのみ所属できる。

| フィールド | 型 | 説明 |
|---|---|---|
| ID | int | 識別子 |
| OrganizationID | int | 所属する組織のID |
| Name | string | グループ名 |
| Enabled | bool | 有効/無効 |

### ActiveUserList (集約)

組織ごとの有効なユーザーIDの集合。ユーザー有効化時にOrganization.MaxActiveUsersと比較して上限をチェックする。

| フィールド | 型 | 説明 |
|---|---|---|
| OrganizationID | int | 組織ID |
| Entries | []int | 有効なユーザーIDの一覧 |

- ユーザー有効化時: Entriesに追加。len(Entries) >= MaxActiveUsers の場合はエラー
- ユーザー無効化時: Entriesから除外

### ActiveGroupList (集約)

組織ごとの有効なグループIDの集合。ActiveUserListと同一パターン。

| フィールド | 型 | 説明 |
|---|---|---|
| OrganizationID | int | 組織ID |
| Entries | []int | 有効なグループIDの一覧 |

### GroupHierarchy (集約)

組織ごとのグループ親子関係の集合。AddEdge時に循環チェックを集約内で完結させる。

| フィールド | 型 | 説明 |
|---|---|---|
| OrganizationID | int | 組織ID |
| Edges | []Edge | 親子関係の一覧 |

Edge:

| フィールド | 型 | 説明 |
|---|---|---|
| ParentGroupID | int | 親グループID |
| ChildGroupID | int | 子グループID |

- AddEdge時にBFS/DFSで循環を検出し、循環があればエラーを返す
- ActiveGroupListで上限があるためEdge数も有界

### GroupUsers (集約)

グループに所属するユーザーIDの集合。

| フィールド | 型 | 説明 |
|---|---|---|
| GroupID | int | グループID |
| UserIDs | []int | 所属するユーザーIDの一覧 |

- ユーザーが無効化されてもメンバーシップは残す。参照時にActiveUserListでフィルタする (結果整合性)

### GroupChildGroups (集約)

グループに所属する子グループIDの集合。

| フィールド | 型 | 説明 |
|---|---|---|
| GroupID | int | グループID |
| ChildGroupIDs | []int | 所属する子グループIDの一覧 |

- グループが無効化されてもメンバーシップは残す。参照時にActiveGroupListでフィルタする (結果整合性)
- 追加・削除時はGroupHierarchyも同時に更新する

### SessionToken (集約)

Cookie認証で使用するトークン。個別トークンの検証・延長・破棄を担当する。

| フィールド | 型 | 説明 |
|---|---|---|
| ID | UUID | 識別子 |
| UserID | int | ユーザーID |
| LoginID | string | ログインID |
| OrganizationName | string | 組織名 |
| TokenHash | string | SHA256ハッシュ (生トークンは保存しない) |
| CreatedAt | time.Time | 作成日時 (absolute timeout計算用) |
| ExpiresAt | time.Time | 有効期限 (sliding window) |
| RevokedAt | *time.Time | 破棄日時 (nilなら有効) |

### SessionTokenWhitelist (集約)

ユーザーごとのSessionToken IDの集合。「ユーザーあたり最新N件を保持する」という不変条件を守る。

SessionTokenとは独立した集約であり、SessionTokenWhitelistはトークンのIDとCreatedAtだけを保持する。トークンの追加時にmaxSizeを超えた場合、CreatedAtが古いものから削除対象として返す。

| フィールド | 型 | 説明 |
|---|---|---|
| UserID | int | ユーザーID |
| Entries | []Entry | トークンのID + CreatedAt の一覧 |
| MaxSize | int | ユーザーあたりの最大保持数 |

Entry:

| フィールド | 型 | 説明 |
|---|---|---|
| ID | string | トークンID |
| CreatedAt | time.Time | 作成日時 |

### RefreshToken (集約)

Token認証で使用するリフレッシュトークン。個別トークンの検証・破棄を担当する。

| フィールド | 型 | 説明 |
|---|---|---|
| ID | UUID | 識別子 |
| UserID | int | ユーザーID |
| LoginID | string | ログインID |
| OrganizationName | string | 組織名 |
| TokenHash | string | SHA256ハッシュ |
| CreatedAt | time.Time | 作成日時 |
| ExpiresAt | time.Time | 有効期限 |
| RevokedAt | *time.Time | 破棄日時 |

### RefreshTokenWhitelist (集約)

ユーザーごとのRefreshToken IDの集合。SessionTokenWhitelistと同一パターン。

### AccessToken (集約)

Token認証で使用するアクセストークン (JWT)。個別トークンの検証・破棄を担当する。

| フィールド | 型 | 説明 |
|---|---|---|
| ID | UUID | 識別子 (= JWTのJTI) |
| RefreshTokenID | string | 紐づくRefreshTokenのID |
| UserID | int | ユーザーID |
| LoginID | string | ログインID |
| OrganizationName | string | 組織名 |
| CreatedAt | time.Time | 作成日時 |
| ExpiresAt | time.Time | 有効期限 |
| RevokedAt | *time.Time | 破棄日時 |

AccessTokenはJWTなのでTokenHashは不要。JTI(=ID)で照合する。

### AccessTokenWhitelist (集約)

ユーザーごとのAccessToken IDの集合。SessionTokenWhitelistと同一パターン。

## 環境変数

| 環境変数 | デフォルト | 説明 |
|---|---|---|
| AUTH_SESSION_TOKEN_TTL_MIN | 30 | SessionTokenのsliding window (分) |
| AUTH_SESSION_MAX_TTL_MIN | 1440 | SessionTokenのabsolute timeout (分, 24時間) |
| AUTH_ACCESS_TOKEN_TTL_MIN | 60 | AccessTokenの有効期限 (分, 1時間) |
| AUTH_REFRESH_TOKEN_TTL_MIN | 43200 | RefreshTokenの有効期限 (分, 30日) |
| AUTH_SIGNING_KEY | - | JWT署名鍵 (最小32文字) |
| AUTH_COOKIE_SECURE | true | CookieのSecure属性 |
| AUTH_COOKIE_SAME_SITE | Lax | CookieのSameSite属性 |
| AUTH_COOKIE_PATH | / | CookieのPath属性 |
| AUTH_TOKEN_WHITELIST_SIZE | 10 | ユーザーあたりのトークン保持数 |

## 認証フロー

### Cookie認証: ログイン

1. クライアントがloginID + passwordをPOST
2. サーバーが認証情報を検証
3. opaque tokenを生成し、SHA256ハッシュをDBに保存
4. インメモリキャッシュにも保存
5. `Set-Cookie: session_token=xxx; HttpOnly; Secure; SameSite=Lax; Path=/` を返す

### Cookie認証: リクエスト (ミドルウェア)

1. `session_token` Cookieからトークンを取得
2. SHA256ハッシュを計算し、キャッシュで照合 (miss時はDB)
3. 未破棄・未期限切れ・absolute timeout内であることを確認
4. ExpiresAtを+30分延長 (ただしCreatedAt+24時間が上限)
5. キャッシュ・DBを更新し、新しい有効期限のCookieを返す

### Token認証: ログイン

1. クライアントがloginID + password + `X-Token-Delivery: json` をPOST
2. サーバーが認証情報を検証
3. RefreshToken (opaque) を生成、SHA256ハッシュをDBに保存
4. AccessToken (JWT, JTI=UUID) を生成、DBに保存
5. JTIをインメモリキャッシュに保存
6. ユーザーのトークンが10件超なら古いものを削除
7. `{accessToken, refreshToken}` をJSONで返す

### Token認証: リクエスト (ミドルウェア)

1. `Authorization: Bearer <JWT>` からトークンを取得
2. JWT署名を検証
3. JTIをキャッシュで照合 (miss時はDB)
4. 未破棄・未期限切れであることを確認

### Token認証: リフレッシュ

1. クライアントがrefreshTokenをPOST
2. SHA256ハッシュでDB照合
3. 未破棄・未期限切れであることを確認
4. 新しいAccessToken (JWT) を生成しDBに保存
5. JTIをキャッシュに保存
6. 10件超なら古いAccessTokenを削除
7. `{accessToken}` をJSONで返す

### Revoke

1. クライアントがtokenをPOST
2. JWT形式ならAccessToken、それ以外ならRefreshTokenと判定
3. AccessToken revoke: DBでrevoke、キャッシュから削除
4. RefreshToken revoke: DBでrevoke + 紐づくAccessToken全てrevoke、キャッシュから削除
